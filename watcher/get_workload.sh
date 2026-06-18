#!/bin/bash

POD_NAME=$1
NAMESPACE=${2:-kubemind}

OWNER_KIND=$(kubectl get pod "$POD_NAME" \
-n "$NAMESPACE" \
-o jsonpath='{.metadata.ownerReferences[0].kind}')

OWNER_NAME=$(kubectl get pod "$POD_NAME" \
-n "$NAMESPACE" \
-o jsonpath='{.metadata.ownerReferences[0].name}')

if [ "$OWNER_KIND" = "ReplicaSet" ]; then

    DEPLOYMENT=$(echo "$OWNER_NAME" | sed 's/-[a-z0-9]\{9,10\}$//')

    echo "$DEPLOYMENT"

else

    echo "$OWNER_NAME"

fi
