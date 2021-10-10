#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Event Processor

set -euo pipefail

kubectl -n vmware-system create secret generic basic-auth \
        --from-literal=basic-auth-user=admin \
        --from-literal=basic-auth-password="${ROOT_PASSWORD}"

VEBA_CONFIG_FILE=/root/config/veba-config.json

for EVENT_PROVIDER in ${EVENT_PROVIDERS[@]};
do
  # Setup Event Processor Configuration File
  EVENT_ROUTER_CONFIG_TEMPLATE=/root/config/event-router/templates/vmware-event-router-config-template.yaml
  EVENT_ROUTER_CONFIG=/root/config/event-router/vmware-event-router-config-${EVENT_PROVIDER}.yaml

  if [ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]; then
    echo -e "\e[92mSetting up Knative Processor ..." > /dev/console

    grep -q "Processor:" /etc/veba-release || echo "Processor: Knative" >> /etc/veba-release
  fi

  ytt --data-value eventProvider=${EVENT_PROVIDER} --data-value-file config=${VEBA_CONFIG_FILE} -f ${EVENT_ROUTER_CONFIG_TEMPLATE} > ${EVENT_ROUTER_CONFIG}
done