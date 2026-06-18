#!/bin/bash

POD_NAME=$1
REASON=$2

case "$REASON" in

    ImagePullBackOff|ErrImagePull)
        ./collector/handlers/imagepullbackoff.sh "$POD_NAME"
        ;;

    CrashLoopBackOff)
        ./collector/handlers/crashloopbackoff.sh "$POD_NAME"
        ;;

    OOMKilled)
        ./collector/handlers/oomkilled.sh "$POD_NAME"
        ;;

    Pending)
        ./collector/handlers/pending.sh "$POD_NAME"
        ;;

    *)
        ./collector/handlers/generic.sh "$POD_NAME"
        ;;
esac