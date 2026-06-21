#!/usr/bin/env bash
# engine/verifier.sh

PLAN_FILE=$1
NAMESPACE="kubemind"

if [ ! -f "$PLAN_FILE" ]; then
    echo "UNKNOWN"
    exit 0
fi

# Retrieve resource keys directly from the JSON plan properties
RESOURCE_KIND=$(jq -r '.resource_type' "$PLAN_FILE" | tr '[:upper:]' '[:lower:]')
RESOURCE_NAME=$(jq -r '.resource_name' "$PLAN_FILE")

if [ -z "$RESOURCE_NAME" ] || [ "$RESOURCE_NAME" = "null" ]; then
    echo "UNRESOLVED"
    exit 0
fi

# Check status natively without passing redundant resource strings
STATUS=$(kubectl get "$RESOURCE_KIND" "$RESOURCE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null || echo "DELETED")

if [ "$STATUS" = "Running" ] || [ "$STATUS" = "Succeeded" ]; then
    echo "✅ RESOLVED"
else
    echo "⚠️ UNRESOLVED ($STATUS)"
fi