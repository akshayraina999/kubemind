#!/bin/bash

POD_NAME=$1
NAMESPACE="kubemind"

OUTPUT_FILE="logs/context/${POD_NAME}.txt"

mkdir -p logs/context

POD_STATUS=$(kubectl get pod "$POD_NAME" \
-n "$NAMESPACE" \
-o jsonpath='{.status.containerStatuses[0].state.waiting.reason}' 2>/dev/null)

NODE_NAME=$(kubectl get pod "$POD_NAME" \
-n "$NAMESPACE" \
-o jsonpath='{.spec.nodeName}')

RESTARTS=$(kubectl get pod "$POD_NAME" \
-n "$NAMESPACE" \
-o jsonpath='{.status.containerStatuses[0].restartCount}')

cat > "$OUTPUT_FILE" << EOF
Pod Name: $POD_NAME
Namespace: $NAMESPACE
Node: $NODE_NAME
Status: $POD_STATUS
Restarts: $RESTARTS

Events:
$(cat logs/events/${POD_NAME}.log)
EOF

echo "Context saved to $OUTPUT_FILE"