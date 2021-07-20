#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Event Processor

set -euo pipefail

kubectl -n vmware-system create secret generic basic-auth \
        --from-literal=basic-auth-user=admin \
        --from-literal=basic-auth-password="${ROOT_PASSWORD}"

VEBA_CONFIG_FILE=/root/config/veba-config.json

# Setup Event Processor Configuration File
EVENT_ROUTER_CONFIG_TEMPLATE=/root/config/event-router/templates/event-router-config-template.yaml
EVENT_ROUTER_CONFIG=/root/config/event-router/event-router-config.yaml

if [ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]; then
  echo -e "\e[92mSetting up Knative Processor ..." > /dev/console

  echo "Processor: Knative" >> /etc/veba-release
elif [ "${EVENT_PROCESSOR_TYPE}" == "AWS EventBridge" ]; then
  echo -e "\e[92mSetting up AWS Event Bridge Processor ..." > /dev/console

  echo "Processor: EventBridge" >> /etc/veba-release
else
  # Setup OpenFaaS
  echo -e "\e[92mSetting up OpenFaas Processor ..." > /dev/console
  kubectl create -f /root/download/faas-netes/namespaces.yml

  # Setup OpenFaaS Secret
  kubectl -n openfaas create secret generic basic-auth \
      --from-literal=basic-auth-user=admin \
      --from-literal=basic-auth-password="${OPENFAAS_PASSWORD}"

  kubectl apply -f /root/download/faas-netes/yaml

  echo "Processor: OpenFaaS" >> /etc/veba-release
fi

ytt --data-value-file config=${VEBA_CONFIG_FILE} -f ${EVENT_ROUTER_CONFIG_TEMPLATE} > ${EVENT_ROUTER_CONFIG}