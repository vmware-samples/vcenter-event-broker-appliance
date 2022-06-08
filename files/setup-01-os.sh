#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# OS Specific Settings where ordering does not matter

set -euo pipefail

if [ "${ENABLE_SSH}" == "true" ]; then
    systemctl enable sshd
    systemctl start sshd
else
    systemctl disable sshd
    systemctl stop sshd
fi

echo -e "\e[92mConfiguring OS Root password ..." > /dev/console
echo "root:${ROOT_PASSWORD}" | /usr/sbin/chpasswd

echo -e "\e[92mConfiguring IP Tables for Antrea ..." > /dev/console
iptables -A INPUT -i gw0 -j ACCEPT
iptables-save > /etc/systemd/scripts/ip4save

echo -e "\e[92mConfiguring Local Storage Volume ..." > /dev/console
parted ${LOCAL_STORAGE_DISK} --script mklabel gpt mkpart primary ext3 0% 100%
mkfs -t ext3 ${LOCAL_STORAGE_DISK}1
mkdir -p ${LOCAL_STOARGE_VOLUME_PATH}
chmod 777 ${LOCAL_STOARGE_VOLUME_PATH}
echo "${LOCAL_STORAGE_DISK}1       ${LOCAL_STOARGE_VOLUME_PATH}       ext3    defaults        0        0" >> /etc/fstab
mount ${LOCAL_STOARGE_VOLUME_PATH}