#!/bin/bash

echo "Testing Function ..."
curl -d@test-payload.json \
    -H "Content-Type: application/json" \
    -H 'ce-specversion: 1.0' \
    -H 'ce-id: 166419' \
    -H 'ce-source: https://hz-01.vmware.corp' \
    -H 'ce-type: com.vmware.event.router/horizon.vlsi_userlogin_rest_failed.v0' \
    -H 'ce-time: 2021-09-03T16:00:28Z' \
    -X POST localhost:8080

echo "See docker container console for output"
