#!/bin/bash

POD_NAME=$1

echo "[Generic Handler]"

./collector/collect_events.sh "$POD_NAME"

./collector/collect_logs.sh "$POD_NAME"