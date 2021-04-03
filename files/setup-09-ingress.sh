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

## Create SSL Certificate & Secret
KEY_FILE=/root/config/eventrouter.key
CERT_FILE=/root/config/eventrouter.crt
CN_NAME=$(hostname -f)
CERT_NAME=eventrouter-tls

openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ${KEY_FILE} -out ${CERT_FILE} -subj "/CN=${CN_NAME}/O=${CN_NAME}"

kubectl -n vmware-system create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}

# Knative Contour for Knative Embedded Broker
if [ "${KNATIVE_DEPLOYMENT_TYPE}" == "embedded" ]; then
  echo -e "\e[92mDeploying Knative Contour ..." > /dev/console

  kubectl create -n contour-external secret tls default-cert --key ${KEY_FILE} --cert ${CERT_FILE}
  kubectl apply -f /root/download/contour-delegation.yaml
  kubectl patch configmap -n knative-serving config-contour -p '{"data":{"default-tls-secret":"contour-external/default-cert"}}'
  kubectl patch configmap -n knative-serving config-domain -p "{\"data\": {\"$CN_NAME\": \"\"}}"
fi

# Deploy Ingress Route

if [ "${EVENT_PROCESSOR_TYPE}" == "OpenFaaS" ]; then
  cat << EOF > /root/config/ingressroute-gateway.yaml
---

apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  labels:
    app: vmware
  name: event-router
  namespace: vmware-system
spec:
  includes:
  - conditions:
    - prefix: /
    name: gateway
    namespace: openfaas
  routes:
  - conditions:
    - prefix: /status
    pathRewritePolicy:
      replacePrefix:
      - replacement: /status
    services:
    - name: tinywww
      port: 8100
  - conditions:
    - prefix: /bootstrap
    pathRewritePolicy:
      replacePrefix:
      - replacement: /bootstrap
    services:
    - name: tinywww
      port: 8100
  - conditions:
    - prefix: /stats
    pathRewritePolicy:
      replacePrefix:
      - replacement: /stats
    services:
    - name: vmware-event-router
      port: 8082
  virtualhost:
    fqdn: ${HOSTNAME}
    tls:
      minimumProtocolVersion: "1.2"
      secretName: ${CERT_NAME}
status: {}
---

apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: gateway
  namespace: openfaas
spec:
  routes:
  - conditions:
    - prefix: /
    services:
    - name: gateway
      port: 8080
status: {}
EOF
elif [[ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]] && [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "embedded" ]]; then
  cat << EOF > /root/config/ingressroute-gateway.yaml
---

apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  annotations:
    kubernetes.io/ingress.class: contour-external
  labels:
    app: vmware
  name: event-router
  namespace: vmware-system
spec:
  routes:
  - conditions:
    - prefix: /status
    pathRewritePolicy:
      replacePrefix:
      - replacement: /status
    services:
    - name: tinywww
      port: 8100
  - conditions:
    - prefix: /bootstrap
    pathRewritePolicy:
      replacePrefix:
      - replacement: /bootstrap
    services:
    - name: tinywww
      port: 8100
  - conditions:
    - prefix: /stats
    pathRewritePolicy:
      replacePrefix:
      - replacement: /stats
    services:
    - name: vmware-event-router
      port: 8082
  - conditions:
    - prefix: /veba-ui
    services:
    - name: veba-ui
      port: 80
  virtualhost:
    fqdn: ${HOSTNAME}
    tls:
      minimumProtocolVersion: "1.2"
      secretName: ${CERT_NAME}
status: {}
EOF
else
  cat << EOF > /root/config/ingressroute-gateway.yaml
---

apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  labels:
    app: vmware
  name: event-router
  namespace: vmware-system
spec:
  routes:
  - conditions:
    - prefix: /status
    pathRewritePolicy:
      replacePrefix:
      - replacement: /status
    services:
    - name: tinywww
      port: 8100
  - conditions:
    - prefix: /bootstrap
    pathRewritePolicy:
      replacePrefix:
      - replacement: /bootstrap
    services:
    - name: tinywww
      port: 8100
  - conditions:
    - prefix: /stats
    pathRewritePolicy:
      replacePrefix:
      - replacement: /stats
    services:
    - name: vmware-event-router
      port: 8082
  virtualhost:
    fqdn: ${HOSTNAME}
    tls:
      minimumProtocolVersion: "1.2"
      secretName: ${CERT_NAME}
status: {}
EOF
fi

kubectl create -f /root/config/ingressroute-gateway.yaml