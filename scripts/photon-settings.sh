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
cd /etc/yum.repos.d/
sed -i 's/dl.bintray.com\/vmware/packages.vmware.com\/photon\/$releasever/g' photon.repo photon-updates.repo photon-extras.repo photon-debuginfo.repo
tdnf -y update photon-repos
tdnf -y remove docker
tdnf clean all
tdnf makecache
tdnf -y update

echo '> Installing Additional Packages...'
tdnf install -y \
  minimal \
  logrotate \
  wget \
  git \
  unzip \
  tar \
  jq \
  parted

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
K8S_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["kubernetes"].gitRepoTag' | sed 's/v//g')
# Ensure kubelet is updated to the latest desired K8s version
tdnf install -y kubelet-${K8S_VERSION} kubectl-${K8S_VERSION} kubeadm-${K8S_VERSION}

echo '> Downloading Kn CLI'
KNATIVE_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["knative-cli"].version')
wget https://github.com/knative/client/releases/download/knative-${KNATIVE_VERSION}/kn-linux-amd64
chmod +x kn-linux-amd64
mv kn-linux-amd64 /usr/local/bin/kn

echo '> Downloading YTT CLI'
YTT_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["ytt-cli"].version')
wget https://github.com/vmware-tanzu/carvel-ytt/releases/download/${YTT_VERSION}/ytt-linux-amd64
chmod +x ytt-linux-amd64
mv ytt-linux-amd64 /usr/local/bin/ytt

echo '> Downloading Containerd'
CONTAINERD_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["containerd"].version')
curl -L https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz -o /root/download/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz
tar -zxvf /root/download/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz -C /usr
rm -f /root/download/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz
containerd config default > /etc/containerd/config.toml
cat > /usr/lib/systemd/system/containerd.service <<EOF
[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target
[Service]
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/bin/containerd
Restart=always
RestartSec=5
KillMode=process
Delegate=yes
OOMScoreAdjust=-999
LimitNOFILE=1048576
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
[Install]
WantedBy=multi-user.target
EOF
systemctl enable containerd
systemctl start containerd

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

echo '> Enable contrackd log rotation...'
cat > /etc/logrotate.d/contrackd << EOF
/var/log/conntrackd*.log {
	missingok
	size 5M
	rotate 3
        maxage 7
	compress
	copytruncate
}
EOF

#TODO - Temp fix until this is resolved in Photon OS 4.x
sed -i '1,238d' /etc/conntrackd/conntrackd.conf

echo '> Done'
