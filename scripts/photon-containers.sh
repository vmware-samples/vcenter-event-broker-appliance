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

mkdir -p /root/download && cd /root/download

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
ANTREA_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["antrea"].gitRepoTag')
wget https://github.com/vmware-tanzu/antrea/releases/download/${ANTREA_VERSION}/antrea.yml -O /root/download/antrea.yml
sed -i '/image:.*/i \        imagePullPolicy: IfNotPresent' /root/download/antrea.yml

echo '> Downloading Knative...'
KNATIVE_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["knative"].gitRepoTag')
curl -L https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-crds.yaml -o /root/download/serving-crds.yaml
curl -L https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-core.yaml -o /root/download/serving-core.yaml

curl -L https://github.com/knative/eventing/releases/download/${KNATIVE_VERSION}/eventing-crds.yaml -o /root/download/eventing-crds.yaml
curl -L https://github.com/knative/eventing/releases/download/${KNATIVE_VERSION}/eventing-core.yaml -o /root/download/eventing-core.yaml

echo '> Downloading RabbitMQ...'
RABBITMQ_OPERATOR_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["rabbitmq-operator"].gitRepoTag')
curl -L https://github.com/rabbitmq/cluster-operator/releases/download/${RABBITMQ_OPERATOR_VERSION}/cluster-operator.yml -o /root/download/cluster-operator.yml
sed -i '/image: rabbitmqoperator.*/i \        imagePullPolicy: IfNotPresent' /root/download/cluster-operator.yml

RABBITMQ_BROKER_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["rabbitmq-broker"].gitRepoTag')
curl -L https://github.com/knative-sandbox/eventing-rabbitmq/releases/download/${RABBITMQ_BROKER_VERSION}/rabbitmq-broker.yaml -o /root/download/rabbitmq-broker.yaml

echo '> Downloading Knative Contour...'
curl -L https://github.com/knative/net-contour/releases/download/${KNATIVE_VERSION}/contour.yaml -o /root/download/knative-contour.yaml
sed -i '1902i\      dnsPolicy: ClusterFirstWithHostNet\n      hostNetwork: true' /root/download/knative-contour.yaml
sed -i 's/type: LoadBalancer/type: NodePort/g' /root/download/knative-contour.yaml
sed -i 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' /root/download/knative-contour.yaml
cat > /root/download/contour-delegation.yaml << EOF
apiVersion: projectcontour.io/v1
kind: TLSCertificateDelegation
metadata:
  name: default-delegation
  namespace: contour-external
spec:
  delegations:
    - secretName: default-cert
      targetNamespaces:
      - "*"
EOF
curl -L https://github.com/knative/net-contour/releases/download/${KNATIVE_VERSION}/net-contour.yaml -o /root/download/net-contour.yaml

echo '> Downloading Local Path Provisioner...'
LOCAL_PROVISIONER_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["csi"].gitRepoTag')
LOCAL_STOARGE_VOLUME_PATH="/data/local-path-provisioner"
curl -L https://raw.githubusercontent.com/rancher/local-path-provisioner/${LOCAL_PROVISIONER_VERSION}/deploy/local-path-storage.yaml -o /root/download/local-path-storage.yaml
sed -i "s#/opt/local-path-provisioner#${LOCAL_STOARGE_VOLUME_PATH}#g" /root/download/local-path-storage.yaml
sed -i 's/busybox/busybox:latest/g' /root/download/local-path-storage.yaml
sed -i '/image: busybox.*/i \            imagePullPolicy: IfNotPresent' /root/download/local-path-storage.yaml