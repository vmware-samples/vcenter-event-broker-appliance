#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Docker and Kubernetes

set -euo pipefail

echo -e "\e[92mStarting Docker ..." > /dev/console
systemctl daemon-reload
systemctl start docker.service
systemctl enable docker.service

echo -e "\e[92mDisabling/Stopping IP Tables  ..." > /dev/console
systemctl stop iptables
systemctl disable iptables

# Setup k8s
echo -e "\e[92mSetting up k8s ..." > /dev/console
HOME=/root
kubeadm init --ignore-preflight-errors SystemVerification --skip-token-print --config /root/config/kubeconfig.yml
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
echo -e "\e[92mDeloying kubeadm ..." > /dev/console

# Customize the POD CIDR Network if provided or else default to 10.99.0.0/20
if [ -z "${POD_NETWORK_CIDR}" ]; then
    POD_NETWORK_CIDR="10.99.0.0/20"
fi

sed -i "s#POD_NETWORK_CIDR#${POD_NETWORK_CIDR}#g" /root/config/weave.yaml

kubectl --kubeconfig /root/.kube/config apply -f /root/config/weave.yaml
kubectl --kubeconfig /root/.kube/config taint nodes --all node-role.kubernetes.io/master-

echo -e "\e[92mStarting k8s ..." > /dev/console
systemctl enable kubelet.service

while [[ $(systemctl is-active kubelet.service) == "inactive" ]]
do
    echo -e "\e[92mk8s service is still inactive, sleeping for 10secs" > /dev/console
    sleep 10
done