#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Login Banner

set -euo pipefail

echo -e "\e[92mCreating Login Banner ..." > /dev/console

cat << EOF > /etc/issue
Welcome to the VMware Event Broker Appliance (VEBA)

EOF

/usr/sbin/agetty --reload
