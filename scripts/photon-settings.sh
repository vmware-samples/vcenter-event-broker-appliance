#!/bin/bash -eux
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

##
## Misc configuration
##

echo '> vCenter Event Broker Appliance Settings...'

VEBA_BOM_FILE=/root/config/veba-bom.json

echo '> Disable IPv6'
echo "net.ipv6.conf.all.disable_ipv6 = 1" >> /etc/sysctl.conf

echo '> Applying latest Updates...'
tdnf -y update

echo '> Installing Additional Packages...'
tdnf install -y \
  less \
  logrotate \
  curl \
  wget \
  git \
  unzip \
  awk \
  tar \
  jq

echo '> Adding K8s Repo'
curl -L https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg -o /etc/pki/rpm-gpg/GOOGLE-RPM-GPG-KEY
rpm --import /etc/pki/rpm-gpg/GOOGLE-RPM-GPG-KEY
cat > /etc/yum.repos.d/kubernetes.repo << EOF
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/GOOGLE-RPM-GPG-KEY
EOF
K8S_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["kubernetes"].version' | sed 's/v//g')
# Ensure kubelet is updated to the latest desired K8s version
tdnf install -y kubelet-${K8S_VERSION} kubectl-${K8S_VERSION} kubeadm-${K8S_VERSION}

echo '> Creating directory for setup scripts and configuration files'
mkdir -p /root/setup

echo '> Creating tools.conf to prioritize eth0 interface...'
cat > /etc/vmware-tools/tools.conf << EOF
[guestinfo]
primary-nics=eth0
low-priority-nics=weave,docker0

[guestinfo]
exclude-nics=veth*,vxlan*,datapath
EOF

cat > /etc/veba-release << EOF
Version: ${VEBA_VERSION}
Commit: ${VEBA_COMMIT}
EOF

echo '> Creating VEBA DCUI systemd unit file...'
mkdir -p /usr/lib/systemd/system/getty@tty1.service.d/
cat > /usr/lib/systemd/system/getty@tty1.service.d/dcui_override.conf << EOF
[Unit]
Description=
Description=VEBA DCUI
After=
After=network-online.target

[Service]
ExecStart=
ExecStart=-/usr/bin/veba-dcui
Restart=always
RestartSec=1sec
StandardOutput=tty
StandardInput=tty
StandardError=journal
TTYPath=/dev/tty1
TTYReset=yes
TTYVHangup=yes
KillMode=process
EOF

systemctl enable getty@tty1.service

echo '> Done'
