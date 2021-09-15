#!/bin/bash

echo "Testing Function ..."
curl -i -vvv -d@test-payload.json \
    -H "Content-Type: application/json" \
    -H 'ce-specversion: 1.0' \
    -H 'ce-id: 41289fef-0727-46f7-b1a9-b8145972c734' \
    -H 'ce-source: https://vcenter.local/sdk' \
    -H 'ce-type: com.vmware.event.router/event' \
    -H 'ce-subject: VmMigratedEvent' \
    -X POST localhost:8080

echo "See docker container console for output"
