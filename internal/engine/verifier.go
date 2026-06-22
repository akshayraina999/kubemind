package engine

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Verifier monitors workloads following mutations to guarantee convergence.
type Verifier struct {
	Clientset *kubernetes.Clientset
}

// NewVerifier initializes a rollout validation layer.
func NewVerifier(clientset *kubernetes.Clientset) *Verifier {
	return &Verifier{Clientset: clientset}
}

// VerifyDeploymentHealth blocks and samples the Deployment state over a 15-second window.
func (v *Verifier) VerifyDeploymentHealth(namespace, deploymentName string) (bool, error) {
	fmt.Printf("⏳ Verification Phase: Watching Deployment [%s] for state convergence...\n", deploymentName)

	// Sample the status every 5 seconds, up to 3 times (15-second timeout boundary)
	for i := 1; i <= 3; i++ {
		time.Sleep(5 * time.Second)

		deploy, err := v.Clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to verify deployment status: %w", err)
		}

		// Check if all desired replicas are updated, available, and running cleanly
		if deploy.Status.UpdatedReplicas == deploy.Status.Replicas &&
			deploy.Status.AvailableReplicas == deploy.Status.Replicas &&
			deploy.Status.UnavailableReplicas == 0 {
			fmt.Printf("💚 Verification Success! Workload [%s] has converged to 1/1 Ready.\n", deploymentName)
			return true, nil
		}

		fmt.Printf("... Still waiting for convergence (Sample %d/3). Replicas: %d Available, %d Unavailable\n", 
			i, deploy.Status.AvailableReplicas, deploy.Status.UnavailableReplicas)
	}

	return false, nil
}
