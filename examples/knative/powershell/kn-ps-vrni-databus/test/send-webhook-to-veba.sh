#!/bin/bash

WEBHOOK_FUNCTION_URL='FILL_IN_FUNCTION_SERVICE_URL' # https://kn-ps-vrni-databus.vmware-functions.veba.vrni.cmbu.local

### DO NOT EDIT BEYOND HERE ###

echo "Testing Function ..."
curl -i -k -d@test-payload.json \
    -H "Content-Type: application/json" \
    -X POST ${WEBHOOK_FUNCTION_URL}
