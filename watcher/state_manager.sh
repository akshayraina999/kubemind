#!/bin/bash

STATE_DIR="state/processed"

mkdir -p "$STATE_DIR"

is_processed() {
    local WORKLOAD_NAME=$1
    local REASON=$2

    local HASH

    HASH=$(echo "${WORKLOAD_NAME}-${REASON}" | md5sum | awk '{print $1}')

    if [ -f "$STATE_DIR/$HASH" ]
    then
        return 0
    else
        return 1
    fi
}

mark_processed() {
    local WORKLOAD_NAME=$1
    local REASON=$2

    local HASH

    HASH=$(echo "${WORKLOAD_NAME}-${REASON}" | md5sum | awk '{print $1}')

    touch "$STATE_DIR/$HASH"
}