#!/bin/bash -x
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

echo "Building OVA ..."
rm -f output-vmware-iso/*.ova

if [[ "$1" -gt "-1" ]] && [[ $1 == "dev" ]]; then
    echo "Applying packer build to photon-dev.json ..."
    packer build -var-file=photon-builder.json -var-file=photon-version.json photon-dev.json
else
    echo "Applying packer build to photon.json ..."
    packer build -var-file=photon-builder.json -var-file=photon-version.json photon.json
fi
