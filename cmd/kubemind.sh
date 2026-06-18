#!/bin/bash

echo ""
echo "=================================="
echo "Starting KubeMind"
echo "=================================="

while true
do

    ./watcher/process_failures.sh

    sleep 30

done