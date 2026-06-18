#!/bin/bash

POD_NAME=$1

mkdir -p logs/pods

kubectl logs \
"$POD_NAME" \
-n kubemind \
--tail=100 \
> "logs/pods/${POD_NAME}.log" 2>/dev/null
