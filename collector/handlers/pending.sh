#!/bin/bash

POD_NAME=$1

echo "[Pending Handler]"

kubectl describe pod "$POD_NAME" \
-n kubemind \
> "logs/context/${POD_NAME}_pending.txt"