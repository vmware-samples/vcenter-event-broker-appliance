#!/bin/bash

echo "Testing Function ..."
curl -d@test-payload.json \
    -H "Content-Type: application/json" \
    -H 'ce-specversion: 1.0' \
    -H 'ce-id: d70079f9-fddd-4b7f-aa76-1193f28b0611' \
    -H 'ce-source: https://vcenter.local/sdk' \
    -H 'ce-type: com.vmware.event.router/event' \
    -H 'ce-subject: TaskEvent' \
    -X POST localhost:8080

echo "See docker container console for output"
