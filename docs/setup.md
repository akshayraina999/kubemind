Markdown
# 🚀 KubeMind Infrastructure Setup & Run-Book

Follow this configuration sequence to spin up the local KubeMind validation testing pipeline using an Apple Silicon MacBook or an equivalent development workspace.

---

## 1. Local Prerequisites Configuration

Before initiating the pipeline loop, ensure your development workstation has the following core command-line binaries installed and globally accessible in your shell `$PATH`:

* **Minikube (v1.32+)** or an active local Docker Desktop context
* **Ollama Engine** (Running locally via background daemon)
* **jq** (High-performance lightweight command-line JSON processor)

---

## 2. LLM Engine Model Hydration

Ensure your local Ollama background engine daemon is fully active and initialized, then execute the standard model fetch to pull down the optimized engineering base model target:

```bash
ollama pull qwen3:8b
To verify that the model's localized JSON API endpoint router interface is responding properly over your machine's unified memory bus, execute a raw validation curl test string:

Bash
curl -i -X POST http://localhost:11434/api/generate \
  -d '{"model": "qwen3:8b", "prompt": "Output json: {\"status\": \"ok\"}", "stream": false, "format": "json"}'
3. Sandboxed Application Manifest Provisioning
Apply the provided buggy testing manifest definitions straight to your cluster to give the KubeMind scanning engine distinct anomaly vectors to isolate and troubleshoot:

Bash
# Create the isolated environment namespace validation space
kubectl create namespace kubemind

# Hydrate the cluster with the sample failure application suites
kubectl apply -f kubernetes/test-apps/broken-app.yaml -n kubemind
kubectl apply -f kubernetes/test-apps/oom-app.yaml -n kubemind
kubectl apply -f kubernetes/test-apps/failed-job.yaml -n kubemind
4. Initializing Clean State Execution Runs
To test your local environment with zero clutter or state metadata history, clear the logging tracing tracking rows and trigger your primary master pipeline intake suite:

Bash
# Set up persistent empty logging framework paths cleanly
mkdir -p logs/ai logs/context logs/events logs/pods logs/remediation logs/remediation-plans logs/verification reports state/processed

# Seed empty tracking spaces with .gitkeep placeholders to maintain Git tree structure
for dir in logs/ai logs/context logs/events logs/pods logs/remediation logs/remediation-plans logs/verification reports state/processed; do
    touch "$dir/.gitkeep"
done

# Clear out duplicate state tracker memory tokens and fire up the master loop once
./cmd/kubemind.sh clear-state
./cmd/kubemind.sh once