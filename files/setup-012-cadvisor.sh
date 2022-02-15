#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Google cAdvisor

set -euo pipefail

VEBA_BOM_FILE=/root/config/veba-bom.json

# Cadvisor Config files
CADVISOR_TEMPLATE=/root/config/cadvisor/templates/cadvisor-ds-template.yaml
CADVISOR_CONFIG=/root/config/cadvisor/cadvisor-ds.yaml

# Apply YTT overlay
ytt --data-value-file bom=${VEBA_BOM_FILE} -f ${CADVISOR_TEMPLATE} > ${CADVISOR_CONFIG}

echo -e "\e[92mDeploying cAdvisor ..." > /dev/console

kubectl apply -f /root/config/cadvisor/cadvisor-preperations.yaml
kubectl apply -f ${CADVISOR_CONFIG}
kubectl apply -f /root/config/cadvisor/cadvisor-svc.yaml
kubectl wait --for=condition=ready pod -l app=cadvisor --timeout=${KUBECTL_WAIT} -n vmware-system