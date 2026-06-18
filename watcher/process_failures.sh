#!/bin/bash
set -euo pipefail

source watcher/state_manager.sh

while IFS="|" read -r POD_NAME REASON
do

    WORKLOAD=$(./watcher/get_workload.sh "$POD_NAME")

    if is_processed "$WORKLOAD" "$REASON"
    then
        echo "Skipping $WORKLOAD ($REASON)"
        continue
    fi

    echo ""
    echo "=================================="
    echo "Processing: $POD_NAME"
    echo "Workload: $WORKLOAD"
    echo "Reason: $REASON"
    echo "=================================="

    echo "[1/4] Collecting Failure Context..."
    ./collector/dispatch_handler.sh "$POD_NAME" "$REASON"

    echo "[2/4] Generating Context..."
    ./collector/generate_context.sh "$POD_NAME"

    echo "[3/4] Running AI Analysis..."
    ./ai/analyze.sh "$POD_NAME"

    echo "[4/4] Generating Report..."
    REPORT=$(./reporter/generate_report.sh "$POD_NAME")

    echo "Report Created:"
    echo "$REPORT"

    mark_processed "$WORKLOAD" "$REASON"

done < <(./watcher/scan_cluster.sh)