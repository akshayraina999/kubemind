# KubeMind Operator — Local Development Prerequisites & Onboarding

Welcome to the `refactor/kubemind-operator` branch. This branch implements a hybrid **Kubernetes Day-2 Operator Architecture**. The operator runs containerized inside your cluster (or natively via `go run`) and routes cluster diagnostic contexts back to a local AI inference engine running on your workstation.

Follow this guide to get your local environment fully configured.

---

## 🛠️ 1. Core Tooling Requirements

Before running the operator, ensure you have the following command-line tools installed and configured on your workstation:

| Tool | Minimum Version | Installation Command (macOS Homebrew) | Purpose |
| :--- | :--- | :--- | :--- |
| **Go** | `1.20+` (Using `v1.26.4`) | `brew install go` | Compiles and runs the operator binary |
| **Docker** | `20.10+` | `brew install --cask docker` | Builds the operator runtime container image |
| **Helm** | `v3.0+` | `brew install helm` | Packages, lints, and installs the operator chart |
| **Kubectl** | `v1.24+` | `brew install kubectl` | Interacts with your active Kubernetes cluster |

Verify your local installations by running:
```zsh
go version
docker version
helm version --short
kubectl version --client