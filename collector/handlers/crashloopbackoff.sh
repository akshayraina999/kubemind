#!/bin/bash

POD_NAME=$1

echo "[CrashLoopBackOff Handler]"

./collector/collect_events.sh "$POD_NAME"

./collector/collect_logs.sh "$POD_NAME"

kubectl get pod "$POD_NAME" \
-n kubemind \
-o jsonpath='{.status.containerStatuses[0].restartCount}' \
> "logs/context/${POD_NAME}_restarts.txt"