# 🏗️ KubeMind Core Architecture Blueprint

KubeMind follows a modular pipeline architecture for diagnosing Kubernetes failures using a local Large Language Model (LLM).

The system intentionally separates data collection, AI analysis, and remediation generation into independent stages. This makes the platform easier to debug, extend, and secure.

Current execution flow:

```text
Kubernetes Resource
        │
        ▼
collector.sh
(Collect cluster context)
        │
        ▼
Context File (.txt)
        │
        ▼
analyze.sh
(Local LLM Analysis)
        │
        ▼
Structured JSON Plan
        │
        ▼
remediator.sh
(Command Generation)
        │
        ▼
Human Review / Future Auto-Execution
```

---

# 1. Core Components

## Data Collection (`engine/collector.sh`)

The collector is responsible for gathering all relevant Kubernetes troubleshooting information for a resource.

For a given Pod, it collects:

* Pod metadata
* Workload ownership information
* Scheduling information
* Pod description
* Events
* Current logs
* Previous logs
* Container status
* Raw Pod YAML

The collected information is stored as a plain-text context file:

```text
logs/context/<resource>.txt
```

Example:

```bash
./engine/collector.sh crash-app-988bbf696-4rwtx
```

Output:

```text
logs/context/crash-app-988bbf696-4rwtx.txt
```

This context file becomes the knowledge source for the AI layer.

---

## AI Root Cause Analysis (`ai/analyze.sh`)

The analysis engine uses a locally running Ollama model (`qwen3:8b`) to perform:

* Root cause analysis
* Failure classification
* Remediation generation

The workflow is:

1. Read collected context.
2. Append system prompt instructions.
3. Send prompt to Ollama.
4. Extract JSON response.
5. Validate JSON using `jq`.
6. Persist remediation plan.

Example:

```bash
./ai/analyze.sh broken-app-5b8f6f4976-qqft7
```

Generated output:

```text
logs/remediation-plans/broken-app-5b8f6f4976-qqft7.json
```

Example JSON:

```json
{
  "resource_type": "Pod",
  "resource_name": "broken-app-5b8f6f4976-qqft7",
  "issue_type": "ImagePullFailure",
  "root_cause": "Image nginx:notfound does not exist",
  "confidence": 100,
  "remediation": [
    {
      "action_type": "update_image",
      "target_kind": "Deployment",
      "target_name": "broken-app",
      "command": "kubectl set image deployment/broken-app nginx=nginx:latest -n kubemind",
      "risk": "low",
      "auto_executable": true
    }
  ],
  "verification": [
    "kubectl get pods -n kubemind"
  ]
}
```

---

## JSON Validation Layer

LLM output is treated as untrusted.

Before a remediation plan is accepted:

* Chat output is stripped.
* ANSI escape characters are removed.
* JSON is extracted from the response.
* `jq` validates structural correctness.

If validation fails:

```text
ERROR: Invalid JSON returned by model
```

The raw model response is stored for debugging:

```text
logs/remediation-plans/<resource>.json.raw
```

This validation layer prevents malformed responses from entering the remediation stage.

---

## Remediation Engine (`engine/remediator.sh`)

The remediation engine consumes the structured JSON plan and generates executable Kubernetes commands.

Current responsibilities:

* Load remediation JSON.
* Display remediation details.
* Extract generated commands.
* Persist commands for review.

Example:

```bash
./engine/remediator.sh broken-app-5b8f6f4976-qqft7
```

Output:

```text
logs/remediation/broken-app-5b8f6f4976-qqft7.txt
```

Current behavior:

```text
REMEDIATION PLAN
GENERATED COMMANDS
```

Commands are currently generated for human review.

Automatic execution has not yet been implemented.

---

# 2. Current Safety Model

KubeMind currently follows a human-in-the-loop remediation approach.

```text
Collect Context
       ↓
AI Generates Plan
       ↓
Validate JSON
       ↓
Generate Commands
       ↓
Human Reviews Output
```

No commands are automatically executed on the cluster.

This guarantees that malformed AI responses cannot directly mutate cluster resources.

---

# 3. Future Architecture Enhancements

The following capabilities are planned:

* Continuous cluster watcher
* Automatic anomaly detection
* Safe command execution engine
* Post-remediation verification
* Risk-based execution policies
* Human approval workflows
* Support for Nodes, PVCs, PDBs, Services, and cluster-wide failures
* Autonomous remediation loops

Future workflow:

```text
Cluster Watcher
       ↓
Collector
       ↓
AI Analysis
       ↓
Validation
       ↓
Risk Assessment
       ↓
Human Approval / Auto Execute
       ↓
Verification
       ↓
Incident Report
```

---

# 4. Current Directory Ownership

```text
ai/
├── analyze.sh
└── prompts/

engine/
├── collector.sh
├── remediator.sh
├── classifier.sh

logs/
├── context/
├── remediation/
└── remediation-plans/

reports/
state/
```

Each stage writes artifacts to disk, making the complete troubleshooting workflow auditable and easy to debug.
