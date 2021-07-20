#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Contour / Ingress

set -euo pipefail

# Standard Contour for OpenFaaS and Knative w/External Broker
if [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "na" ]] || [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "external" ]]; then
  echo -e "\e[92mDeploying Contour ..." > /dev/console
  kubectl create -f /root/download/contour/examples/contour/
fi

KEY_FILE=/root/config/eventrouter.key
CERT_FILE=/root/config/eventrouter.crt
CERT_NAME=eventrouter-tls
CN_NAME=$(hostname -f)

# Customer provided TLS Certificate
if [[ ! -z ${CUSTOM_VEBA_TLS_PRIVATE_KEY} ]] && [[ ! -z ${CUSTOM_VEBA_TLS_CA_CERT} ]]; then
  echo ${CUSTOM_VEBA_TLS_PRIVATE_KEY} | /usr/bin/base64 -d > ${KEY_FILE}
  echo ${CUSTOM_VEBA_TLS_CA_CERT} | /usr/bin/base64 -d > ${CERT_FILE}
else
  # Create Self Sign TLS Certifcate
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ${KEY_FILE} -out ${CERT_FILE} -subj "/CN=${CN_NAME}/O=${CN_NAME}"
fi

kubectl -n vmware-system create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}

# Knative Contour for Knative Embedded Broker
if [ "${KNATIVE_DEPLOYMENT_TYPE}" == "embedded" ]; then
  echo -e "\e[92mDeploying Knative Contour ..." > /dev/console

  kubectl create -n contour-external secret tls default-cert --key ${KEY_FILE} --cert ${CERT_FILE}
  kubectl apply -f /root/download/contour-delegation.yaml
  kubectl patch configmap -n knative-serving config-contour -p '{"data":{"default-tls-secret":"contour-external/default-cert"}}'
  kubectl patch configmap -n knative-serving config-domain -p "{\"data\": {\"$CN_NAME\": \"\"}}"
fi

# Ingress Route Configuration for OpenFaaS
if [ "${EVENT_PROCESSOR_TYPE}" == "OpenFaaS" ]; then
  INGRESS_TEMPLATE=/root/config/ingress/templates/openfaas-ingressroute-gateway-template.yaml
# Ingress Route Configuration for AWS EventBridge
elif [ "${EVENT_PROCESSOR_TYPE}" == "AWS EventBridge" ]; then
  INGRESS_TEMPLATE=/root/config/ingress/templates/eventbridge-ingressroute-gateway-template.yaml
# Ingress Route Configuration for Knative External
elif [[ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]] && [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "external" ]]; then
  INGRESS_TEMPLATE=/root/config/ingress/templates/knative-external-ingressroute-gateway-template.yaml
# Ingress Route Configuration for Knative Embedded w/VEBA UI
elif [[ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]] && [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "embedded" ]] && [[ ! -z ${VCENTER_USERNAME_FOR_VEBA_UI} ]] && [[ ! -z ${VCENTER_PASSWORD_FOR_VEBA_UI} ]]; then
  INGRESS_TEMPLATE=/root/config/ingress/templates/knative-embedded-veba-ui-ingressroute-gateway-template.yaml
# Ingress Route Configuration for Knative Embedded w/o VEBA UI
elif [[ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]] && [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "embedded" ]]; then
  INGRESS_TEMPLATE=/root/config/ingress/templates/knative-embedded-ingressroute-gateway-template.yaml
fi

if [ ! -z ${INGRESS_TEMPLATE} ]; then
  echo -e "\e[92mDeploying Ingress using configuration ${INGRESS_TEMPLATE} ..." > /dev/console

  VEBA_CONFIG_FILE=/root/config/veba-config.json

  # Ingress Config files
  INGRESS_OVERLAY=/root/config/ingress/overlay.yaml
  INGRESS_CONFIG=/root/config/ingress/$(basename ${INGRESS_TEMPLATE} | sed 's/-template//g')

  # Apply YTT overlay
  ytt --data-value secretName=${CERT_NAME} --data-value-file config=${VEBA_CONFIG_FILE} -f ${INGRESS_OVERLAY} -f ${INGRESS_TEMPLATE} > ${INGRESS_CONFIG}
  kubectl create -f ${INGRESS_CONFIG}
else
  echo -e "\e[91mUnable to match a supported Ingress configuration ..." > /dev/console
  exit 1
fi