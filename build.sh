#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail

VEBA_BOM_FILE=veba-bom.json

if [ ! -e ${VEBA_BOM_FILE} ]; then
    echo "Unable to locate veba-bom.json in current directory which is required"
    exit 1
fi

#command -v jqw > /dev/null 2>&1
if [ ! -e /usr/local/bin/jq ]; then
    echo "jq utility is not installed on this system"
    exit 1
fi

if [ $# -ne 1 ]; then
    echo -e "\n\tUsage: $0 [master|release]\n"
    exit 1
fi

if [[ ! -z $(git status -s) ]]; then
    echo "Dirty Git repository, please clean up any untracked files or commit them before building"
    exit
fi

rm -f output-vmware-iso/*.ova

VEBA_VERSION_FROM_BOM=$(jq -r < ${VEBA_BOM_FILE} '.veba.version')

if [ "$1" == "release" ]; then
    echo "Building VEBA OVA release ..."
    packer build -var "VEBA_VERSION=${VEBA_VERSION_FROM_BOM}-release" -var "VEBA_COMMIT=$(git rev-parse --short HEAD)" -var-file=photon-builder.json -var-file=photon-version.json photon.json
elif [ "$1" == "master" ]; then
    echo "Building VEBA OVA master ..."
    packer build -var "VEBA_VERSION=${VEBA_VERSION_FROM_BOM}" -var "VEBA_COMMIT=$(git rev-parse --short HEAD)" -var-file=photon-builder.json -var-file=photon-version.json photon.json
else
    echo -e "\nPlease specify release or master to build ...\n"
fi