#!/bin/bash -eux
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

echo '> Pre-Downloading Kubeadm Docker Containers'

CONTAINERS=(
k8s.gcr.io/kube-apiserver:v1.14.9
k8s.gcr.io/kube-controller-manager:v1.14.9
k8s.gcr.io/kube-scheduler:v1.14.9
k8s.gcr.io/kube-proxy:v1.14.9
k8s.gcr.io/pause:3.1
k8s.gcr.io/etcd:3.3.10
k8s.gcr.io/coredns:1.3.1
antrea/antrea-ubuntu:v0.6.0
embano1/tinywww:latest
projectcontour/contour:v1.0.0-beta.1
openfaas/faas-netes:0.9.0
openfaas/gateway:0.17.4
openfaas/basic-auth-plugin:0.17.0
openfaas/queue-worker:0.8.0
openfaas/faas-idler:0.2.1
envoyproxy/envoy:v1.11.1
prom/prometheus:v2.11.0
prom/alertmanager:v0.18.0
nats-streaming:0.11.2
vmware/veba-event-router:${VEBA_VERSION}
)

for i in ${CONTAINERS[@]};
do
	docker pull $i
done

mkdir -p /root/download && cd /root/download

echo '> Downloading FaaS-Netes...'
git clone https://github.com/openfaas/faas-netes
cd faas-netes
git checkout 0.9.2
sed -i 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' yaml/*.yml
cd ..

echo '> Downloading Contour...'
git clone https://github.com/projectcontour/contour.git
cd contour
git checkout v1.0.0-beta.1
sed -i '/^---/i \      dnsPolicy: ClusterFirstWithHostNet\n      hostNetwork: true' examples/contour/03-envoy.yaml
sed -i 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' examples/contour/*.yaml
cd ..

echo '> Downloading Antrea...'
wget https://github.com/vmware-tanzu/antrea/releases/download/v0.6.0/antrea.yml -O /root/download/antrea.yml
sed -i 's/image: antrea\/antrea-ubuntu:.*/image: antrea\/antrea-ubuntu:v0.6.0/g' /root/download/antrea.yml
sed -i '/image:.*/i \        imagePullPolicy: IfNotPresent' /root/download/antrea.yml