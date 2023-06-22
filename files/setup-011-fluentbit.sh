#!/bin/bash
# Copyright 2023 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup FluentBit

set -euo pipefail

VEBA_BOM_FILE=/root/config/veba-bom.json
VEBA_CONFIG_FILE=/root/config/veba-config.json

# Flutenbit Config files
FLUENTBIG_CONFIGMAP_OVERLAY=/root/download/fluentbit/overlay.yaml
FLUENTBIT_CONFIGMAP_TEMPLATE=/root/config/fluentbit/templates/fluentbit-configmap-template.yaml
FLUENTBIT_CONFIGMAP_CONFIG=/root/config/fluentbit/fluentbit-configmap.yaml

FLUENTBIT_DEPLOYMENT_TEMPLATE=/root/config/fluentbit/templates/fluentbit-ds-template.yaml
FLUENTBIT_DEPLOYMENT_CONFIG=/root/config/fluentbit/fluentbit-ds.yaml

# Apply YTT overlay
ytt --data-value-file config=${VEBA_CONFIG_FILE} -f ${FLUENTBIG_CONFIGMAP_OVERLAY} -f ${FLUENTBIT_CONFIGMAP_TEMPLATE} > ${FLUENTBIT_CONFIGMAP_CONFIG}
ytt --data-value-file bom=${VEBA_BOM_FILE} -f ${FLUENTBIT_DEPLOYMENT_TEMPLATE} > ${FLUENTBIT_DEPLOYMENT_CONFIG}

kubectl apply -f /root/config/fluentbit/fluentbit-preperations.yaml
kubectl apply -f ${FLUENTBIT_CONFIGMAP_CONFIG}
kubectl apply -f ${FLUENTBIT_DEPLOYMENT_CONFIG}
kubectl wait --for=condition=ready pod -l k8s-app=fluent-bit --timeout=${KUBECTL_WAIT} -n vmware-system