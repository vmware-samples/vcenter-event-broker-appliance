#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
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
  INGRESS_CONFIG_YAML=/root/config/openfaas-ingressroute-gateway.yaml
# Ingress Route Configuration for AWS EventBridge
elif [ "${EVENT_PROCESSOR_TYPE}" == "AWS EventBridge" ]; then
  INGRESS_CONFIG_YAML=/root/config/eventbridge-ingressroute-gateway.yaml
# Ingress Route Configuration for Knative External
elif [[ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]] && [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "external" ]]; then
  INGRESS_CONFIG_YAML=/root/config/knative-external-ingressroute-gateway.yaml
# Ingress Route Configuration for Knative Embedded w/VEBA UI
elif [[ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]] && [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "embedded" ]] && [[ ! -z ${VCENTER_USERNAME_FOR_VEBA_UI} ]] && [[ ! -z ${VCENTER_PASSWORD_FOR_VEBA_UI} ]]; then
  INGRESS_CONFIG_YAML=/root/config/knative-embedded-veba-ui-ingressroute-gateway.yaml
# Ingress Route Configuration for Knative Embedded w/o VEBA UI
elif [[ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]] && [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "embedded" ]]; then
  INGRESS_CONFIG_YAML=/root/config/knative-embedded-ingressroute-gateway.yaml
fi

if [ ! -z ${INGRESS_CONFIG_YAML} ]; then
  echo -e "\e[92mDeploying Ingress using configuration ${INGRESS_CONFIG_YAML} ..." > /dev/console
  sed -i "s/##HOSTNAME##/${HOSTNAME}/g" ${INGRESS_CONFIG_YAML}
  sed -i "s/##CERT_NAME##/${CERT_NAME}/g" ${INGRESS_CONFIG_YAML}
  kubectl create -f ${INGRESS_CONFIG_YAML}
else
  echo -e "\e[91mUnable to match a supported Ingress configuration ..." > /dev/console
  exit 1
fi