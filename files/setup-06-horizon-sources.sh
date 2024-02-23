#!/bin/bash
# Copyright 2023 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Horizon Sources

set -euo pipefail

# Create Horizon Secret
echo -e "\e[92mCreating Horizon Secret ..." > /dev/console
kubectl -n vmware-functions create secret generic horizon-creds --from-literal=domain=${HORIZON_DOMAIN} --from-literal=username=${HORIZON_USERNAME} --from-literal=password=${HORIZON_PASSWORD}

# Create vSphere Source
echo -e "\e[92mCreating Horizon Source ..." > /dev/console

echo -e "\e[92mCreating Horizon ServiceAccount ..." > /dev/console
kubectl -n vmware-functions create sa horizon-source-sa

HORIZON_SOURCE_CONFIG_TEMPLATE=/root/config/horizon-source/templates/horizon-source-template.yml
HORIZON_SOURCE_CONFIG=/root/config/horizon-source/horizon-source.yml

ytt --data-value-file config=${VEBA_CONFIG_FILE} -f ${HORIZON_SOURCE_CONFIG_TEMPLATE} > ${HORIZON_SOURCE_CONFIG}

kubectl -n vmware-functions create -f ${HORIZON_SOURCE_CONFIG}
kubectl wait --for=condition=Available deploy/horizon-source-adapter --timeout=${KUBECTL_WAIT} -n vmware-functions
