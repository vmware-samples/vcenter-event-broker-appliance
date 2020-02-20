#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Networking

set -euo pipefail

echo -e "\e[92mConfiguring Static IP Address ..." > /dev/console

NETWORK_CONFIG_FILE=$(ls /etc/systemd/network | grep .network)
cat > /etc/systemd/network/${NETWORK_CONFIG_FILE} << __CUSTOMIZE_PHOTON__
[Match]
Name=e*

[Network]
Address=${IP_ADDRESS}/${NETMASK}
Gateway=${GATEWAY}
DNS=${DNS_SERVER}
Domain=${DNS_DOMAIN}
__CUSTOMIZE_PHOTON__

echo -e "\e[92mConfiguring NTP ..." > /dev/console
cat > /etc/systemd/timesyncd.conf << __CUSTOMIZE_PHOTON__

[Match]
Name=e*

[Time]
NTP=${NTP_SERVER}
__CUSTOMIZE_PHOTON__

echo -e "\e[92mConfiguring hostname ..." > /dev/console
echo "${IP_ADDRESS} ${HOSTNAME}" >> /etc/hosts
hostnamectl set-hostname ${HOSTNAME}

echo -e "\e[92mRestarting Network ..." > /dev/console
systemctl restart systemd-networkd

echo -e "\e[92mRestarting Timesync ..." > /dev/console
systemctl restart systemd-timesyncd
