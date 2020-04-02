#!/bin/bash -x
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

echo "Building OVA ..."
rm -f output-vmware-iso/*.ova

if [[ ! -z $(git status -s) ]]; then
    echo "Dirty Git repository, please clean up any untracked files or commit them before building"
    exit
fi

echo "Applying packer build to photon.json ..."
packer build -var "VEBA_VERSION=$(cat VERSION)" -var "VEBA_COMMIT=$(git rev-parse --short HEAD)" -var-file=photon-builder.json -var-file=photon-version.json photon.json
