# 📦 KubeMind Operator Installation Guide

Follow the steps below to install and run the **KubeMind Operator** in your local Kubernetes environment.

---

## 1️⃣ Clone the Repository

Clone the operator branch and move into the project directory.

```bash
git clone -b refactor/kubemind-operator https://github.com/akshayraina999/kubemind.git

cd kubemind
```

---

## 2️⃣ Verify Prerequisites

Ensure the following tools are installed and available in your shell.

| Tool               | Minimum Version                  |
| ------------------ | -------------------------------- |
| Go                 | 1.20+                            |
| Docker             | Latest                           |
| Helm               | v3+                              |
| kubectl            | v1.24+                           |
| Kubernetes Cluster | Minikube / Kind / Docker Desktop |
| Ollama             | Latest                           |

Verify your environment:

```bash
go version
docker version
helm version --short
kubectl version --client
```

---

## 3️⃣ Start the Local AI Engine

KubeMind uses a **locally hosted LLM** for diagnostics and remediation generation.

### Configure Ollama for Parallel Processing (Recommended)

To avoid request bottlenecks when multiple cluster failures occur simultaneously, configure Ollama for parallel execution.

```bash
export OLLAMA_NUM_PARALLEL=4
```

### Download the Required Model

```bash
ollama pull qwen3:8b
```

### Start the Model

```bash
ollama run qwen3:8b
```

Verify that Ollama is running:

```bash
curl http://localhost:11434
```

Expected output:

```text
Ollama is running
```

---

## 4️⃣ Create Required Namespaces

Create the namespaces used by KubeMind.

```bash
kubectl create namespace kubemind-system

kubectl create namespace kubemind
```

---

## 5️⃣ Configure Slack Notifications

Create a Kubernetes Secret containing your Slack Webhook URL.

```bash
kubectl create secret generic slack-secret \
  --from-literal=webhook-url="https://hooks.slack.com/services/YOUR/WEBHOOK/URL" \
  -n kubemind-system
```

---

## 6️⃣ Create the Ollama Network Bridge

The KubeMind Operator runs inside Kubernetes, while Ollama runs on your local workstation.

To allow the operator to communicate with Ollama, create an `ExternalName` Service that forwards requests to your host machine.

Create a file named:

```text
deploy/test-apps/ollama-bridge.yaml
```

Add the following content:

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

Apply the manifest:

```bash
kubectl apply -f deploy/test-apps/ollama-bridge.yaml
```

Verify that the service was created successfully:

```bash
kubectl get svc -n kubemind-system ollama-local
```

---

## 7️⃣ Install the Operator

### Validate the Helm Chart

```bash
helm lint charts/kubemind-operator
```

### Deploy the Operator

```bash
helm install kubemind-operator charts/kubemind-operator \
  --namespace kubemind-system \
  --set image.pullPolicy=Always
```

---

## 8️⃣ Verify the Installation

Ensure that the operator Pod is running successfully.

```bash
kubectl get pods -n kubemind-system
```

Expected output:

```text
NAME                                    READY   STATUS    RESTARTS   AGE
kubemind-operator-xxxxx                1/1     Running   0          30s
```

Stream the operator logs:

```bash
kubectl logs \
  -n kubemind-system \
  -l app.kubernetes.io/name=kubemind-operator \
  -f
```

You should begin seeing reconciliation events and diagnostic processing logs.

---

# 🧪 Validate the Setup

Deploy a sample faulty workload to trigger the AI troubleshooting workflow.

```bash
kubectl apply -f deploy/test-apps/restart-app.yaml -n kubemind
```

Within a few seconds, KubeMind will:

* 🔍 Detect the workload failure automatically.
* 📥 Collect Kubernetes events, logs, and manifests.
* 🧠 Send diagnostic context to the local Ollama model.
* 📄 Generate structured root cause analysis.
* 🛠️ Produce remediation recommendations.
* 📢 Publish a detailed incident report directly to Slack.

---

# 🔄 High-Level Workflow

```text
Broken Workload
       │
       ▼
+-------------------+
| KubeMind Operator |
+-------------------+
       │
       ▼
Collect Logs + Events
       │
       ▼
Send Context to Ollama
       │
       ▼
AI Root Cause Analysis
       │
       ▼
Generate Remediation Plan
       │
       ▼
Publish Incident to Slack
```

---

# 🎉 Success

Your **KubeMind Operator** is now fully installed and ready to provide **AI-assisted Kubernetes troubleshooting, automated diagnostics, and intelligent incident reporting** for your cluster.
