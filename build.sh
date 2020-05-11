#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail


if [ $# -ne 1 ]; then
    echo -e "\n\tUsage: $0 [master|release]\n"
    exit 1
fi

if [[ ! -z $(git status -s) ]]; then
    echo "Dirty Git repository, please clean up any untracked files or commit them before building"
    exit
fi

rm -f output-vmware-iso/*.ova

if [ "$1" == "release" ]; then
    echo "Building VEBA OVA release ..."
    packer build -var "VEBA_VERSION=$(cat VERSION)-release" -var "VEBA_COMMIT=$(git rev-parse --short HEAD)" -var-file=photon-builder.json -var-file=photon-version.json photon.json
elif [ "$1" == "master" ]; then
    echo "Building VEBA OVA master ..."
    packer build -var "VEBA_VERSION=$(cat VERSION)" -var "VEBA_COMMIT=$(git rev-parse --short HEAD)" -var-file=photon-builder.json -var-file=photon-version.json photon.json
else
    echo -e "\nPlease specify release or master to build ...\n"
fi