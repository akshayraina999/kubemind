# 🏗️ KubeMind Core Architecture Blueprint

KubeMind approaches automated operations with a zero-trust model toward LLM output. Instead of executing generated strings blindly, the platform splits the failure-resolution lifecycle into four decoupled execution rings:


[ Active Cluster Anomaly ]
│
▼

DATA COLLECTION  ──> (Transforms volatile state events & logs into text payload)
│
▼

LOCAL INFERENCE  ──> (Ollama / qwen3:8b enforces strict schema generation boundaries)
│
▼

SANITY FILTER    ──> (Bypasses LLM hallucinations / cuts off truncated string bugs)
│
▼

RISK TARGETING   ──> (Executes low-risk changes; halts on state mutations)


## 1. The Core Execution Rings

### Data Collection & Token Isolation (`engine/collector.sh`)
When an anomaly passes cluster scanning, the engine bundles its context (Kubernetes configuration state manifest, descriptive events, standard descriptor vectors, and text logs) into a compressed staging document. 

Once compiled, it checks the cryptographic hash signature database inside `state/processed/`. If an anomaly state pattern matches a known tracking ledger signature, the engine cancels execution immediately, stopping compute resource loops on duplicate alerts.

### Constrained Inference Engine (`ai/analyze.sh`)
To keep operational processing fast and secure on unified hardware frameworks (such as a 16GB Apple Silicon MacBook Air), KubeMind uses a local `Ollama` layer running `qwen3:8b` configured with strict schema enforcement constraints. 

By hard-coding parameter options (`num_predict: 4096`, `temperature: 0.0`), the system controls text token paths, suppressing chat output and forcing direct JSON outputs.

### The Shell Sanity Filtering Matrix (`engine/remediator.sh`)
LLMs often struggle with nested backslash character string escaping rules (`\\\\\"`). This can cause truncation failures where smaller models split arrays mid-string to meet global document requirements. 

KubeMind implements a rigorous validation loop that interceptively parses execution strings using character pattern detection rules:
* Flags mid-payload array breaks (`* "}{" *`).
* Catches structural string truncation anomalies.
* Safeguards against empty string executions.

---

## ⚡ 2. Immutable Lifecycle Adaptation Logic
A standard platform anti-pattern is attempting to apply standard `kubectl patch` mutations to completely immutable structural entities—specifically Kubernetes `Job` resource matrices.

To combat this, the systemic prompt matrix configures the underlying inference engine to automatically shift patterns when tracking a `FailedJob`. Instead of a standard patch instruction payload, the engine drops the mutating step completely and switches to a deterministic sequential pattern:
1. `Delete` target resource block (`kubectl delete job/...`).
2. `Create` clean state replacement using multi-line `stdin` streams (`kubectl apply -f - <<EOF`).