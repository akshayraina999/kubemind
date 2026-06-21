# 🚀 KubeMind Infrastructure Setup & Run-Book

This guide walks through setting up and testing KubeMind locally using Minikube and Ollama.

---

# 1. Prerequisites

Ensure the following tools are installed and available in your shell `$PATH`:

* **Kubernetes Cluster**

  * Minikube (`v1.32+`) or Docker Desktop Kubernetes

* **Local LLM Runtime**

  * Ollama

* **CLI Utilities**

  * `kubectl`
  * `jq`
  * `bash`

Verify installations:

```bash
kubectl version --client
minikube version
ollama --version
jq --version
```

---

# 2. Start Your Local Kubernetes Cluster

If using Minikube:

```bash
minikube start
```

Verify cluster access:

```bash
kubectl get nodes
```

You should see your local node in the `Ready` state.

---

# 3. Install and Start Ollama

Start the Ollama daemon:

```bash
ollama serve
```

In another terminal, pull the model used by KubeMind:

```bash
ollama pull qwen3:8b
```

Verify that Ollama is working:

```bash
ollama run qwen3:8b "Return only JSON: {\"status\":\"ok\"}"
```

Expected output:

```json
{"status":"ok"}
```

---

# 4. Create the Testing Namespace

Create the isolated namespace used by KubeMind:

```bash
kubectl create namespace kubemind
```

Verify:

```bash
kubectl get ns kubemind
```

---

# 5. Deploy Sample Faulty Applications

KubeMind currently ships with intentionally broken workloads for testing.

Deploy them:

```bash
kubectl apply -f kubernetes/test-apps/broken-app.yaml -n kubemind
kubectl apply -f kubernetes/test-apps/crash-app.yaml -n kubemind
kubectl apply -f kubernetes/test-apps/oom-app.yaml -n kubemind
kubectl apply -f kubernetes/test-apps/pending-app.yaml -n kubemind
kubectl apply -f kubernetes/test-apps/failed-job.yaml -n kubemind
```

Check workload status:

```bash
kubectl get pods -n kubemind
```

---

# 6. Initialize Local Workspace

Create required directories:

```bash
mkdir -p \
logs/context \
logs/remediation \
logs/remediation-plans \
reports \
state/processed
```

(Optional) preserve empty directories in Git:

```bash
touch logs/context/.gitkeep
touch logs/remediation/.gitkeep
touch logs/remediation-plans/.gitkeep
touch reports/.gitkeep
touch state/processed/.gitkeep
```

---

# 7. Run the KubeMind Pipeline

KubeMind currently operates as a pipeline of independent stages.

## Step 1: Collect Kubernetes Context

```bash
./engine/collector.sh <pod-name>
```

Example:

```bash
./engine/collector.sh broken-app-5b8f6f4976-qqft7
```

This generates:

```text
logs/context/<pod-name>.txt
```

---

## Step 2: Generate AI Remediation Plan

```bash
./ai/analyze.sh <pod-name>
```

Example:

```bash
./ai/analyze.sh broken-app-5b8f6f4976-qqft7
```

This invokes the local Ollama model and generates:

```text
logs/remediation-plans/<pod-name>.json
```

Example output:

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

## Step 3: Generate Executable Commands

```bash
./engine/remediator.sh <pod-name>
```

Example:

```bash
./engine/remediator.sh broken-app-5b8f6f4976-qqft7
```

Generated commands are written to:

```text
logs/remediation/<pod-name>.txt
```

---

# 8. Current Workflow

```text
Kubernetes Pod
       │
       ▼
collector.sh
       │
       ▼
logs/context/*.txt
       │
       ▼
analyze.sh
(Local Ollama LLM)
       │
       ▼
logs/remediation-plans/*.json
       │
       ▼
remediator.sh
       │
       ▼
Executable kubectl commands
```

---

# 9. Current Project Status

✅ Kubernetes context collection

✅ Failure classification

✅ Local LLM-based root cause analysis

✅ Structured remediation plan generation

✅ Command generation

🚧 Safe command execution engine

🚧 Verification engine

🚧 Continuous cluster watcher

🚧 Autonomous remediation approval workflow

🚧 Multi-resource support (Nodes, PVCs, PDBs, etc.)

---

# 10. Clean Previous Runs

Remove generated artifacts:

```bash
rm -rf logs/context/*
rm -rf logs/remediation/*
rm -rf logs/remediation-plans/*
```

Reset state tracking:

```bash
rm -rf state/processed/*
```
