#!/usr/bin/env bash
set -euo pipefail

PLAN_FILE=${1:-""}
REPORT_FILE=${2:-""}

if [[ -z "$PLAN_FILE" || ! -f "$PLAN_FILE" || -z "$REPORT_FILE" ]]; then
    echo "❌ Error: Missing validation arguments inside reporter context."
    exit 1
fi

# 1. Extract Metadata cleanly using your schema mapping keys
RESOURCE_KIND=$(jq -r '.resource_type // "Unknown Kind"' "$PLAN_FILE")
RESOURCE_NAME=$(jq -r '.resource_name // "Unknown Name"' "$PLAN_FILE")
ISSUE_TYPE=$(jq -r '.issue_type // "Unhandled Issue"' "$PLAN_FILE")
ROOT_CAUSE=$(jq -r '.root_cause // "No analysis provided by AI Engine."' "$PLAN_FILE")

{
    echo "# 🧠 KubeMind Incident Analysis Report"
    echo ""
    echo "**Generated:** $(date)"
    echo "**Target Resource:** \`${RESOURCE_KIND}/${RESOURCE_NAME}\`"
    echo "**Classification Label:** \`${ISSUE_TYPE}\`"
    echo ""
    echo "---"
    echo ""
    echo "## 🔍 Deep Root Cause Analysis"
    echo "$ROOT_CAUSE"
    echo ""
    echo "## ⚡ Proposed Remediation Strategy Tasks"
    echo ""
    
    # 2. Iterate dynamically over the remediation matrix keys
    STEPS_COUNT=$(jq '.remediation | length' "$PLAN_FILE")
    if [ "$STEPS_COUNT" -eq 0 ]; then
        echo "_No automatic remediation steps proposed for this anomaly workload._"
    else
        for ((i=0; i<STEPS_COUNT; i++)); do
            ACTION=$(jq -r ".remediation[$i].action_type" "$PLAN_FILE")
            KIND=$(jq -r ".remediation[$i].target_kind" "$PLAN_FILE")
            NAME=$(jq -r ".remediation[$i].target_name" "$PLAN_FILE")
            COMMAND=$(jq -r ".remediation[$i].command" "$PLAN_FILE")
            RISK=$(jq -r ".remediation[$i].risk" "$PLAN_FILE")
            AUTO=$(jq -r ".remediation[$i].auto_executable" "$PLAN_FILE")
            
            echo "### Step $((i+1)): $ACTION on $KIND/$NAME"
            echo "* **Risk Level:** $RISK"
            echo "* **Autonomous Execution Allowed:** $AUTO"
            echo "* **Proposed Execution Command:**"
            echo "  \`\`\`bash"
            echo "  $COMMAND"
            echo "  \`\`\`"
            echo ""
        done
    fi
    
    echo "## 📋 Manual Verification Tasks"
    jq -r '.verification[] | "- " + .' "$PLAN_FILE" 2>/dev/null || echo "_No explicit manual verification tasks mapped._"

} > "$REPORT_FILE"

echo "Report successfully generated: $REPORT_FILE"