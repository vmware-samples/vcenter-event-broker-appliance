#!/bin/bash

### DO NOT EDIT BEYOND HERE ###

echo "Testing Function ..."
curl -i -v -d@test-payload.json \
    -H "Content-Type: application/json" \
    -X POST http://localhost:8080

echo "See docker container console for output"
