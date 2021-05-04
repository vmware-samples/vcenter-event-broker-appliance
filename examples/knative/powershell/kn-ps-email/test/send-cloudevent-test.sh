#!/bin/bash

echo "Testing Function ..."
curl -d@test-payload.json -H "Content-Type: application/json" -H 'ce-specversion: 1.0' -H 'ce-id: id-123' -H 'ce-source: source-123' -H 'ce-type: com.vmware.event.router/event' -H 'ce-subject: VmRemovedEvent' -X POST localhost:8080

echo "See docker container console for output"
