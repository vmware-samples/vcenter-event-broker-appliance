#!/bin/bash

WEBHOOK_USERNAME='FILL_IN_WEBHOOK_USERNAME'
WEBHOOK_PASSWORD='FILL_IN_WEBHOOK_PASSWORD'

### DO NOT EDIT BEYOND HERE ###

WEBHOOK_BASE64=$(echo -n "${WEBHOOK_USERNAME}:${WEBHOOK_PASSWORD}" | openssl base64)

echo "Testing Function ..."
curl -i -v -d@test-payload.json \
    -H "Content-Type: application/json" \
    -H "Authorization: Basic ${WEBHOOK_BASE64}" \
    -X POST localhost:8080

echo "See docker container console for output"
