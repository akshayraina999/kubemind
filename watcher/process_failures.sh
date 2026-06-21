#!/usr/bin/env bash
set -euo pipefail

# Safely source configuration properties
source ./config/settings.env

# Source your custom state tracking engine
source watcher/state_manager.sh

echo "🔍 Scanning cluster state pipe..."

while read -r RESOURCE_NAME
do
    # Ensure blank strings from loop streaming are dropped safely
    [ -z "$RESOURCE_NAME" ] && continue

    REASON=$(./engine/classifier.sh "$RESOURCE_NAME")

    # Use the state engine tracking deduplication layer
    if is_processed "$RESOURCE_NAME" "$REASON"; then
        echo "⏭️ Skipping already-analyzed target: $RESOURCE_NAME ($REASON)"
        continue
    fi

    echo ""
    echo "=========================================================="
    echo "🧠 KubeMind Processing Anomaly: $RESOURCE_NAME"
    echo "🚨 Classification: $REASON"
    echo "=========================================================="

    echo "⚙️ [1/6] Collecting Context logs..."
    ./engine/collector.sh "$RESOURCE_NAME" "kubemind"

    echo "🧠 [2/6] Running AI Analysis Inference Pipeline..."
    ./ai/analyze.sh "$RESOURCE_NAME"

    # Define the generated plan layout file pathway explicitly
    PLAN_FILE="./logs/remediation-plans/${RESOURCE_NAME}.json"

    echo "📋 Plan Context Summary:"
    if [ -f "$PLAN_FILE" ]; then
        cat "$PLAN_FILE"
    else
        echo "❌ Critical Error: Plan specification file missing."
        continue
    fi

    echo "⚙️ [3/6] Building & Executing Remediation Strategy..."
    ./engine/remediator.sh "$PLAN_FILE"

    echo "🛡️ [4/6] Verifying Workload Status Patch Stability..."
    ./engine/verifier.sh "$PLAN_FILE"

    # ====================================================================
    # 📊 [5/6] Generating Incident Analysis Report Markdown Summary
    # ====================================================================
    echo "📊 [5/6] Generating Incident Analysis Report Markdown Summary..."
    
    # Dynamically assemble the two exact positional parameters needed by the reporter
    TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
    REPORT_FILE="./reports/${TIMESTAMP}_${RESOURCE_NAME}.md"
    
    ./engine/reporter.sh "$PLAN_FILE" "$REPORT_FILE"
    # ====================================================================

    echo "💾 [6/6] Writing state token metadata down to persistent ledger..."
    mark_processed "$RESOURCE_NAME" "$REASON"

# Feed the process loop directly from the stdout stream of the scanner script
done < <(./watcher/scan_cluster.sh)