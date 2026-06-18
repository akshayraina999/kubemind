                     Kubernetes Cluster
                             │
                             ▼
                    KubeMind Watcher
                             │
          ┌──────────────────┼──────────────────┐
          ▼                  ▼                  ▼
     Pod Issues         Events            Pod Logs
          │                  │                  │
          └──────────┬───────┴──────────┬───────┘
                     ▼                  ▼
                  Context Builder
                           │
                           ▼
                    AI RCA Engine
                           │
                 ┌─────────┴─────────┐
                 ▼                   ▼
            RCA Report         Remediation Plan
                 │
                 ▼
             Reports/



kubemind.sh
      │
      ▼
scan_cluster.sh
      │
      ▼
process_failures.sh
      │
      ├── collect_events.sh
      ├── collect_logs.sh
      ├── generate_context.sh
      │
      ▼
analyze.sh
      │
      ▼
generate_report.sh
      │
      ▼
reports/