package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/akshayraina999/kubemind/internal/apis/v1alpha1"
	"github.com/akshayraina999/kubemind/internal/ai"
	"github.com/akshayraina999/kubemind/internal/engine"
	"github.com/akshayraina999/kubemind/internal/watcher"
)

// TriageWatcherReconciler reconciles a TriageWatcher object
type TriageWatcherReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	KubeClientset *kubernetes.Clientset
	LastAlerted   map[string]time.Time
	TrackMutex    sync.RWMutex
}

// sendOperatorSlackAlert ships immediate remediation details down to the designated cluster channel
func sendOperatorSlackAlert(webhookURL, namespace, resourceName, targetKind, reason, patchSuggestion, manifestSnippet string) {
	if webhookURL == "" {
		return
	}
	var parsedAI struct {
		Explanation string `json:"human_explanation"`
	}
	_ = json.Unmarshal([]byte(patchSuggestion), &parsedAI)

	remediationSteps := parsedAI.Explanation
	if remediationSteps == "" {
		remediationSteps = "Automated prescription analysis failed to generate a clear summary. Please inspect the resource manifests manually."
	}

	slackMessage := fmt.Sprintf(
		"🚨 *KubeMind Operator Incident Alert*\n"+
			"• *Namespace:* `%s`\n"+
			"• *Resource Kind:* `%s`\n"+
			"• *Target Resource:* `%s`\n"+
			"• *Intercepted Condition:* *%s*\n\n"+
			"🛠️ *Recommended Actionable Fix:* \n>%s\n\n"+
			"📄 *Failing Manifest Code Snippet:* \n```yaml\n%s\n```",
		namespace, targetKind, resourceName, reason, remediationSteps, manifestSnippet,
	)

	payload := map[string]string{"text": slackMessage}
	jsonBytes, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonBytes))
	if err == nil {
		defer resp.Body.Close()
	}
}

// +kubebuilder:rbac:groups=v1alpha1.kubemind.io,resources=triagewatchers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=v1alpha1.kubemind.io,resources=triagewatchers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets;pods;persistentvolumeclaims;events,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch

func (r *TriageWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Fetch the user's declarative TriageWatcher resource
	var watcherInstance v1alpha1.TriageWatcher
	if err := r.Get(ctx, req.NamespacedName, &watcherInstance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	targetNS := watcherInstance.Spec.TargetNamespace
	logger.Info("Starting active reconciliation sweep loop", "targetNamespace", targetNS)

	// 2. Resolve the Slack Webhook token via the Secret Reference safely
	var secret corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: watcherInstance.Spec.SlackSecretRef}, &secret); err != nil {
		logger.Error(err, "Failed to locate referenced Slack Secret string", "secretName", watcherInstance.Spec.SlackSecretRef)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	webhookBytes, exists := secret.Data["webhook-url"]
	if !exists {
		logger.Error(nil, "Secret data object missing mandatory 'webhook-url' key string mapping")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	slackURL := string(webhookBytes)

	// 3. Extract the System Prompt from disk or environments
	promptPath := filepath.Join("ai", "prompts", "remediation_prompt.txt")
	systemPromptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		systemPromptBytes = []byte("Analyze the following Kubernetes error and provide remediation steps.")
	}

	// 4. Initialize our sub-engines bound to this specific target namespace configuration
	collectorEngine := engine.NewCollector(r.KubeClientset)
	triageRouter := engine.NewTriageRouter(r.KubeClientset)

	// Pull environmental variables for local Ollama configurations
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	modelName := os.Getenv("MODEL_NAME")
	if modelName == "" {
		modelName = "llama3"
	}
	aiClient := ai.NewClient(ollamaURL, modelName)

	localScanner := &watcher.Scanner{
		Clientset: r.KubeClientset,
		Namespace: targetNS,
	}

	// 5. Gather anomalies using our dual-layer status verification method
	anomalies, err := localScanner.CollectAnomalies()
	if err != nil {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	// 6. Process any detected failures sequentially
	for _, anomaly := range anomalies {
		targetKind := anomaly.ResourceKind
		targetResourceName := anomaly.ResourceName
		rawPodName := anomaly.ResourceName

		if targetKind == "Pod" {
			resolvedName, err := collectorEngine.FindRootDeployment(targetNS, anomaly.ResourceName)
			if err == nil {
				targetResourceName = resolvedName
				targetKind = "Deployment"
			}
		}

		snoozeKey := fmt.Sprintf("%s-%s-%s", targetKind, targetResourceName, anomaly.Reason)

		r.TrackMutex.Lock()
		lastAlertTime, alertedExists := r.LastAlerted[snoozeKey]
		if alertedExists && time.Since(lastAlertTime) < 30*time.Minute {
			r.TrackMutex.Unlock()
			continue
		}
		r.LastAlerted[snoozeKey] = time.Now()
		r.TrackMutex.Unlock()

		// Run async triage pipeline logic
		go func(kind, name, podStr, reason, msg string) {
			log := log.FromContext(context.Background()).WithValues("pod", podStr, "reason", reason)
			log.Info("🎯 Intercepted anomaly. Initiating AI diagnostic pipeline...")

			manifestSnippet, err := triageRouter.FetchManifestSnippet(targetNS, kind, name)
			if err != nil {
				manifestSnippet = fmt.Sprintf("# Warning: Could not harvest spec for %s: %v", kind, err)
			}

			// Fetch container logs only if it's not a pulling or configuration block
			var telemetryData string
			if reason == "ImagePullBackOff" || reason == "ErrImagePull" || reason == "CreateContainerConfigError" {
				log.Info("⚠️ Container hasn't started yet. Utilizing event description messages for context...")
				telemetryData = fmt.Sprintf("Kubernetes Event Message: %s", msg)
			} else {
				telemetryData, err = collectorEngine.FetchPodLogs(targetNS, podStr)
				if err != nil {
					log.Error(err, "Failed to pull container log streams; falling back to description.")
					telemetryData = fmt.Sprintf("Fallback context text details: %s", msg)
				}
			}

			compiledPrompt := fmt.Sprintf("%s\n\n[CONTEXT]:\nTarget Object: %s/%s\nTelemetry:\n%s\nError Message:\n%s\nActive Manifest YAML Snippet:\n%s",
				string(systemPromptBytes), kind, name, telemetryData, msg, manifestSnippet)

			log.Info("🧠 Forwarding diagnostic payload to local Ollama instance...")
			rawJSONPlan, err := aiClient.Generate(compiledPrompt, true)
			if err != nil {
				log.Error(err, "❌ Ollama inference engine failed to compile remediation layout.")
				r.TrackMutex.Lock()
				delete(r.LastAlerted, snoozeKey)
				r.TrackMutex.Unlock()
				return
			}

			log.Info("📤 Remediation plan synthesized successfully. Shipping alert to Slack...")
			sendOperatorSlackAlert(slackURL, targetNS, name, kind, reason, rawJSONPlan, manifestSnippet)
		}(targetKind, targetResourceName, rawPodName, anomaly.Reason, anomaly.Message)
	}

	// Requeue the event to inspect changes continuously every 10 seconds
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *TriageWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.LastAlerted == nil {
		r.LastAlerted = make(map[string]time.Time)
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TriageWatcher{}).
		Complete(r)
}