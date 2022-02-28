#!/bin/bash

# The ce-subject value should match the event router subject in function.yaml
echo "Testing Function ..."
PAYLOAD_PATH="test-payload.json"
if [ $# -gt 0 ]; then
    if test -f "$1"; then
        PAYLOAD_PATH=$1
    else
        echo "$1 not found"
        exit 1
    fi
fi
curl -d@$PAYLOAD_PATH \
    -H "Content-Type: application/json" \
    -H 'ce-specversion: 1.0' \
    -H 'ce-id: d70079f9-fddd-4b7f-aa76-1193f28b0611' \
    -H 'ce-source: https://vcenter.local/sdk' \
    -H 'ce-type: com.vmware.event.router/event' \
    -H 'ce-subject: VmReconfiguredEvent' \
    -X POST localhost:8080

echo "See docker container console for output"
