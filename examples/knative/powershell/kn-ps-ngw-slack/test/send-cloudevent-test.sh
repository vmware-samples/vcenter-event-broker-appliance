#!/bin/bash

echo "Testing Function ..."
curl -d@test-payload.json \
    -H "Content-Type: application/json" \
    -H 'ce-specversion: 1.0' \
    -H 'ce-id: df8cbd23-01f0-4003-9a6b-d9f16a59f6be' \
    -H 'ce-source: https://vmc.vmware.com/console/sddcs/b8f349e8-48f1-4517-99fe-0bddc753e899' \
    -H 'ce-type: vmware.vmc.SDDC-PROVISION.v0' \
    -X POST localhost:8080

echo "See docker container console for output"
