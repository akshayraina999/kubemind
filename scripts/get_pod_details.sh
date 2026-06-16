#!/bin/bash

NAMESPACE="kubemind"

for pod in $(kubectl get pods -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}')
do
    echo "================================="
    echo "Pod: $pod"
    echo "================================="

    kubectl get pod "$pod" \
        -n "$NAMESPACE" \
        -o json | jq -r '

        .status.containerStatuses[]? |
        {
            name: .name,
            ready: .ready,
            reason: (
                .state.waiting.reason //
                .state.terminated.reason //
                "Running"
            )
        }
        '
done