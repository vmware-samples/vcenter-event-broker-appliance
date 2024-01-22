#!/bin/bash

echo "Testing Function ..."
curl -d@test-payload.json \
    -H "Content-Type: application/json" \
    -H 'ce-specversion: 1.0' \
    -H 'ce-id: 2112913' \
    -H 'ce-source: https://vcenter.primp-industries.local/sdk' \
    -H 'ce-type: com.vmware.vsphere.com.vmware.applmgmt.backup.job.failed.event.v0' \
    -X POST localhost:8080

echo "See docker container console for output"
