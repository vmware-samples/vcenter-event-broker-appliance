#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup VMware Event Router

set -euo pipefail

for EVENT_PROVIDER in ${EVENT_PROVIDERS[@]};
do
    echo -e "\e[92mDeploying VMware Event Router for ${EVENT_PROVIDER} ..." > /dev/console
    EVENT_ROUTER_CONFIG=/root/config/event-router/vmware-event-router-config-${EVENT_PROVIDER}.yaml

    kubectl -n vmware-system create secret generic vmware-event-router-config-${EVENT_PROVIDER} --from-file=${EVENT_ROUTER_CONFIG}

    VEBA_BOM_FILE=/root/config/veba-bom.json

    # Event Router Config files
    EVENT_ROUTER_K8S_TEMPLATE=/root/config/event-router/templates/vmware-event-router-k8s-template.yaml
    EVENT_ROUTER_K8S_CONFIG=/root/config/event-router/vmware-event-router-k8s-${EVENT_PROVIDER}.yaml

    # Apply YTT overlay
    ytt --data-value eventProvider=${EVENT_PROVIDER} --data-value-file bom=${VEBA_BOM_FILE} -f ${EVENT_ROUTER_K8S_TEMPLATE} > ${EVENT_ROUTER_K8S_CONFIG}

    kubectl apply -f /root/config/event-router/vmware-event-router-clusterrole.yaml
    kubectl -n vmware-system apply -f ${EVENT_ROUTER_K8S_CONFIG}
    kubectl wait deployment --all --timeout=3m --for=condition=Available -n vmware-system
done