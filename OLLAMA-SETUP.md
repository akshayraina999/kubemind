# 🌐 Connecting KubeMind to a Local Ollama Instance

KubeMind Operator performs AI-powered diagnostics by sending Kubernetes cluster context to a locally running Large Language Model (LLM). Since the operator runs inside Kubernetes while Ollama runs on your workstation, a network bridge is required to allow secure communication between the cluster and your local machine.

KubeMind uses a Kubernetes `ExternalName` Service to expose the local Ollama runtime to the cluster in a platform-agnostic and maintenance-free manner.

This approach eliminates the need for:

* Hardcoded workstation IP addresses
* Manual endpoint management
* Frequent network reconfiguration when switching Wi-Fi networks

It works seamlessly with:

* Minikube
* Kind
* Docker Desktop
* Most local Kubernetes environments

---

## 📁 Step 1: Create the Ollama Bridge Manifest

Create the following file:

```text
deploy/test-apps/ollama-bridge.yaml
```

Add the following configuration:

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

This creates an internal DNS alias inside the cluster that automatically routes traffic to your workstation.

Apply the manifest:

```bash
# Remove legacy configurations if present
kubectl delete svc ollama-local \
  -n kubemind-system \
  --ignore-not-found

kubectl delete endpoints ollama-local \
  -n kubemind-system \
  --ignore-not-found

# Create the new bridge
kubectl apply -f deploy/test-apps/ollama-bridge.yaml
```

---

## ⚙️ Step 2: Configure the Operator

Update your Helm values file:

```text
charts/kubemind-operator/values.yaml
```

Configure the environment variables:

```yaml
env:
  - name: OLLAMA_URL
    value: "http://ollama-local.kubemind-system.svc.cluster.local:11434"

  - name: MODEL_NAME
    value: "qwen3:8b"
```

Deploy the updated configuration:

```bash
helm upgrade kubemind-operator charts/kubemind-operator \
  --namespace kubemind-system \
  --rollback-on-failure
```

---

## 🔍 Step 3: Verify Connectivity

After the operator has been upgraded, verify that Kubernetes can successfully communicate with Ollama.

Launch a temporary test pod:

```bash
kubectl run curl-test \
  --rm -it \
  --image=curlimages/curl \
  --restart=Never \
  -- curl -X POST \
  http://ollama-local.kubemind-system.svc.cluster.local:11434/api/generate \
  -d '{
        "model":"qwen3:8b",
        "prompt":"say pong",
        "stream":false
      }'
```

Expected output:

```json
{
  "model": "qwen3:8b",
  "response": "pong",
  "done": true
}
```

A successful response confirms that:

* Kubernetes DNS resolution is functioning correctly.
* The operator can reach the local Ollama endpoint.
* AI analysis requests can be processed successfully.

---

## 🏗️ Architecture Overview

```text
+-------------------------------------------------------------+
|                     Kubernetes Cluster                      |
|                                                             |
|   +----------------------------+                            |
|   |     KubeMind Operator      |                            |
|   +-------------+--------------+                            |
|                 |                                           |
|                 | HTTP Request                              |
|                 v                                           |
|   ollama-local.kubemind-system.svc.cluster.local           |
|                 |                                           |
|                 v                                           |
|          CoreDNS Resolution Layer                           |
|                 |                                           |
|                 v                                           |
|        ExternalName Service Mapping                         |
|                 |                                           |
|                 v                                           |
|         host.docker.internal                                |
+-----------------+-------------------------------------------+
                  |
                  v
        Workstation Host Machine
                  |
                  v
        Local Ollama Runtime
             (Port 11434)
```

---

## 🔧 How It Works

1. The KubeMind Operator detects a Kubernetes anomaly.

2. Diagnostic context is collected and prepared.

3. The operator sends an HTTP request to:

   ```text
   http://ollama-local.kubemind-system.svc.cluster.local:11434
   ```

4. CoreDNS resolves the request using the `ExternalName` Service.

5. The request is transparently redirected to:

   ```text
   host.docker.internal
   ```

6. The request reaches the Ollama daemon running on your workstation.

7. The LLM performs analysis and returns structured remediation guidance to the operator.

---

## ✅ Benefits of This Approach

* No hardcoded workstation IP addresses.
* No manual endpoint updates.
* Works across different networks and environments.
* Fully Kubernetes-native DNS routing.
* Simple, portable, and easy to maintain.

KubeMind can now securely leverage local LLM inference while remaining fully containerized inside Kubernetes.
