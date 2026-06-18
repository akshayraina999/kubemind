#!/bin/bash

NAMESPACE="kubemind"

kubectl get pods -n "$NAMESPACE" -o json |
jq -r '
.items[]
|
{
  pod: .metadata.name,
  phase: .status.phase,
  reason: (
      .status.containerStatuses[0].state.waiting.reason
      //
      .status.containerStatuses[0].state.terminated.reason
      //
      "Running"
  )
}
|
select(
      .reason == "ImagePullBackOff"
   or .reason == "ErrImagePull"
   or .reason == "CrashLoopBackOff"
   or .reason == "OOMKilled"
   or .phase == "Failed"
   or .phase == "Pending"
)
|
"\(.pod)|\(.reason)"
'