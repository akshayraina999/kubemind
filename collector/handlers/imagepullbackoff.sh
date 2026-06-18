#!/bin/bash

POD_NAME=$1

echo "[ImagePullBackOff Handler]"

./collector/collect_events.sh "$POD_NAME"

kubectl get pod "$POD_NAME" \
-n kubemind \
-o jsonpath='{.spec.containers[*].image}' \
> "logs/context/${POD_NAME}_image.txt"