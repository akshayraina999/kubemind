package engine

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// ActionPlan represents the top-level JSON schema returned by the LLM.
type ActionPlan struct {
	ActionType     string                 `json:"action_type"`
	TargetKind     string                 `json:"target_kind"`
	TargetName     string                 `json:"target_name"`
	Risk           string                 `json:"risk"`
	AutoExecutable bool                   `json:"auto_executable"`
	PatchPayload   map[string]interface{} `json:"patch_payload,omitempty"`
	FallbackCmd    string                 `json:"fallback_command,omitempty"`
}

// Remediator executes type-safe patches against cluster controllers.
type Remediator struct {
	Clientset *kubernetes.Clientset
}

// NewRemediator builds a remediator instance.
func NewRemediator(clientset *kubernetes.Clientset) *Remediator {
	return &Remediator{Clientset: clientset}
}

// ProcessAndPatch decodes the LLM's response and enforces production safety guardrails.
func (r *Remediator) ProcessAndPatch(namespace string, rawLLMResponse string) error {
	var plan ActionPlan

	// 1. Unmarshal the raw token directly into our strongly-typed Go structure
	err := json.Unmarshal([]byte(rawLLMResponse), &plan)
	if err != nil {
		return fmt.Errorf("AI payload failed schema validation checks: %w", err)
	}

	fmt.Printf("\n🛡️ Remediator Triage: Evaluating Action Plan [%s] for %s/%s...\n", plan.ActionType, plan.TargetKind, plan.TargetName)
	fmt.Printf("⚠️ Risk Assessment Level: [%s] | Auto-Execute Allowed: %t\n", plan.Risk, plan.AutoExecutable)

	// 2. Production Safety Fence / Guardrail Checks
	if !plan.AutoExecutable || plan.Risk == "HIGH" {
		return fmt.Errorf("🚨 EXECUTION FENCE TRIGGERED: High risk or non-autonomous plan halted for operator sign-off")
	}

	if plan.TargetKind != "Deployment" {
		return fmt.Errorf("❌ Security Boundary Violation: Mutating target kind [%s] is blocked on this runtime profile", plan.TargetKind)
	}

	// 3. Serialize the dynamic map back to a clean JSON string for the patch payload
	patchBytes, err := json.Marshal(plan.PatchPayload)
	if err != nil {
		return fmt.Errorf("failed to compile patch payload: %w", err)
	}

	fmt.Printf("🚀 Applying safe Strategic Merge Patch to Deployment [%s]...\n", plan.TargetName)

	// 4. Execute Native Strategic Merge Patch against the live API Server
	_, err = r.Clientset.AppsV1().Deployments(namespace).Patch(
		context.TODO(),
		plan.TargetName,
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("kubernetes API rejected the patch execution: %w", err)
	}

	fmt.Println("✅ Workload successfully patched! Monitoring cluster for state convergence...")
	return nil
}