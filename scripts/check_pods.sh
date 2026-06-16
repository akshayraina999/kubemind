#!/bin/bash

NAMESPACE="kubemind"

kubectl get pods -n "$NAMESPACE" -o json \
| jq -r '
.items[]
| select(.status.phase != "Running")
| .metadata.name
'