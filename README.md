# 🚀 Getting Started with KubeMind Operator

Welcome to the **`refactor/kubemind-operator`** branch.

This branch introduces **KubeMind Operator**, a Kubernetes-native Day-2 Operations platform built using the standard Go project layout and deployed as a Kubernetes Operator.

The operator continuously watches cluster resources, detects anomalies, gathers diagnostics, and leverages a locally hosted Large Language Model (LLM) running through Ollama to generate AI-assisted root cause analysis and remediation suggestions.

---

# 🏗️ Architecture Overview

KubeMind follows a hybrid architecture where the operator runs inside Kubernetes while AI inference remains local on your workstation.

```text
Kubernetes Cluster Anomalies
(Pods, PVCs, Jobs, Nodes, etc.)
                │
                ▼
       KubeMind Operator
  (Watches cluster events & states)
                │
                ▼
    Builds Diagnostic Context
                │
                ▼
Routes Context over Internal DNS Bridge
                │
                ▼
ollama-local.kubemind-system.svc.cluster.local
                │
                ▼
Local Workstation Running Ollama
            (qwen3:8b)
                │
        ┌───────┴────────┐
        ▼                ▼
Root Cause        Remediation
 Analysis          Suggestions
        │                │
        └───────┬────────┘
                ▼
       Slack Incident Alerts
```

---

# 🛠️ Prerequisites

Before running KubeMind, ensure the following tools are installed and available in your shell.

| Tool               | Minimum Version   | Purpose                        |
| ------------------ | ----------------- | ------------------------------ |
| Go                 | 1.20+             | Compile and run the operator   |
| Docker             | 20.10+            | Build container images         |
| Helm               | v3+               | Package and deploy charts      |
| kubectl            | v1.24+            | Interact with Kubernetes       |
| Kubernetes Cluster | Any local cluster | Minikube, Kind, Docker Desktop |
| Ollama             | Latest            | Local AI inference engine      |

Verify your environment:

```bash
go version
docker version
helm version --short
kubectl version --client
```

---

# 🧠 Local AI Engine Setup (Ollama)

KubeMind uses a locally hosted LLM to provide:

* Complete data privacy
* No cloud inference costs
* Faster diagnostics
* Offline troubleshooting capabilities

## Install Ollama

Download and install Ollama from:

https://ollama.com

## Download and Run the Model

KubeMind is optimized for the **Qwen 3 8B** model.

```bash
# Optional: Enable parallel inference
export OLLAMA_NUM_PARALLEL=4

# Download model
ollama pull qwen3:8b

# Start local inference server
ollama run qwen3:8b
```

## Verify Ollama

```bash
curl http://localhost:11434
```

Expected output:

```text
Ollama is running
```

---

# 🌐 Networking Architecture

The KubeMind Operator runs inside Kubernetes, while Ollama runs on your workstation.

Because Pods cannot directly access your local `localhost`, KubeMind uses a Kubernetes DNS bridge.

```text
Operator Pod
      │
      ▼
ollama-local.kubemind-system.svc.cluster.local
      │
      ▼
ExternalName Service
      │
      ▼
host.docker.internal
      │
      ▼
Local Ollama Server (Port 11434)
```

This approach eliminates the need for hardcoded workstation IP addresses.

---

# 🚀 Production Deployment Guide

## Step 1 — Create Namespaces

```bash
kubectl create namespace kubemind-system
kubectl create namespace kubemind
```

---

## Step 2 — Configure Slack Notifications

Create the Slack webhook secret used by KubeMind to send incident notifications.

```bash
kubectl create secret generic slack-secret \
  --from-literal=webhook-url="https://hooks.slack.com/services/YOUR/WEBHOOK/URL" \
  -n kubemind-system
```

---

## Step 3 — Deploy the Ollama Bridge

Apply the networking bridge manifest:

```bash
kubectl apply -f deploy/test-apps/ollama-bridge.yaml
```

Example bridge:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ollama-local
  namespace: kubemind-system
spec:
  type: ExternalName
  externalName: host.docker.internal
```

---

## Step 4 — Build the Operator Image

```bash
docker build -t akshayraina/kubemind-operator:v0.1.0 .
```

---

## Step 5 — Install the Helm Chart

```bash
helm install kubemind-operator charts/kubemind-operator \
  --namespace kubemind-system \
  --set image.pullPolicy=Always
```

---

## Step 6 — Verify Deployment

Check that the operator Pod is running:

```bash
kubectl get pods -n kubemind-system
```

Stream operator logs:

```bash
kubectl logs \
  -n kubemind-system \
  -l app.kubernetes.io/name=kubemind-operator \
  -f
```

---

# 🖥️ Local Development Mode

During active development, you can run the operator directly from source without rebuilding containers.

Configure environment variables:

```bash
export OLLAMA_URL="http://localhost:11434"
export MODEL_NAME="qwen3:8b"
```

Start the operator:

```bash
go run ./cmd/manager/main.go
```

---

# 🧪 Validation Workflows

Deploy intentionally broken workloads to validate the end-to-end troubleshooting pipeline.

```bash
kubectl apply -f deploy/test-apps/restart-app.yaml -n kubemind

kubectl apply -f deploy/test-apps/crash-app.yaml -n kubemind

kubectl apply -f deploy/test-apps/oom-app.yaml -n kubemind
```

Once deployed, KubeMind will:

1. Detect the anomaly.
2. Gather cluster diagnostics.
3. Generate structured AI analysis.
4. Produce remediation suggestions.
5. Send formatted incident reports to Slack.

---

# 🧹 Cleanup & Teardown

Remove the operator and all test resources:

```bash
helm uninstall kubemind-operator -n kubemind-system

kubectl delete ns kubemind

kubectl delete svc ollama-local -n kubemind-system
```

---

# 🎉 Success

Your KubeMind Operator is now fully configured and ready to perform AI-assisted Kubernetes Day-2 operations, diagnostics, and incident triage.
