#!/bin/bash -eux
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

##
## Misc configuration
##

echo '> VEBA Settings...'

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
  kubernetes-kubeadm

echo '> Done'
