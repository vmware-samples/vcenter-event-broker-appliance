#!/bin/bash -eux
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

echo '> Pre-Downloading Kubeadm Docker Containers'

VEBA_BOM_FILE=/root/config/veba-bom.json

for component_name in $(jq '. | keys | .[]' ${VEBA_BOM_FILE});
do
    HAS_CONTAINERS=$(jq ".$component_name | select(.containers != null)" ${VEBA_BOM_FILE})
    if [ "${HAS_CONTAINERS}" != "" ]; then
        for i in $(jq ".$component_name.containers | keys | .[]" ${VEBA_BOM_FILE}); do
            value=$(jq -r ".$component_name.containers[$i]" ${VEBA_BOM_FILE});
            container_name=$(jq -r '.name' <<< "$value");
            container_version=$(jq -r '.version' <<< "$value");
            if [ "${container_name}" != "kindest/node" ]; then
                crictl pull "$container_name:$container_version"
            fi
        done
    fi
done

cd /root/download

echo '> Downloading Antrea...'
mkdir -p /root/download/antrea/templates
ANTREA_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["antrea"].gitRepoTag')
ANTREA_TEMPLATE=/root/download/antrea/templates/antrea-template.yml
ANTREA_OVERLAY=/root/download/antrea/overlay.yaml
ANTREA_CONFIG=/root/download/antrea.yml
curl -L https://github.com/vmware-tanzu/antrea/releases/download/${ANTREA_VERSION}/antrea.yml -o ${ANTREA_TEMPLATE}
ytt -f ${ANTREA_OVERLAY} -f ${ANTREA_TEMPLATE} > ${ANTREA_CONFIG}

echo '> Downloading Knative...'
KNATIVE_SERVING_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["knative-serving"].gitRepoTag')
curl -L https://github.com/knative/serving/releases/download/knative-${KNATIVE_SERVING_VERSION}/serving-crds.yaml -o /root/download/serving-crds.yaml
curl -L https://github.com/knative/serving/releases/download/knative-${KNATIVE_SERVING_VERSION}/serving-core.yaml -o /root/download/serving-core.yaml

KNATIVE_EVENTING_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["knative-eventing"].gitRepoTag')
curl -L https://github.com/knative/eventing/releases/download/knative-${KNATIVE_EVENTING_VERSION}/eventing-crds.yaml -o /root/download/eventing-crds.yaml
curl -L https://github.com/knative/eventing/releases/download/knative-${KNATIVE_EVENTING_VERSION}/eventing-core.yaml -o /root/download/eventing-core.yaml

echo '> Downloading RabbitMQ Broker/Operator...'
mkdir -p /root/download/rabbitmq-operator/templates
RABBITMQ_OPERATOR_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["rabbitmq-operator"].gitRepoTag')
RABBITMQ_OPERATOR_TEMPLATE=/root/download/rabbitmq-operator/templates/cluster-operator-template.yml
RABBITMQ_OPERATOR_OVERLAY=/root/download/rabbitmq-operator/overlay.yaml
RABBITMQ_OPERATOR_CONFIG=/root/download/cluster-operator.yml
curl -L https://github.com/rabbitmq/cluster-operator/releases/download/${RABBITMQ_OPERATOR_VERSION}/cluster-operator.yml -o ${RABBITMQ_OPERATOR_TEMPLATE}
ytt -f ${RABBITMQ_OPERATOR_OVERLAY} -f ${RABBITMQ_OPERATOR_TEMPLATE} > ${RABBITMQ_OPERATOR_CONFIG}

RABBITMQ_BROKER_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["rabbitmq-broker"].gitRepoTag')
curl -L https://github.com/knative-sandbox/eventing-rabbitmq/releases/download/knative-${RABBITMQ_BROKER_VERSION}/rabbitmq-broker.yaml -o /root/download/rabbitmq-broker.yaml

echo '> Downloading RabbitMQ Messaging Operator...'
mkdir -p /root/download/rabbitmq-messaging-operator/templates
RABBITMQ_MESSAGING_OPERATOR_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["rabbitmq-messaging-topology-operator"].gitRepoTag')
RABBITMQ_MESSAGING_OPERATOR_TEMPLATE=/root/download/rabbitmq-messaging-operator/templates/messaging-topology-operator-with-certmanager-template.yaml
curl -L https://github.com/rabbitmq/messaging-topology-operator/releases/download/${RABBITMQ_MESSAGING_OPERATOR_VERSION}/messaging-topology-operator-with-certmanager.yaml -o ${RABBITMQ_MESSAGING_OPERATOR_TEMPLATE}
RABBITMQ_MESSAGING_OPERATOR_OVERLAY=/root/download/rabbitmq-messaging-operator/overlay.yaml
RABBITMQ_MESSAGING_OPERATOR_CONFIG=/root/download/messaging-topology-operator-with-certmanager.yaml
ytt -f ${RABBITMQ_MESSAGING_OPERATOR_OVERLAY} -f ${RABBITMQ_MESSAGING_OPERATOR_TEMPLATE} > ${RABBITMQ_MESSAGING_OPERATOR_CONFIG}

echo '> Downloading Cert-Manager...'
mkdir -p /root/download/cert-manager/templates
CERT_MANAGER_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["cert-manager"].gitRepoTag')
CERT_MANAGER_TEMPLATE=/root/download/cert-manager/templates/cert-manager-template.yaml
curl -L https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml -o ${CERT_MANAGER_TEMPLATE}
CERT_MANAGER_OVERLAY=/root/download/cert-manager/overlay.yaml
CERT_MANAGER_CONFIG=/root/download/cert-manager.yaml
ytt -f ${CERT_MANAGER_OVERLAY} -f ${CERT_MANAGER_TEMPLATE} > ${CERT_MANAGER_CONFIG}

echo '> Downloading Knative Contour...'
KNATIVE_CONTOUR_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["knative-contour"].gitRepoTag')
mkdir -p /root/download/knative-contour/templates
KNATIVE_CONTOUR_TEMPLATE=/root/download/knative-contour/templates/knative-contour.yaml
KNATIVE_CONTOUR_OVERLAY=/root/download/knative-contour/overlay.yaml
KNATIVE_CONTOUR_CONFIG=/root/download/knative-contour.yaml
curl -L https://github.com/knative/net-contour/releases/download/knative-${KNATIVE_CONTOUR_VERSION}/contour.yaml -o ${KNATIVE_CONTOUR_TEMPLATE}
ytt -f ${KNATIVE_CONTOUR_OVERLAY} -f ${KNATIVE_CONTOUR_TEMPLATE} > ${KNATIVE_CONTOUR_CONFIG}

curl -L https://github.com/knative/net-contour/releases/download/knative-${KNATIVE_CONTOUR_VERSION}/net-contour.yaml -o /root/download/net-contour.yaml

echo '> Downloading Local Path Provisioner...'
mkdir -p /root/download/local-provisioner/templates
LOCAL_PROVISIONER_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["csi"].gitRepoTag')
LOCAL_STOARGE_VOLUME_PATH="/data/local-path-provisioner"
LOCAL_PROVISIONER_TEMPLATE=/root/download/local-provisioner/templates/local-path-storage-template.yaml
LOCAL_PROVISIONER_OVERLAY=/root/download/local-provisioner/overlay.yaml
LOCAL_PROVISIONER_CONFIG=/root/download/local-path-storage.yaml
curl -L https://raw.githubusercontent.com/rancher/local-path-provisioner/${LOCAL_PROVISIONER_VERSION}/deploy/local-path-storage.yaml -o ${LOCAL_PROVISIONER_TEMPLATE}
ytt --data-value path=${LOCAL_STOARGE_VOLUME_PATH} --data-value-file bom=${VEBA_BOM_FILE} -f ${LOCAL_PROVISIONER_OVERLAY} -f ${LOCAL_PROVISIONER_TEMPLATE} > ${LOCAL_PROVISIONER_CONFIG}