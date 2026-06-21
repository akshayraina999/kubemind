# 🧠 KubeMind

KubeMind is an AI-assisted Kubernetes troubleshooting engine that collects cluster context, analyzes failing workloads using a local LLM, and generates structured remediation plans.

The project is designed as the foundation for a future self-healing Kubernetes platform engineering agent.

---

# 🚀 What KubeMind Does Today

KubeMind currently performs four major tasks:

1. **Collects detailed Kubernetes context**
2. **Identifies failing workloads**
3. **Sends workload context to a local LLM (Ollama)**
4. **Generates remediation recommendations in structured JSON**

The system runs entirely locally and does not require external AI APIs.

---

# 🏗 Current Architecture

```text
Kubernetes Cluster
        |
        v
+-----------------------+
| collector.sh          |
| Collect pod context   |
+-----------------------+
        |
        v
logs/context/*.txt
        |
        v
+-----------------------+
| analyze.sh            |
| Local LLM Analysis    |
| (Ollama + Qwen)       |
+-----------------------+
        |
        v
logs/remediation-plans/*.json
        |
        v
+-----------------------+
| remediator.sh         |
| Generate kubectl      |
| remediation commands  |
+-----------------------+
        |
        v
logs/remediation/*.txt
```

---

# 📂 Project Structure

```text
kubemind/
│
├── ai/
│   ├── analyze.sh
│   └── prompts/
│       └── remediation_prompt.txt
│
├── cmd/
│   └── kubemind.sh
│
├── engine/
│   ├── classifier.sh
│   ├── collector.sh
│   └── remediator.sh
│
├── kubernetes/
│   └── test-apps/
│
├── logs/
│   ├── context/
│   ├── remediation/
│   └── remediation-plans/
│
└── reports/
```

---

# ⚙️ How It Works

## Step 1 — Context Collection

```bash
./engine/collector.sh <pod-name>
```

The collector gathers:

* Pod YAML
* Pod description
* Events
* Current logs
* Previous logs
* Container status
* Scheduling information
* Owner workload details

Output:

```text
logs/context/<pod>.txt
```

Example:

```bash
./engine/collector.sh crash-app-988bbf696-4rwtx
```

---

## Step 2 — AI Analysis

```bash
./ai/analyze.sh <pod-name>
```

The analyzer:

* Reads collected context
* Sends it to Ollama
* Uses a local LLM (Qwen)
* Produces a remediation plan

Output:

```text
logs/remediation-plans/<pod>.json
```

Example output:

```json
{
  "resource_type": "Pod",
  "resource_name": "broken-app",
  "issue_type": "ImagePullBackOff",
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

## Step 3 — Remediation Generation

```bash
./engine/remediator.sh <pod-name>
```

The remediator:

* Reads AI-generated JSON
* Extracts remediation actions
* Produces executable kubectl commands

Output:

```text
logs/remediation/<pod>.txt
```

---

# 🧪 Supported Failure Scenarios

Current test applications include:

| Scenario         | Status |
| ---------------- | ------ |
| ImagePullBackOff | ✅      |
| CrashLoopBackOff | ✅      |
| OOMKilled        | ✅      |
| Pending Pods     | ✅      |
| Failed Jobs      | 🚧     |

---

# 🖥 Local Development Setup

## Prerequisites

Install:

* Kubernetes cluster (Minikube or Docker Desktop)
* kubectl
* jq
* Ollama

---

## Install Local Model

Start Ollama:

```bash
ollama serve
```

Pull model:

```bash
ollama pull qwen3:8b
```

Verify:

```bash
ollama run qwen3:8b
```

---

## Create Namespace

```bash
kubectl create namespace kubemind
```

---

## Deploy Test Workloads

```bash
kubectl apply -f kubernetes/test-apps/
```

---

# ▶ Example End-to-End Flow

Collect context:

```bash
./engine/collector.sh broken-app-xxxx
```

Analyze with AI:

```bash
./ai/analyze.sh broken-app-xxxx
```

View generated remediation plan:

```bash
cat logs/remediation-plans/broken-app-xxxx.json
```

Generate commands:

```bash
./engine/remediator.sh broken-app-xxxx
```

View commands:

```bash
cat logs/remediation/broken-app-xxxx.txt
```

---

# 🔮 Planned Features

The long-term vision for KubeMind includes:

* Continuous cluster scanning
* Autonomous remediation execution
* Risk-based approval workflows
* JSON schema enforcement
* Post-remediation verification
* Incident report generation
* Support for cluster-wide failures
* Node health analysis
* PDB analysis
* Eviction analysis
* Storage issue remediation
* Network issue remediation
* Multi-resource dependency graph analysis
* GitOps integration
* ArgoCD integration
* Human approval gates

---

# 🎯 Vision

KubeMind aims to evolve into a self-healing Kubernetes platform engineering agent capable of:

1. Detecting failures automatically.
2. Understanding root causes using AI.
3. Generating safe remediation actions.
4. Executing low-risk fixes autonomously.
5. Verifying cluster health after remediation.

The project follows an **AI-assisted but safety-first** design philosophy where AI recommendations are always validated before execution.
