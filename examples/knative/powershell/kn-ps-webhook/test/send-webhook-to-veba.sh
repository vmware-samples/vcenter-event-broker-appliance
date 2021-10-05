#!/bin/bash

WEBHOOK_USERNAME='FILL_IN_WEBHOOK_USERNAME'
WEBHOOK_PASS='FILL_IN_WEBHOOK_PASSWORD'
WEBHOOK_FUNCTION_URL='FILL_IN_FUNCTION_SERVICE_URL' # https://kn-ps-webhook.vmware-functions.veba.primp-industries.local

### DO NOT EDIT BEYOND HERE ###

WEBHOOK_BASE64=$(echo -n "${WEBHOOK_USERNAME}:${WEBHOOK_PASS}" | openssl base64)

echo "Testing Function ..."
curl -i -k -d@test-payload.json \
    -H "Content-Type: application/json" \
    -H "Authorization: Basic ${WEBHOOK_BASE64}" \
    -X POST ${WEBHOOK_FUNCTION_URL}
