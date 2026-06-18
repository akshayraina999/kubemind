#!/bin/bash

POD_NAME=$1

echo "[OOMKilled Handler]"

./collector/collect_events.sh "$POD_NAME"

./collector/collect_logs.sh "$POD_NAME"

kubectl describe pod "$POD_NAME" -n kubemind \
> "logs/context/${POD_NAME}_resources.txt"