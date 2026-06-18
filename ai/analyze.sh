#!/bin/bash

POD_NAME=$1

CONTEXT_FILE="logs/context/${POD_NAME}.txt"

if [ ! -f "$CONTEXT_FILE" ]; then
    echo "Context file not found."
    echo "Run generate_context.sh first."
    exit 1
fi

CONTEXT=$(cat "$CONTEXT_FILE")

PROMPT=$(cat <<EOF
You are a Senior Kubernetes Site Reliability Engineer.

Analyze the following Kubernetes issue.

Respond ONLY in this format:

Root Cause:
<root cause>

Explanation:
<explanation>

Suggested Fix:
<fix>

Cluster Context:

$CONTEXT
EOF
)

jq -n \
  --arg model "qwen3:8b" \
  --arg prompt "$PROMPT" \
  '{
      model: $model,
      prompt: $prompt,
      stream: false
   }' |
curl -s http://localhost:11434/api/generate \
-H "Content-Type: application/json" \
-d @- |
jq -r '.response'