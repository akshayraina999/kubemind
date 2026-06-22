package watcher

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Anomaly holds information about any intercepted cluster-wide resource failure state.
type Anomaly struct {
	ResourceKind string
	ResourceName string
	Reason       string
	Message      string
}

// Scanner encapsulates our connection to the active cluster space.
type Scanner struct {
	Clientset *kubernetes.Clientset
	Namespace string
}

// NewScanner builds a type-safe client configuration map connecting to the API server.
func NewScanner(namespace string) (*Scanner, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig configurations: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize clientset interfaces: %w", err)
	}
	return &Scanner{Clientset: clientset, Namespace: namespace}, nil
}

// StartDaemon spins up an infinite monitoring channel ticker until the context is canceled.
func (s *Scanner) StartDaemon(ctx context.Context, interval time.Duration, handleAnomaly func(Anomaly)) {
	fmt.Printf("🚀 KubeMind Hybrid Daemon Active: Monitoring state for [%s] every %v...\n", s.Namespace, interval)
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Watcher Daemon received shutdown signal. Draining monitoring loops...")
			return
		case <-ticker.C:
			anomalies, err := s.CollectAnomalies()
			if err != nil {
				fmt.Printf("⚠️ Scanning Error encountered: %v\n", err)
				continue
			}

			for _, anomaly := range anomalies {
				handleAnomaly(anomaly)
			}
		}
	}
}

// CollectAnomalies inspects direct resource specifications to bypass ephemeral event drops.
func (s *Scanner) CollectAnomalies() ([]Anomaly, error) {
	var foundAnomalies []Anomaly
	ctx := context.TODO()

	// 1. DIRECT STATE INSPECTION: PersistentVolumeClaims (PVCs)
	pvcs, err := s.Clientset.CoreV1().PersistentVolumeClaims(s.Namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, pvc := range pvcs.Items {
			// If a PVC is stuck in Pending, it is an active anomaly regardless of age
			if pvc.Status.Phase == "Pending" {
				// 🔍 FIXED: Safely dereference the pointer with a fallback string to squash pointer address bugs
				scName := "default"
				if pvc.Spec.StorageClassName != nil {
					scName = *pvc.Spec.StorageClassName
				}

				foundAnomalies = append(foundAnomalies, Anomaly{
					ResourceKind: "PersistentVolumeClaim",
					ResourceName: pvc.Name,
					Reason:       "FailedBinding",
					Message:      fmt.Sprintf("PersistentVolumeClaim is stuck in Pending state indefinitely. Target StorageClass requested: %s", scName),
				})
			}
		}
	}

	// 2. DIRECT STATE INSPECTION: Pod container status checks
	pods, err := s.Clientset.CoreV1().Pods(s.Namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, pod := range pods.Items {
			for _, status := range pod.Status.ContainerStatuses {
				if status.State.Waiting != nil {
					reason := status.State.Waiting.Reason
					// Target explicit, long-running container initialization blocks
					if reason == "CrashLoopBackOff" || reason == "ImagePullBackOff" || reason == "ErrImagePull" || reason == "CreateContainerConfigError" {
						foundAnomalies = append(foundAnomalies, Anomaly{
							ResourceKind: "Pod",
							ResourceName: pod.Name,
							Reason:       reason,
							Message:      fmt.Sprintf("Container %s is down under state: %s. Message details can be found inside log streams.", status.Name, reason),
						})
					}
				}
			}
		}
	}

	// 3. DIRECT STATE INSPECTION: Ingress Endpoint Targets Check
	ingresses, err := s.Clientset.NetworkingV1().Ingresses(s.Namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, ing := range ingresses.Items {
			// If an Ingress lacks an active controller routing address allocation block, trap it
			if len(ing.Status.LoadBalancer.Ingress) == 0 {
				foundAnomalies = append(foundAnomalies, Anomaly{
					ResourceKind: "Ingress",
					ResourceName: ing.Name,
					Reason:       "NoAddressAssigned",
					Message:      "Ingress routing matrix is active but has not been assigned a public load balancer gateway IP address block.",
				})
			}
		}
	}

	return foundAnomalies, nil
}