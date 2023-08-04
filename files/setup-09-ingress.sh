#!/bin/bash
# Copyright 2023 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Contour / Ingress

set -euo pipefail

# Setup Contour AuthServer
echo -e "\e[92mConfiguring Contour Ingress AuthServer ..." > /dev/console
kubectl create namespace projectcontour-auth

# Contour Auth Config files
INGRESS_AUTHSERVER_TEMPLATE=/root/config/ingress/templates/ingress-authserver-template.yaml
INGRESS_AUTHSERVER_CONFIG=/root/config/ingress/$(basename ${INGRESS_AUTHSERVER_TEMPLATE} | sed 's/-template//g')

VEBA_BOM_FILE=/root/config/veba-bom.json
INGRESS_AUTHSERVER_AUTH_FILE=/root/config/auth

# Apply YTT overlay
ytt --data-value-file bom=${VEBA_BOM_FILE} -f ${INGRESS_AUTHSERVER_TEMPLATE} > ${INGRESS_AUTHSERVER_CONFIG}
kubectl apply -f ${INGRESS_AUTHSERVER_CONFIG}

# Configure Auth file with admin user
htpasswd -b -c ${INGRESS_AUTHSERVER_AUTH_FILE} ${ENDPOINT_USERNAME} ${ENDPOINT_PASSWORD}
kubectl create secret generic -n projectcontour-auth passwords --from-file=${INGRESS_AUTHSERVER_AUTH_FILE}
kubectl annotate secret -n projectcontour-auth passwords projectcontour.io/auth-type=basic

# Create Extension Service
kubectl apply -f /root/config/ingress/ingress-authserver-extensionservice.yaml

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
echo -e "\e[92mDeploying Knative Contour ..." > /dev/console

kubectl create -n contour-external secret tls default-cert --key ${KEY_FILE} --cert ${CERT_FILE}
kubectl apply -f /root/download/contour-delegation.yaml
kubectl patch configmap -n knative-serving config-contour -p '{"data":{"default-tls-secret":"contour-external/default-cert"}}'
kubectl patch configmap -n knative-serving config-domain -p "{\"data\": {\"$CN_NAME\": \"\"}}"

echo -e "\e[92mDeploying Ingress ..." > /dev/console

VEBA_CONFIG_FILE=/root/config/veba-config.json

# Ingress Config files
INGRESS_TEMPLATE=/root/config/ingress/templates/ingressroute-gateway-template.yaml
INGRESS_CONFIG=/root/config/ingress/$(basename ${INGRESS_TEMPLATE} | sed 's/-template//g')

# Apply YTT overlay
ytt --data-value secretName=${CERT_NAME} --data-value-file config=${VEBA_CONFIG_FILE} -f ${INGRESS_TEMPLATE} > ${INGRESS_CONFIG}
kubectl create -f ${INGRESS_CONFIG}
