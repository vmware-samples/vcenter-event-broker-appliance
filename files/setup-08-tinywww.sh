#!/bin/bash
# Copyright 2023 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Deploy TinyWWW Pod

set -euo pipefail

VEBA_BOM_FILE=/root/config/veba-bom.json
VEBA_CONFIG_FILE=/root/config/veba-config.json

# Event Router Config files
TINYWWW_TEMPLATE=/root/config/tinywww/templates/tinywww-template.yaml
TINYWWW_CONFIG=/root/config/tinywww/tinywww.yaml

# Basic Auth for TinyWWW endpoints
kubectl -n vmware-system create secret generic basic-auth \
        --from-literal=basic-auth-user="${ENDPOINT_USERNAME}" \
        --from-literal=basic-auth-password="${ENDPOINT_PASSWORD}"

# Apply YTT overlay
ytt --data-value-file bom=${VEBA_BOM_FILE} --data-value-file config=${VEBA_CONFIG_FILE} -f ${TINYWWW_TEMPLATE} > ${TINYWWW_CONFIG}

kubectl apply -f ${TINYWWW_CONFIG}
