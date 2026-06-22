package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Reporter handles structuring audit data and writing filesystem reports.
type Reporter struct {
	TargetDir string
}

// NewReporter initializes the reporting filesystem structure.
func NewReporter(targetDir string) *Reporter {
	// Ensure the reports output folder exists
	_ = os.MkdirAll(targetDir, 0755)
	return &Reporter{TargetDir: targetDir}
}

// WritePostMortem compiles a clean Markdown post-mortem document.
func (r *Reporter) WritePostMortem(namespace, podName, deploymentName, anomalyReason, rawPlan string, telemetryBytes int) (string, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.md", timestamp, deploymentName)
	fullPath := filepath.Join(r.TargetDir, filename)

	markdownContent := fmt.Sprintf(`# 🧠 KubeMind Auto-Remediation Post-Mortem
**Timestamp:** %s  
**Target Workspace Namespace:** %s  
**Root Incident Controller:** Deployment/%s  
**Triggering Ephemeral Identity:** Pod/%s  

---

## 🔍 Telemetry Diagnostics Triage
* **Intercepted Anomaly Reason:** %s
* **Harvested Context Weight:** %d bytes of live stream diagnostic events.

---

## 🔮 Executed Remediation Blueprint
The automated advisory model compiled the following operational action matrix:

%s

---
## 🛡️ Audit Validation Status
* **Status:** Strategic Merge Patch Dispatched Natively via Client-SDK.
* **Execution Gate:** Guardrail-validated; matching schema requirements.
`, time.Now().Format("2006-01-02 15:04:05 Mon"), namespace, deploymentName, podName, anomalyReason, telemetryBytes, "```json\n"+rawPlan+"\n```")

	// Write the file to your reports directory
	err := os.WriteFile(fullPath, []byte(markdownContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to commit report markdown file: %w", err)
	}

	return filename, nil
}
