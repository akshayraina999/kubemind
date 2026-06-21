#!/usr/bin/env bash
# engine/collector.sh
set -euo pipefail

RESOURCE_NAME=${1:-""}
NAMESPACE=${2:-"kubemind"}

if [[ -z "$RESOURCE_NAME" ]]; then
    echo "❌ Error: Missing resource arguments for context collection target."
    exit 1
fi

CONTEXT_FILE="./logs/context/${RESOURCE_NAME}.txt"
mkdir -p "$(dirname "$CONTEXT_FILE")"

# Establish a clean context file footprint
echo "=== KUBEMIND CONTEXT GENERATION ===" > "$CONTEXT_FILE"
echo "Timestamp: $(date)" >> "$CONTEXT_FILE"
echo "Target Resource: $RESOURCE_NAME" >> "$CONTEXT_FILE"
echo "Namespace: $NAMESPACE" >> "$CONTEXT_FILE"
echo "-----------------------------------" >> "$CONTEXT_FILE"

# Extract the raw object manifest layout state
echo "=== Active Manifest Configuration Spec ===" >> "$CONTEXT_FILE"
kubectl get pod "$RESOURCE_NAME" -n "$NAMESPACE" -o yaml 2>/dev/null >> "$CONTEXT_FILE" || true

# Extract localized operational failure event tracks
echo "=== Associated System Timeline Events ===" >> "$CONTEXT_FILE"
kubectl get events -n "$NAMESPACE" --field-selector involvedObject.name="$RESOURCE_NAME" >> "$CONTEXT_FILE" 2>/dev/null || true

# Extract container termination trace histories
echo "=== Running Stdout/Stderr Execution Logs ===" >> "$CONTEXT_FILE"
kubectl logs "$RESOURCE_NAME" -n "$NAMESPACE" --tail=100 2>/dev/null >> "$CONTEXT_FILE" || true

# ====================================================================
# 🚀 OPTIMIZED DEEP PLATFORM ENRICHMENT HARVESTER
# ====================================================================
CLASS_LABEL=$(./engine/classifier.sh "$RESOURCE_NAME" "$NAMESPACE")

# Case 1: Intercept Storage Engine Stalls - Pull STATUS, not complete object yaml definitions
if [[ "$CLASS_LABEL" == "PVCBindingStall" || "$CLASS_LABEL" == "VolumeMountFailure" ]]; then
    echo "=== Storage Infrastructure Summary ===" >> "$CONTEXT_FILE"
    kubectl get pvc,pv -n "$NAMESPACE" --no-headers 2>/dev/null >> "$CONTEXT_FILE" || true
    echo "=== Storage Classes ===" >> "$CONTEXT_FILE"
    kubectl get storageclass --no-headers 2>/dev/null >> "$CONTEXT_FILE" || true
fi

# Case 2: Intercept Scheduling Taint & Node Constraints - Pull Allocations & Taints ONLY
if [[ "$CLASS_LABEL" == "TaintTolerationMismatch" || "$CLASS_LABEL" == "NodeResourceExhaustion" ]]; then
    echo "=== Host Node Resource Allocation & Taints ===" >> "$CONTEXT_FILE"
    kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints,CPU_ALLOCATABLE:.status.allocatable.cpu,MEM_ALLOCATABLE:.status.allocatable.memory 2>/dev/null >> "$CONTEXT_FILE" || true
fi

# Case 3: Intercept Pod Disruption Budget Eviction Lockouts - Pull lightweight tabular specs
if [[ "$CLASS_LABEL" == "PDBViolation" ]]; then
    echo "=== Active Pod Disruption Budgets ===" >> "$CONTEXT_FILE"
    kubectl get pdb -n "$NAMESPACE" --no-headers 2>/dev/null >> "$CONTEXT_FILE" || true
fi

echo "-----------------------------------" >> "$CONTEXT_FILE"
echo "Context successfully compressed and written to $CONTEXT_FILE"