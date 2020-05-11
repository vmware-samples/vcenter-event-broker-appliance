#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Contour / Ingress

set -euo pipefail

echo -e "\e[92mDeploying Contour ..." > /dev/console
kubectl --kubeconfig /root/.kube/config create -f /root/download/contour/examples/contour/

## Create SSL Certificate & Secret
KEY_FILE=/root/config/eventrouter.key
CERT_FILE=/root/config/eventrouter.crt
CN_NAME=$(hostname -f)
CERT_NAME=eventrouter-tls

openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ${KEY_FILE} -out ${CERT_FILE} -subj "/CN=${CN_NAME}/O=${CN_NAME}"

kubectl --kubeconfig /root/.kube/config -n vmware create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}

# Deploy Ingress Route

if [ "${EVENT_PROCESSOR_TYPE}" == "AWS EventBridge" ]; then
  cat << EOF > /root/config/ingressroute-gateway.yaml
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  labels:
    app: vmware
  name: event-router
  namespace: vmware
spec:
  virtualhost:
    fqdn: ${HOSTNAME}
    tls:
      secretName: ${CERT_NAME}
      minimumProtocolVersion: "1.2"
  routes:
    - match: /status
      prefixRewrite: /status
      services:
      - name: tinywww
        port: 8100
    - match: /bootstrap
      prefixRewrite: /bootstrap
      services:
      - name: tinywww
        port: 8100
    - match: /stats
      prefixRewrite: /stats
      services:
      - name: vmware-event-router
        port: 8080
EOF
else
  cat << EOF > /root/config/ingressroute-gateway.yaml
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  labels:
    app: vmware
  name: event-router
  namespace: vmware
spec:
  virtualhost:
    fqdn: ${HOSTNAME}
    tls:
      secretName: ${CERT_NAME}
      minimumProtocolVersion: "1.2"
  routes:
    - match: /status
      prefixRewrite: /status
      services:
      - name: tinywww
        port: 8100
    - match: /bootstrap
      prefixRewrite: /bootstrap
      services:
      - name: tinywww
        port: 8100
    - match: /stats
      prefixRewrite: /stats
      services:
      - name: vmware-event-router
        port: 8080
    - match: /
      delegate:
        name: gateway
        namespace: openfaas
---
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  name: gateway
  namespace: openfaas
spec:
  routes:
    - match: /
      services:
      - name: gateway
        port: 8080
EOF
fi

kubectl --kubeconfig /root/.kube/config create -f /root/config/ingressroute-gateway.yaml