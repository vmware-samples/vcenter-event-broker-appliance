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

echo -e "\e[92mDeploying Ingress ..." > /dev/console

VEBA_CONFIG_FILE=/root/config/veba-config.json

# Ingress Config files
INGRESS_TEMPLATE=/root/config/ingress/templates/ingressroute-gateway-template.yaml
INGRESS_CONFIG=/root/config/ingress/$(basename ${INGRESS_TEMPLATE} | sed 's/-template//g')

# Apply YTT overlay
ytt --data-value secretName=${CERT_NAME} --data-value-file config=${VEBA_CONFIG_FILE} -f ${INGRESS_TEMPLATE} > ${INGRESS_CONFIG}
kubectl create -f ${INGRESS_CONFIG}
