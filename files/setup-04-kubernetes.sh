#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Containerd and Kubernetes

set -euo pipefail

echo -e "\e[92mStarting Containerd ..." > /dev/console
systemctl enable containerd
systemctl start containerd

echo -e "\e[92mDisabling/Stopping IP Tables  ..." > /dev/console
systemctl stop iptables
systemctl disable iptables

# Setup k8s
echo -e "\e[92mSetting up k8s ..." > /dev/console

VEBA_BOM_FILE=/root/config/veba-bom.json
VEBA_CONFIG_FILE=/root/config/veba-config.json

# Kubernetes Config Files
K8S_TEMPLATE=/root/config/kubernetes/templates/kubeconfig-template.yaml
K8S_CONFIG=/root/config/kubernetes/kubeconfig.yaml

# Apply YTT overlay
ytt --data-value-file bom=${VEBA_BOM_FILE} --data-value-file config=${VEBA_CONFIG_FILE} -f ${K8S_TEMPLATE} > ${K8S_CONFIG}

echo -e "\e[92mDeploying kubeadm ..." > /dev/console
HOME=/root
kubeadm init --ignore-preflight-errors SystemVerification --skip-token-print --config ${K8S_CONFIG}
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
kubectl taint nodes --all node-role.kubernetes.io/master-

echo -e "\e[92mDeploying Antrea ..." > /dev/console
kubectl apply -f /root/download/antrea.yml

echo -e "\e[92mStarting k8s ..." > /dev/console
systemctl enable kubelet.service

while [[ $(systemctl is-active kubelet.service) == "inactive" ]]
do
    echo -e "\e[92mk8s service is still inactive, sleeping for 10secs" > /dev/console
    sleep 10
done

echo -e "\e[92mDeploying Local Storage Provisioner ..." > /dev/console
mkdir -p ${LOCAL_STOARGE_VOLUME_PATH}/local-path-provisioner
chmod 777 ${LOCAL_STOARGE_VOLUME_PATH}/local-path-provisioner
kubectl apply -f /root/download/local-path-storage.yaml
kubectl patch sc local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

echo -e "\e[92mCreating VMware namespaces ..." > /dev/console
kubectl create namespace vmware-system
kubectl create namespace vmware-functions