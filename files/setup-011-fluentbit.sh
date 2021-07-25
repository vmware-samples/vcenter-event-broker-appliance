#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup FluentBit

set -euo pipefail

VEBA_BOM_FILE=/root/config/veba-bom.json

# Flutenbit Config files
FLUENTBIT_TEMPLATE=/root/config/fluentbit/templates/fluentbit-ds-template.yaml
FLUENTBIT_CONFIG=/root/config/fluentbit/fluentbit-ds.yaml

# Apply YTT overlay
ytt --data-value-file bom=${VEBA_BOM_FILE} -f ${FLUENTBIT_TEMPLATE} > ${FLUENTBIT_CONFIG}

kubectl apply -f /root/config/fluentbit/fluentbit-preperations.yaml
kubectl apply -f /root/config/fluentbit/fluentbit-configmap.yaml
kubectl apply -f ${FLUENTBIT_CONFIG}
kubectl wait --for=condition=ready pod -l k8s-app=fluent-bit --timeout=3m -n vmware-system