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
            docker pull "$container_name:$container_version"
        done
    fi
done

cd /root/download

echo '> Downloading FaaS-Netes...'
OPENFAAS_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["openfaas"].gitRepoTag')
git clone https://github.com/openfaas/faas-netes
cd faas-netes
git checkout ${OPENFAAS_VERSION}
sed -i 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' yaml/*.yml
sed -i '15i\ \ namespace: "openfaas"' yaml/prometheus-rbac.yml
cd ..

echo '> Downloading Contour...'
CONTOUR_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["contour"].gitRepoTag')
git clone https://github.com/projectcontour/contour.git
cd contour
git checkout ${CONTOUR_VERSION}
sed -i "s/latest/${CONTOUR_VERSION}/g" examples/contour/02-job-certgen.yaml
cat >> examples/contour/03-envoy.yaml << EOF
      dnsPolicy: ClusterFirstWithHostNet
      hostNetwork: true
EOF
sed -i 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' examples/contour/*.yaml
cd ..

echo '> Downloading Antrea...'
mkdir -p /root/download/antrea/templates
ANTREA_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["antrea"].gitRepoTag')
ANTREA_TEMPLATE=/root/download/antrea/templates/antrea-template.yml
ANTREA_OVERLAY=/root/download/antrea/overlay.yaml
ANTREA_CONFIG=/root/download/antrea.yml
curl -L https://github.com/vmware-tanzu/antrea/releases/download/${ANTREA_VERSION}/antrea.yml -o ${ANTREA_TEMPLATE}
ytt -f ${ANTREA_OVERLAY} -f ${ANTREA_TEMPLATE} > ${ANTREA_CONFIG}

echo '> Downloading Knative...'
KNATIVE_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["knative"].gitRepoTag')
curl -L https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-crds.yaml -o /root/download/serving-crds.yaml
curl -L https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-core.yaml -o /root/download/serving-core.yaml

curl -L https://github.com/knative/eventing/releases/download/${KNATIVE_VERSION}/eventing-crds.yaml -o /root/download/eventing-crds.yaml
curl -L https://github.com/knative/eventing/releases/download/${KNATIVE_VERSION}/eventing-core.yaml -o /root/download/eventing-core.yaml

echo '> Downloading RabbitMQ...'
mkdir -p /root/download/rabbitmq-operator/templates
RABBITMQ_OPERATOR_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["rabbitmq-operator"].gitRepoTag')
RABBITMQ_OPERATOR_TEMPLATE=/root/download/rabbitmq-operator/templates/cluster-operator-template.yml
RABBITMQ_OPERATOR_OVERLAY=/root/download/rabbitmq-operator/overlay.yaml
RABBITMQ_OPERATOR_CONFIG=/root/download/cluster-operator.yml
curl -L https://github.com/rabbitmq/cluster-operator/releases/download/${RABBITMQ_OPERATOR_VERSION}/cluster-operator.yml -o ${RABBITMQ_OPERATOR_TEMPLATE}
ytt -f ${RABBITMQ_OPERATOR_OVERLAY} -f ${RABBITMQ_OPERATOR_TEMPLATE} > ${RABBITMQ_OPERATOR_CONFIG}

RABBITMQ_BROKER_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["rabbitmq-broker"].gitRepoTag')
curl -L https://github.com/knative-sandbox/eventing-rabbitmq/releases/download/${RABBITMQ_BROKER_VERSION}/rabbitmq-broker.yaml -o /root/download/rabbitmq-broker.yaml

echo '> Downloading Knative Contour...'
mkdir -p /root/download/knative-contour/templates
KNATIVE_CONTOUR_TEMPLATE=/root/download/knative-contour/templates/knative-contour.yaml
KNATIVE_CONTOUR_OVERLAY=/root/download/knative-contour/overlay.yaml
KNATIVE_CONTOUR_CONFIG=/root/download/knative-contour.yaml
curl -L https://github.com/knative/net-contour/releases/download/${KNATIVE_VERSION}/contour.yaml -o ${KNATIVE_CONTOUR_TEMPLATE}
ytt -f ${KNATIVE_CONTOUR_OVERLAY} -f ${KNATIVE_CONTOUR_TEMPLATE} > ${KNATIVE_CONTOUR_CONFIG}

curl -L https://github.com/knative/net-contour/releases/download/${KNATIVE_VERSION}/net-contour.yaml -o /root/download/net-contour.yaml

echo '> Downloading Local Path Provisioner...'
mkdir -p /root/download/local-provisioner/templates
LOCAL_PROVISIONER_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["csi"].gitRepoTag')
LOCAL_STOARGE_VOLUME_PATH="/data/local-path-provisioner"
LOCAL_PROVISIONER_TEMPLATE=/root/download/local-provisioner/templates/local-path-storage-template.yaml
LOCAL_PROVISIONER_OVERLAY=/root/download/local-provisioner/overlay.yaml
LOCAL_PROVISIONER_CONFIG=/root/download/local-path-storage.yaml
curl -L https://raw.githubusercontent.com/rancher/local-path-provisioner/${LOCAL_PROVISIONER_VERSION}/deploy/local-path-storage.yaml -o ${LOCAL_PROVISIONER_TEMPLATE}
ytt --data-value path=${LOCAL_STOARGE_VOLUME_PATH} --data-value-file bom=${VEBA_BOM_FILE} -f ${LOCAL_PROVISIONER_OVERLAY} -f ${LOCAL_PROVISIONER_TEMPLATE} > ${LOCAL_PROVISIONER_CONFIG}