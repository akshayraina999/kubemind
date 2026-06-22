package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/akshayraina999/kubemind/internal/ai"
	"github.com/akshayraina999/kubemind/internal/config"
	"github.com/akshayraina999/kubemind/internal/engine"
	"github.com/akshayraina999/kubemind/internal/watcher"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func sendSlackAlert(webhookURL, namespace, resourceName, targetKind, reason, patchSuggestion, manifestSnippet string) {
	if webhookURL == "" || webhookURL == "YOUR_FALLBACK_PLACEHOLDER" {
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
		"🚨 *KubeMind Incident Triage Alert*\n"+
			"• *Namespace:* `%s`\n"+
			"• *Resource Kind:* `%s`\n"+
			"• *Target Resource:* `%s`\n"+
			"• *Intercepted Event Reason:* *%s*\n\n"+
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

func main() {
	fmt.Println("🧠 KubeMind Engine: Initializing Go Runtime...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("🚨 Critical Initialization Failure: %v", err)
	}
	fmt.Printf("✅ Config loaded successfully. Target Namespace: %s\n", cfg.TargetNS)

	promptPath := filepath.Join("ai", "prompts", "remediation_prompt.txt")
	systemPromptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		log.Fatalf("🚨 Failed to read system prompt file: %v", err)
	}
	systemPrompt := string(systemPromptBytes)

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("🚨 Kubeconfig Error: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatalf("🚨 Clientset Initialization Error: %v", err)
	}
	fmt.Println("🚀 Authenticated cleanly with Kubernetes API Server.")

	collectorEngine := engine.NewCollector(clientset)
	triageRouter := engine.NewTriageRouter(clientset)
	aiClient := ai.NewClient(cfg.OllamaURL, cfg.ModelName)
	clusterScanner, err := watcher.NewScanner(cfg.TargetNS)
	if err != nil {
		log.Fatalf("🚨 Scanner Engine Initialization Failure: %v", err)
	}

	var trackMutex sync.RWMutex
	lastAlertedMap := make(map[string]time.Time)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sem := make(chan struct{}, 2)

	anomalyHandler := func(a watcher.Anomaly) {
		go func(anomaly watcher.Anomaly) {
			targetKind := anomaly.ResourceKind
			targetResourceName := anomaly.ResourceName
			rawPodName := anomaly.ResourceName // 🔍 Preserve the actual Pod name instance string

			// Standardize pod tracking architecture down to its deployment owner references
			if targetKind == "Pod" {
				resolvedName, err := collectorEngine.FindRootDeployment(cfg.TargetNS, anomaly.ResourceName)
				if err == nil {
					targetResourceName = resolvedName
					targetKind = "Deployment"
				}
			}

			snoozeKey := fmt.Sprintf("%s-%s-%s", targetKind, targetResourceName, anomaly.Reason)

			// ─── 🔒 ATOMIC THREAD-SAFE COOLDOWN GATE ───────────────────────
			trackMutex.Lock()
			lastAlertTime, exists := lastAlertedMap[snoozeKey]

			if exists && time.Since(lastAlertTime) < 30*time.Minute {
				trackMutex.Unlock()
				return
			}

			lastAlertedMap[snoozeKey] = time.Now()
			trackMutex.Unlock()
			// ───────────────────────────────────────────────────────────────

			fmt.Printf("\n🚨 Multi-Directional Incident Triggered: %s [%s] dropped state via reason [%s]!\n", 
				targetKind, targetResourceName, anomaly.Reason)

			sem <- struct{}{}        
			defer func() { <-sem }() 

			manifestSnippet, err := triageRouter.FetchManifestSnippet(cfg.TargetNS, targetKind, targetResourceName)
			if err != nil {
				manifestSnippet = fmt.Sprintf("# Warning: Could not harvest manifest snippet for %s: %v", targetKind, err)
			}

			// 🔍 FIXED: Passing rawPodName so the log scavenger tracks the real failing pod instead of the Deployment name
			telemetryData, err := collectorEngine.FetchPodLogs(cfg.TargetNS, rawPodName)
			if err != nil {
				telemetryData = fmt.Sprintf("Event message detail logs context: %s", anomaly.Message)
			}

			compiledPrompt := fmt.Sprintf("%s\n\n[CONTEXT]:\nTarget Object: %s/%s\nTelemetry Matrix Context:\n%s\nSystem Event Error Message:\n%s", 
				systemPrompt, targetKind, targetResourceName, telemetryData, anomaly.Message)
			
			fmt.Printf("🤖 Querying Ollama for %s/%s triage path...\n", targetKind, targetResourceName)
			rawJSONPlan, err := aiClient.Generate(compiledPrompt, true)
			if err != nil {
				trackMutex.Lock()
				delete(lastAlertedMap, snoozeKey)
				trackMutex.Unlock()
				fmt.Printf("❌ AI Anomaly process failed: %v\n", err)
				return
			}

			fmt.Printf("📢 Shipping multi-resource triage card for [%s] to Slack...\n", targetResourceName)
			sendSlackAlert(cfg.SlackURL, cfg.TargetNS, targetResourceName, targetKind, anomaly.Reason, rawJSONPlan, manifestSnippet)
		}(a)
	}

	clusterScanner.StartDaemon(ctx, 10*time.Second, anomalyHandler)
	fmt.Println("👋 KubeMind execution loop cleanly terminated. Offline.")
}