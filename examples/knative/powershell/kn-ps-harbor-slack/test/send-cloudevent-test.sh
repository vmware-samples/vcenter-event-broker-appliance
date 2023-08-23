#!/bin/bash

echo "Testing Function ..."
curl -d@test-payload.json \
    -H "Content-Type: application/json" \
    -H 'ce-specversion: 1.0' \
    -H 'ce-id: 291ee129-1d27-415c-bbe1-3ca45d5f230a' \
    -H 'ce-source: /projects/2/webhook/policies/1' \
    -H 'ce-type: harbor.artifact.pushed' \
    -H 'ce-operator: admin' \
    -H 'ce-time: 2023-08-22T15:57:41Z' \
    -X POST localhost:8080

echo "See docker container console for output"
