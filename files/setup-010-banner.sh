#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Login Banner

set -euo pipefail

echo -e "\e[92mCreating Login Banner ..." > /dev/console

if [ "${EVENT_PROCESSOR_TYPE}" == "OpenFaaS" ]; then
    cat << EOF > /etc/issue
Welcome to the vCenter Event Broker Appliance

EOF
fi

/usr/sbin/agetty --reload