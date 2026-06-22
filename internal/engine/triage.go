package engine

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

type TriageRouter struct {
	clientset *kubernetes.Clientset
}

func NewTriageRouter(c *kubernetes.Clientset) *TriageRouter {
	return &TriageRouter{clientset: c}
}

// FetchManifestSnippet dynamically routes based on resource kind and builds a high-signal YAML summary snippet.
func (t *TriageRouter) FetchManifestSnippet(namespace, resourceKind, resourceName string) (string, error) {
	ctx := context.Background()

	switch resourceKind {

	case "Deployment":
		dep, err := t.clientset.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		// Isolate container configuration block
		yamlBytes, err := yaml.Marshal(dep.Spec.Template.Spec.Containers)
		return string(yamlBytes), err

	case "PersistentVolumeClaim", "PVC":
		pvc, err := t.clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		
		// Create a clean, lightweight summary map to keep the Slack alert dense and clear
		pvcSummary := map[string]interface{}{
			"metadata": map[string]string{
				"name": pvc.Name,
			},
			"spec": map[string]interface{}{
				"accessModes":      pvc.Spec.AccessModes,
				"storageClassName": pvc.Spec.StorageClassName,
				"resources":        pvc.Spec.Resources,
				"volumeName":       pvc.Spec.VolumeName,
			},
			"status": map[string]interface{}{
				"phase": pvc.Status.Phase,
			},
		}
		
		yamlBytes, err := yaml.Marshal(pvcSummary)
		return string(yamlBytes), err

	case "Ingress":
		ing, err := t.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		// Extract routing rules, backends, and TLS parameters
		ingSummary := map[string]interface{}{
			"metadata": map[string]string{
				"name": ing.Name,
			},
			"spec": map[string]interface{}{
				"rules": ing.Spec.Rules,
				"tls":   ing.Spec.TLS,
			},
		}

		yamlBytes, err := yaml.Marshal(ingSummary)
		return string(yamlBytes), err

	default:
		return fmt.Sprintf("# Dynamic context mapping fallback for resource kind: %s/%s", resourceKind, resourceName), nil
	}
}