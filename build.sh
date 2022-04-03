#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail

VEBA_BOM_FILE=veba-bom.json

if [ ! -e ${VEBA_BOM_FILE} ]; then
    echo "Unable to locate veba-bom.json in current directory which is required"
    exit 1
fi

if ! hash jq 2>/dev/null; then
    echo "jq utility is not installed on this system"
    exit 1
fi

if [[ ! -z $(git status -s | grep -vE 'photon-builder.json|test/.*\.sh') ]]; then
    echo "Dirty Git repository, please clean up any untracked files or commit them before building"
    exit 1
fi

rm -f output-vmware-iso/*.ova

VEBA_VERSION_FROM_BOM=$(jq -r < ${VEBA_BOM_FILE} '.veba.version')
VEBA_COMMIT=$(git rev-parse --short HEAD)

echo "Building VEBA OVA from ${VEBA_VERSION_FROM_BOM} ..."
packer build -var "VEBA_VERSION=${VEBA_VERSION_FROM_BOM}" -var "VEBA_COMMIT=${VEBA_COMMIT}" -var-file=photon-builder.json -var-file=photon-version.json photon.json

