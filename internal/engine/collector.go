package engine

import (
	"context"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Collector struct {
	clientset *kubernetes.Clientset
}

func NewCollector(c *kubernetes.Clientset) *Collector {
	return &Collector{clientset: c}
}

// FindRootDeployment walks the OwnerReferences tree to locate the managing deployment name.
func (c *Collector) FindRootDeployment(namespace, podName string) (string, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	owners := pod.GetOwnerReferences()
	if len(owners) == 0 {
		return podName, nil
	}

	if owners[0].Kind == "ReplicaSet" {
		rs, err := c.clientset.AppsV1().ReplicaSets(namespace).Get(context.Background(), owners[0].Name, metav1.GetOptions{})
		if err != nil {
			return owners[0].Name, nil
		}

		rsOwners := rs.GetOwnerReferences()
		if len(rsOwners) > 0 && rsOwners[0].Kind == "Deployment" {
			return rsOwners[0].Name, nil
		}
		return owners[0].Name, nil
	}

	return owners[0].Name, nil
}

// FetchPodLogs tries to read active logs, falls back to the previous crashed container logs, 
// and finally falls back to the system event stream if the container never started.
func (c *Collector) FetchPodLogs(namespace, podName string) (string, error) {
	ctx := context.Background()

	// 1. Fetch Pod spec metadata to isolate the target container name
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil || len(pod.Spec.Containers) == 0 {
		return c.scrapeEventStream(namespace, podName)
	}
	containerName := pod.Spec.Containers[0].Name

	// 2. Try to harvest the CRASHED container runtime logs (Previous: true)
	tailLines := int64(50)
	logOptions := &corev1.PodLogOptions{
		Container:  containerName,
		Previous:   true, // 🔍 Pulls the logs of the container right before it died
		TailLines: &tailLines,
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	stream, err := req.Stream(ctx)
	
	// 3. Fallback: If no previous container logs exist (e.g. ImagePullBackOff), try reading live logs
	if err != nil {
		logOptions.Previous = false
		req = c.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
		stream, err = req.Stream(ctx)
	}

	// 4. Final Fallback: If logs are completely unreachable, scrape the cluster event stream
	if err != nil {
		return c.scrapeEventStream(namespace, podName)
	}
	defer stream.Close()

	logBytes, err := io.ReadAll(stream)
	if err != nil || len(logBytes) == 0 || strings.TrimSpace(string(logBytes)) == "" {
		return c.scrapeEventStream(namespace, podName)
	}

	var output strings.Builder
	output.WriteString("--- CRASHED CONTAINER APPLICATION RUNTIME STACK TRACE ---\n")
	output.WriteString(string(logBytes))
	return output.String(), nil
}

// Private helper to extract cluster events when container engines fail to initialize
func (c *Collector) scrapeEventStream(namespace, podName string) (string, error) {
	events, err := c.clientset.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: "involvedObject.name=" + podName + ",involvedObject.kind=Pod",
	})
	if err != nil {
		return "", err
	}

	if len(events.Items) == 0 {
		return "No diagnostic telemetry logs or cluster events discovered for this target resource.", nil
	}

	var eventReport strings.Builder
	eventReport.WriteString("--- KUBERNETES SYSTEM EVENT STREAM TELEMETRY ---\n")
	for _, item := range events.Items {
		eventReport.WriteString("[" + item.LastTimestamp.Time.Format("15:04:05") + "] Reason: " + item.Reason + " | Message: " + item.Message + "\n")
	}

	return eventReport.String(), nil
}