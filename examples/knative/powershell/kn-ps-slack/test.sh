#!/bin/bash

echo "Testing Function ..."
curl -d@binary-payload -H "Content-Type: application/json" -H 'ce-specversion: 1.0' -H 'ce-id: id-123' -H 'ce-source: source-123' -H 'ce-type: binary' -H 'ce-subject: subject-123' -X POST localhost:8080

echo "See docker container console for output"
