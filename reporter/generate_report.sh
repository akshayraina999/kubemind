#!/bin/bash

POD_NAME=$1

TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")

REPORT_FILE="reports/${TIMESTAMP}_${POD_NAME}.md"

AI_OUTPUT_FILE="logs/ai/${POD_NAME}.txt"

cat > "$REPORT_FILE" <<EOF
# KubeMind RCA Report

Generated: $(date)

## Pod

$POD_NAME

## AI Analysis

$(cat "$AI_OUTPUT_FILE")

EOF

echo "$REPORT_FILE"