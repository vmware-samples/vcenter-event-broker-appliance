#!/bin/bash +x

curdir=$(pwd)
cd "$(git rev-parse --show-toplevel)"
docker build -t veba-test -f files/test/dockerfile .
cd "${curdir}"
