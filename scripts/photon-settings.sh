#!/bin/bash -eux
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

##
## Misc configuration
##

echo '> vCenter Event Broker Appliance Settings...'

echo '> Disable IPv6'
echo "net.ipv6.conf.all.disable_ipv6 = 1" >> /etc/sysctl.conf

echo '> Applying latest Updates...'
tdnf -y update
#tdnf upgrade linux-esx

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
  kubernetes-kubeadm \
  jq

echo '> Creating directory for setup scripts and configuration files'
mkdir -p /root/setup
mkdir -p /root/config

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
