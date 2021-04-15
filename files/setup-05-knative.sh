#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Knative

set -euo pipefail

echo -e "\e[92mDeploying Knative Serving ..." > /dev/console
kubectl apply -f /root/download/serving-crds.yaml
kubectl apply -f /root/download/serving-core.yaml
kubectl wait deployment --all --timeout=-1s --for=condition=Available -n knative-serving
kubectl apply -f /root/download/knative-contour.yaml
kubectl apply -f /root/download/net-contour.yaml
kubectl patch configmap/config-network --namespace knative-serving --type merge --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
kubectl wait deployment --all --timeout=-1s --for=condition=Available -n contour-external
kubectl wait deployment --all --timeout=-1s --for=condition=Available -n contour-internal

# Github Issue #332
kubectl -n knative-serving get cm config-deployment -o json | jq '.data.registriesSkippingTagResolving="projects.registry.vmware.com"' > /root/config/skip-tag-patch.json
kubectl -n knative-serving patch cm config-deployment --type=merge --patch-file /root/config/skip-tag-patch.json

echo -e "\e[92mDeploying Knative Eventing ..." > /dev/console
kubectl apply -f /root/download/eventing-crds.yaml
kubectl apply -f /root/download/eventing-core.yaml
kubectl wait pod --timeout=-1s --for=condition=Ready -l '!job-name' -n knative-eventing

echo -e "\e[92mDeploying RabbitMQ Cluster Operator ..." > /dev/console
kubectl apply -f /root/download/cluster-operator.yml

echo -e "\e[92mDeploying RabbitMQ Broker ..." > /dev/console
kubectl apply -f /root/download/rabbitmq-broker.yaml

cat > /root/config/rabbit.yaml << EOF
apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: veba-rabbit
  namespace: vmware-system
spec:
  resources:
    requests:
      memory: 200Mi
      cpu: 100m
  replicas: 1
---
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
  name: default
  namespace: vmware-functions
  annotations:
    eventing.knative.dev/broker.class: RabbitMQBroker
spec:
  config:
    apiVersion: rabbitmq.com/v1beta1
    kind: RabbitmqCluster
    name: veba-rabbit
    namespace: vmware-system
EOF

kubectl apply -f /root/config/rabbit.yaml

echo -e "\e[92mDeploying Sockeye ..." > /dev/console

VEBA_BOM_FILE=/root/config/veba-bom.json
SOCKEYE_IMAGE=$(jq -r < ${VEBA_BOM_FILE} '.["sockeye"].containers[0].name')
SOCKEYE_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["sockeye"].containers[0].version')

cat > /root/config/sockeye.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: sockeye
  name: sockeye
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sockeye
  template:
    metadata:
      labels:
        app: sockeye
    spec:
      containers:
      - image: ${SOCKEYE_IMAGE}:${SOCKEYE_VERSION}
        name: sockeye
        ports:
          - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: sockeye
  name: sockeye
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: sockeye
  sessionAffinity: None
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: sockeye-trigger
spec:
  broker: default
  subscriber:
    ref:
      apiVersion: v1
      kind: Service
      name: sockeye
EOF

kubectl -n vmware-functions apply -f /root/config/sockeye.yaml