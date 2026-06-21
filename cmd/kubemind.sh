#!/usr/bin/env bash
# cmd/kubemind.sh

MODE=${1:-once}

show_help() {
cat <<EOF
KubeMind - AI Kubernetes Incident Investigator

Usage:
  ./cmd/kubemind.sh once
      Run one scan

  ./cmd/kubemind.sh daemon
      Run continuously every 30 seconds
EOF
}

case "$MODE" in
    once)
        echo "=================================="
        echo "🧠 Starting KubeMind AI Engine"
        echo "📡 Mode: Once"
        echo "=================================="

        # FIX: Call process_failures.sh directly. 
        # It handles running scan_cluster.sh internally via input redirection!
        ./watcher/process_failures.sh
        ;;

    daemon)
        echo "=================================="
        echo "🧠 Starting KubeMind AI Engine"
        echo "📡 Mode: Daemon (Interval: 30s)"
        echo "=================================="

        while true
        do
            ./watcher/process_failures.sh
            echo "💤 Scan cycle complete. Sleeping for 30 seconds..."
            sleep 30
        done
        ;;

    status)
        echo ""
        echo "📋 Processed Anomaly Incidents State"
        COUNT=$(ls state/processed 2>/dev/null | wc -l)
        echo "Total Acted Anomaly Tokens: $COUNT"
        ls -1 state/processed 2>/dev/null
        ;;

    clear-state)
        echo ""
        echo "🧹 Clearing processed incident state tracker data..."
        rm -rf state/processed/*
        echo "✅ State records cleared successfully."
        ;;

    *)
        show_help
        exit 1
        ;;
esac