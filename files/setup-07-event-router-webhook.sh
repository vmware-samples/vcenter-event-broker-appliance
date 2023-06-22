#!/bin/bash
# Copyright 2023 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup VMware Event Router Provider Webhook

set -euo pipefail

VEBA_CONFIG_FILE=/root/config/veba-config.json
VEBA_BOM_FILE=/root/config/veba-bom.json
EVENT_PROVIDER="webhook"

echo -e "\e[92mDeploying VMware Event Router Provider ${EVENT_PROVIDER} ..." > /dev/console
grep -q "Processor:" /etc/veba-release || echo "Processor: Knative" >> /etc/veba-release

# Setup Event Processor Configuration File
EVENT_ROUTER_CONFIG_TEMPLATE=/root/config/event-router/templates/vmware-event-router-config-template.yaml
EVENT_ROUTER_CONFIG=/root/config/event-router/vmware-event-router-config-${EVENT_PROVIDER}.yaml

ytt --data-value eventProvider=${EVENT_PROVIDER} --data-value-file config=${VEBA_CONFIG_FILE} -f ${EVENT_ROUTER_CONFIG_TEMPLATE} > ${EVENT_ROUTER_CONFIG}

kubectl -n vmware-system create secret generic vmware-event-router-config-${EVENT_PROVIDER} --from-file=${EVENT_ROUTER_CONFIG}

# Event Router Config files
EVENT_ROUTER_K8S_TEMPLATE=/root/config/event-router/templates/vmware-event-router-k8s-template.yaml
EVENT_ROUTER_K8S_CONFIG=/root/config/event-router/vmware-event-router-k8s-${EVENT_PROVIDER}.yaml

# Apply YTT overlay
ytt --data-value eventProvider=${EVENT_PROVIDER} --data-value-file bom=${VEBA_BOM_FILE} -f ${EVENT_ROUTER_K8S_TEMPLATE} > ${EVENT_ROUTER_K8S_CONFIG}

kubectl apply -f /root/config/event-router/vmware-event-router-clusterrole.yaml
kubectl -n vmware-system apply -f ${EVENT_ROUTER_K8S_CONFIG}
kubectl wait deployment --all --timeout=${KUBECTL_WAIT} --for=condition=Available -n vmware-system