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
OPENFAAS_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["openfaas"].version')
git clone https://github.com/openfaas/faas-netes
cd faas-netes
git checkout ${OPENFAAS_VERSION}
sed -i 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' yaml/*.yml
cd ..

echo '> Downloading Contour...'
CONTOUR_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["contour"].version')
git clone https://github.com/projectcontour/contour.git
cd contour
git checkout ${CONTOUR_VERSION}
sed -i '/^---/i \      dnsPolicy: ClusterFirstWithHostNet\n      hostNetwork: true' examples/contour/03-envoy.yaml
sed -i 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' examples/contour/*.yaml
cd ..

echo '> Downloading Antrea...'
ANTREA_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["antrea"].version')
ANTREA_CONTAINER_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["antrea"].containers | .[] | select(.name | contains("antrea/antrea-ubuntu")).version')
wget https://github.com/vmware-tanzu/antrea/releases/download/${ANTREA_VERSION}/antrea.yml -O /root/download/antrea.yml
sed -i "s/image: antrea\/antrea-ubuntu:.*/image: antrea\/antrea-ubuntu:${ANTREA_CONTAINER_VERSION}/g" /root/download/antrea.yml
sed -i '/image:.*/i \        imagePullPolicy: IfNotPresent' /root/download/antrea.yml