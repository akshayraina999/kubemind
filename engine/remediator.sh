#!/usr/bin/env bash
# engine/remediator.sh
set -euo pipefail

PLAN_FILE=${1:-""}

if [[ -z "$PLAN_FILE" || ! -f "$PLAN_FILE" ]]; then
    echo "❌ Error: Remediation plan file path required or file does not exist."
    exit 1
fi

echo "⚙️ Initializing KubeMind Remediator Engine..."
CONFIDENCE=$(jq -r '.confidence // 0.0' "$PLAN_FILE")
echo "📊 AI Engine Confidence Score: $CONFIDENCE"

STEPS_COUNT=$(jq '.remediation | length' "$PLAN_FILE")

# Master loop scanning through the generated plan array
for ((i=0; i<STEPS_COUNT; i++)); do
    ACTION_TYPE=$(jq -r ".remediation[$i].action_type" "$PLAN_FILE")
    TARGET_KIND=$(jq -r ".remediation[$i].target_kind" "$PLAN_FILE")
    TARGET_NAME=$(jq -r ".remediation[$i].target_name" "$PLAN_FILE")
    RISK=$(jq -r ".remediation[$i].risk // \"Medium\"" "$PLAN_FILE")
    AUTO_EXEC=$(jq -r ".remediation[$i].auto_executable // false" "$PLAN_FILE")

    # ====================================================================
    # 🎯 REDIRECT TARGET TO DEPLOYMENT CONTROLLER NATIVELY
    # ====================================================================
    # Read the pre-calculated true deployment name from the context metadata file
    # Handles lookup even if the model passes a pod or replicaset name
    ORIGINAL_TARGET_NAME="$TARGET_NAME"
    CONTEXT_FILE="./logs/context/${TARGET_NAME}.txt"
    
    if [[ -f "$CONTEXT_FILE" ]]; then
        TRUE_DEPLOYMENT=$(grep "Calculated Root Owner Controller Name:" "$CONTEXT_FILE" | awk -F': ' '{print $2}' | tr -d '[:space:]')
        if [[ -n "$TRUE_DEPLOYMENT" ]]; then
            TARGET_KIND="Deployment"
            TARGET_NAME="$TRUE_DEPLOYMENT"
        fi
    fi

    if [[ "$ACTION_TYPE" == "Patch" ]]; then
        PATCH_TMP="/tmp/kube-patch-${TARGET_NAME}.json"
        
        # Stream, extract, and wrap the payload natively in memory using a single jq call.
        # This guarantees zero shell formatting artifacts break the JSON structure!
        jq --arg container_name "$TARGET_NAME" --argjson step_idx "$i" '
            .remediation[$step_idx].patch_payload as $payload |
            {
                spec: {
                    template: {
                        spec: {
                            containers: [
                                ($payload + {name: $container_name})
                            ]
                        }
                    }
                }
            }
        ' "$PLAN_FILE" > "$PATCH_TMP"
        
        # Enforce that patches hit the deployment spec template instead of raw pods
        EXEC_CMD="kubectl patch ${TARGET_KIND}/${TARGET_NAME} -n kubemind --patch-file ${PATCH_TMP}"
    else
        # For non-patch items, check if we need to fall back to the raw generated command strings
        RAW_CMD=$(jq -r ".remediation[$i].fallback_command // .remediation[$i].command" "$PLAN_FILE")
        
        # Dynamically correct the command string if it contains references to an ephemeral pod/replicaset name
        if [[ "$ORIGINAL_TARGET_NAME" != "$TARGET_NAME" ]]; then
            EXEC_CMD=$(echo "$RAW_CMD" | sed "s/${ORIGINAL_TARGET_NAME}/${TARGET_NAME}/g")
        else
            EXEC_CMD="$RAW_CMD"
        fi
    fi
    # ====================================================================

    echo "---------------------------------------------------"
    echo "🎯 Step Action: $ACTION_TYPE on $TARGET_KIND/$TARGET_NAME"
    echo "⚡ Proposed Execution: $EXEC_CMD"
    echo "---------------------------------------------------"

    if [[ "$RISK" == "Low" && "$AUTO_EXEC" == "true" ]]; then
        echo "🚀 Verification Passed. Executing action autonomously..."
        if eval "$EXEC_CMD"; then
            echo "✅ Action executed successfully."
        else
            echo "❌ Action execution failed."
        fi
    else
        echo "✋ MANUAL APPROVAL REQUIRED: Action risk is $RISK or is not marked for auto-execution."
        echo "⏭️  Action skipped by user choice (Simulation Mode)."
    fi
done