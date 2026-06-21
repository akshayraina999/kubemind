#!/usr/bin/env bash
# engine/classifier.sh
set -euo pipefail

RESOURCE_NAME=${1:-""}
NAMESPACE=${2:-"kubemind"}

if [[ -z "$RESOURCE_NAME" ]]; then
    echo "UnknownAnomaly"
    exit 0
fi

# Fetch status conditions and last events for the targeted resource
POD_STATUS=$(kubectl get pod "$RESOURCE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
STATUS_REASON=$(kubectl get pod "$RESOURCE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.containerStatuses[0].state.waiting.reason}' 2>/dev/null || echo "")
EVENT_MESSAGE=$(kubectl get events -n "$NAMESPACE" --field-selector involvedObject.name="$RESOURCE_NAME" --sort-by='.metadata.creationTimestamp' -o jsonpath='{.items[-1:].message}' 2>/dev/null || echo "")

# ====================================================================
# 🔍 ANOMALY CLASSIFICATION MATRIX
# ====================================================================

# 1. Native Container Runtime Failure Signatures
if [[ "$STATUS_REASON" == "ImagePullBackOff" || "$STATUS_REASON" == "ErrImagePull" ]]; then
    echo "ImagePullBackOff"
elif [[ "$STATUS_REASON" == "CrashLoopBackOff" ]]; then
    echo "CrashLoopBackOff"
elif [[ "$EVENT_MESSAGE" == *"OOMKilled"* || "$EVENT_MESSAGE" == *"Exit Code 137"* ]]; then
    echo "OOMKilled"

# 2. Storage / PersistentVolumeClaim Bottlenecks
elif [[ "$EVENT_MESSAGE" == *"waiting for a volume to be created"* || "$EVENT_MESSAGE" == *"VolumeBindingImmediate"* ]]; then
    echo "PVCBindingStall"
elif [[ "$EVENT_MESSAGE" == *"FailedAttachVolume"* || "$EVENT_MESSAGE" == *"FailedMount"* ]]; then
    echo "VolumeMountFailure"

# 3. Node Resource Constraints & Toleration/Taint Conflicts
elif [[ "$EVENT_MESSAGE" == *"node(s) had untolerated taint"* ]]; then
    echo "TaintTolerationMismatch"
elif [[ "$EVENT_MESSAGE" == *"Insufficient memory"* || "$EVENT_MESSAGE" == *"Insufficient cpu"* ]]; then
    echo "NodeResourceExhaustion"

# 4. Eviction / Pod Disruption Budget Stalls
elif [[ "$EVENT_MESSAGE" == *"FailedEviction"* || "$EVENT_MESSAGE" == *"disruption budget"* ]]; then
    echo "PDBViolation"

# Fallback Default Routing
elif [[ "$POD_STATUS" == "Pending" ]]; then
    echo "Pending"
elif [[ "$EVENT_MESSAGE" == *"Error"* ]]; then
    echo "FailedJob"
else
    echo "UnclassifiedAnomaly"
fi