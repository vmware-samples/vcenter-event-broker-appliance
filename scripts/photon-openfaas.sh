#!/bin/bash -eux
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

mkdir -p /root/download && cd /root/download
echo '> Downloading FaaS-Netes...'
git clone https://github.com/openfaas/faas-netes
cd faas-netes
git checkout 0.9.2
sed -i 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' yaml/*.yml
cd ..

echo '> Downloading OpenFaaS vCenter Connector...'
git clone https://github.com/openfaas-incubator/vcenter-connector
cd vcenter-connector
git checkout fefb5881ab2dfe05207f7f0b65c37ef5d1db34af
cd ..
mv vcenter-connector/yaml/kubernetes/connector-dep.yml vcenter-connector/yaml/kubernetes/connector-dep.yml.orig
cp vcenter-connector/yaml/kubernetes/connector-dep.yml.orig vcenter-connector/yaml/kubernetes/connector-dep.yml
sed -i '/image:.*/a \        imagePullPolicy: IfNotPresent' vcenter-connector/yaml/kubernetes/connector-dep.yml

echo '> Downloading Contour...'
git clone https://github.com/projectcontour/contour.git
cd contour
git checkout v1.0.0-beta.1
sed -i '/^---/i \      dnsPolicy: ClusterFirstWithHostNet\n      hostNetwork: true' examples/contour/03-envoy.yaml
sed -i 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' examples/contour/*.yaml
cd ..

cat << EOF > /etc/issue
Welcome to the vCenter Event Broker Appliance

Appliance Status: https://[IP]/status
Install Logs: https://[IP]/bootstrap
OpenFaaS UI: https://[IP]


EOF
