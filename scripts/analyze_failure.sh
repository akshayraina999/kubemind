#!/bin/bash

POD_NAME=$1

echo "=================================="
echo "AI INPUT PAYLOAD"
echo "=================================="

cat logs/events/${POD_NAME}.log