#!/bin/bash

POD_NAME=$1
NAMESPACE="kubemind"

kubectl describe pod "$POD_NAME" \
-n "$NAMESPACE" \
| sed -n '/Events:/,$p' \
> logs/events/${POD_NAME}.log
