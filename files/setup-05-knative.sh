#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Knative

set -euo pipefail

echo -e "\e[92mDeploying Knative Serving ..." > /dev/console
kubectl apply -f /root/download/serving-crds.yaml
kubectl apply -f /root/download/serving-core.yaml
kubectl wait deployment --all --timeout=${KUBECTL_WAIT} --for=condition=Available -n knative-serving
kubectl apply -f /root/download/knative-contour.yaml
kubectl apply -f /root/download/net-contour.yaml
kubectl patch configmap/config-network --namespace knative-serving --type merge --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
kubectl wait deployment --all --timeout=${KUBECTL_WAIT} --for=condition=Available -n contour-external
kubectl wait deployment --all --timeout=${KUBECTL_WAIT} --for=condition=Available -n contour-internal

echo -e "\e[92mDeploying Knative Eventing ..." > /dev/console
kubectl apply -f /root/download/eventing-crds.yaml
kubectl apply -f /root/download/eventing-core.yaml
kubectl wait pod --timeout=${KUBECTL_WAIT} --for=condition=Ready -l '!job-name' -n knative-eventing

echo -e "\e[92mDeploying RabbitMQ Cluster Operator ..." > /dev/console
kubectl apply -f /root/download/cluster-operator.yml

echo -e "\e[92mDeploying Cert-Manager ..." > /dev/console
kubectl apply -f /root/download/cert-manager.yaml
kubectl wait deployment --all --timeout=${KUBECTL_WAIT} --for=condition=Available -n cert-manager

echo -e "\e[92mDeploying RabbitMQ Messaging Operator ..." > /dev/console
kubectl apply -f /root/download/messaging-topology-operator-with-certmanager.yaml
kubectl wait deployment --all --timeout=${KUBECTL_WAIT} --for=condition=Available -n rabbitmq-system

echo -e "\e[92mDeploying RabbitMQ Broker ..." > /dev/console
kubectl apply -f /root/download/rabbitmq-broker.yaml

echo -e "\e[92mDeploying RabbitMQ Cluster ..." > /dev/console
RABBITMQ_CONFIG_TEMPLATE=/root/config/knative/templates/rabbit-template.yaml
RABBITMQ_CONFIG=/root/config/knative/rabbit.yaml
ytt --data-value-file bom=${VEBA_BOM_FILE} -f ${RABBITMQ_CONFIG_TEMPLATE} > ${RABBITMQ_CONFIG}
kubectl apply -f ${RABBITMQ_CONFIG}

echo -e "\e[92mDeploying Sockeye ..." > /dev/console

VEBA_BOM_FILE=/root/config/veba-bom.json
VEBA_CONFIG_FILE=/root/config/veba-config.json

# Sockeye Config files
SOCKEYE_TEMPLATE=/root/config/knative/templates/sockeye-template.yaml
SOCKEYE_CONFIG=/root/config/knative/sockeye.yaml

# Apply YTT overlay
ytt --data-value-file bom=${VEBA_BOM_FILE} -f ${SOCKEYE_TEMPLATE} > ${SOCKEYE_CONFIG}

kubectl -n vmware-functions apply -f ${SOCKEYE_CONFIG}