#!/bin/bash
# Copyright 2023 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup vSphere Sources

set -euo pipefail

echo -e "\e[92mCreating vSphere Secret ..." > /dev/console
kubectl -n vmware-functions create secret generic vsphere-creds --from-literal=username=${VCENTER_USERNAME} --from-literal=password=${VCENTER_PASSWORD}

echo -e "\e[92mCreating vSphere ServiceAccount ..." > /dev/console
kubectl -n vmware-functions create sa vsphere-source-sa

echo -e "\e[92mCreating vSphere Source ..." > /dev/console
# Create vSphere Source
VSPHERE_SOURCE_CONFIG_TEMPLATE=/root/config/vsphere-source/templates/vsphere-source-template.yml
VSPHERE_SOURCE_CONFIG=/root/config/vsphere-source/vsphere-source.yml

ytt --data-value-file config=${VEBA_CONFIG_FILE} -f ${VSPHERE_SOURCE_CONFIG_TEMPLATE} > ${VSPHERE_SOURCE_CONFIG}

kubectl -n vmware-functions create -f ${VSPHERE_SOURCE_CONFIG}
kubectl wait --for=condition=ready vspheresource.sources.tanzu.vmware.com/vsphere-source --timeout=${KUBECTL_WAIT} -n vmware-functions
kubectl wait --for=condition=ready vspherebinding.sources.tanzu.vmware.com/vsphere-source-vspherebinding --timeout=${KUBECTL_WAIT} -n vmware-functions

