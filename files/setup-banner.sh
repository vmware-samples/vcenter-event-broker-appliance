#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

HOSTNAME_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.hostname")
HOSTNAME=$(echo "${HOSTNAME_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')

sed -i "s/\[IP\]/${HOSTNAME}/g" /etc/issue
PID=$(ps -ef | grep agetty | grep -v grep|awk '{print $2}')
kill -9 ${PID}
#systemctl restart getty@tty1
