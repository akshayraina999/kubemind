#!/usr/bin/env bash
# ai/analyze.sh
set -euo pipefail

RESOURCE_NAME=${1:-""}
NAMESPACE=${2:-"kubemind"}

if [[ -z "$RESOURCE_NAME" ]]; then
    echo "❌ Error: Missing targeted anomaly resource identifier name parameter."
    exit 1
fi

# Configuration properties fallback boundaries
MODEL_NAME="qwen3:8b"
OLLAMA_URL="http://localhost:11434/api/generate"
CONTEXT_FILE="./logs/context/${RESOURCE_NAME}.txt"
PLAN_OUTPUT="./logs/remediation-plans/${RESOURCE_NAME}.json"

if [[ ! -f "$CONTEXT_FILE" ]]; then
    echo "❌ Error: Target application context file not found at $CONTEXT_FILE."
    exit 1
fi

# ====================================================================
# 🎯 DETERMINISTIC PARENT CONTROLLER LOOKUP
# ====================================================================
PARENT_OWNER=$(kubectl get pod "$RESOURCE_NAME" -n "$NAMESPACE" -o jsonpath='{.metadata.ownerReferences[0].name}' 2>/dev/null || echo "")

if [[ "$PARENT_OWNER" == *-[0-9a-f][0-9a-f]* ]]; then
    PARENT_CONTROLLER=$(kubectl get replicaset "$PARENT_OWNER" -n "$NAMESPACE" -o jsonpath='{.metadata.ownerReferences[0].name}' 2>/dev/null || echo "$PARENT_OWNER")
else
    PARENT_CONTROLLER=$PARENT_OWNER
fi

echo "" >> "$CONTEXT_FILE"
echo "=== KUBEMIND DETERMINISTIC ENGINE RESOLUTION ===" >> "$CONTEXT_FILE"
echo "Calculated Root Owner Controller Name: $PARENT_CONTROLLER" >> "$CONTEXT_FILE"
# ====================================================================

# Gather standard system prompts
RCA_PROMPT=$(cat ./ai/prompts/rca_prompt.txt 2>/dev/null || echo "Analyze root cause.")
REMEDIATION_PROMPT=$(cat ./ai/prompts/remediation_prompt.txt 2>/dev/null || echo "Output JSON mapping updates.")

# ====================================================================
# 🛡️ XML STRATEGIC SCHEMA ISOLATION (FIXED CLOSING QUOTE LOCATION)
# ====================================================================
PROMPT_BODY="<system_instructions>
You are the KubeMind AI Engine core automation router module. Your goal is to process the following cluster anomaly data and output a strictly valid JSON object matching the requested schema. 

CRITICAL: You MUST use these exact structural keys in your root JSON output object. Do not rename them:
- \"root_cause\": (String explaining the root failure)
- \"confidence\": (Float between 0.0 and 1.0)
- \"remediation\": (Array of remediation step objects)
- \"verification\": (Array of verification strings)

Do not use keys like 'root_cause_analysis' or 'remediation_steps'. Match the properties exactly.
</system_instructions>

<cluster_telemetry>
$(cat "$CONTEXT_FILE")
</cluster_telemetry>

<analysis_rules>
$RCA_PROMPT
</analysis_rules>

<remediation_rules>
$REMEDIATION_PROMPT
</remediation_rules>

YOUR JSON RESPONSE:"
# ====================================================================

echo "🧠 Sending $RESOURCE_NAME context to Ollama ($MODEL_NAME) with forced JSON constraints..."

# Compile API payload map using robust structural boundaries
PAYLOAD=$(jq -n \
  --arg model "$MODEL_NAME" \
  --arg prompt "$PROMPT_BODY" \
  '{
    model: $model,
    prompt: $prompt,
    stream: false,
    format: "json",
    options: {
      num_predict: 4096,
      temperature: 0.0,
      num_ctx: 4096
    }
  }')

# Fire endpoint request transaction stream
RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "$PAYLOAD" "$OLLAMA_URL")

# Validate tracking transaction states
if [[ -z "$RESPONSE" ]]; then
    echo "❌ Critical Error: Ollama interface transaction dropped or timed out."
    exit 1
fi

# Extract text payload directly from tracking tokens
RAW_PLAN=$(echo "$RESPONSE" | jq -r '.response' 2>/dev/null || echo "")

if [[ -z "$RAW_PLAN" || "$RAW_PLAN" == "null" ]]; then
    echo "❌ Critical Error: Ollama output was structural garbage, empty, or truncated."
    echo "$RESPONSE" > "${PLAN_OUTPUT}.broken"
    exit 1
fi

# Verify compiled plan properties syntax validity
if echo "$RAW_PLAN" | jq . >/dev/null 2>&1; then
    echo "$RAW_PLAN" | jq . > "$PLAN_OUTPUT"
    echo "✅ Valid Remediation Plan compiled: $PLAN_OUTPUT"
else
    echo "❌ Critical Error: AI schema output contains malformed data blocks."
    echo "$RAW_PLAN" > "${PLAN_OUTPUT}.broken"
    exit 1
fi