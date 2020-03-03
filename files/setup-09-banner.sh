#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Login Banner

set -euo pipefail

echo -e "\e[92mCreating Login Banner ..." > /dev/console

HOSTNAME=$(hostname -f)

if [ "${EVENT_PROCESSOR_TYPE}" == "OpenFaaS" ]; then
    cat << EOF > /etc/issue
Welcome to the vCenter Event Broker Appliance

Appliance Status: https://${HOSTNAME}/status
Install Logs: https://${HOSTNAME}/bootstrap
Appliance Statistics: https://${HOSTNAME}/stats
OpenFaaS UI: https://${HOSTNAME}

EOF
else
    cat << EOF > /etc/issue
Welcome to the vCenter Event Broker Appliance

Appliance Status: https://${HOSTNAME}/status
Install Logs: https://${HOSTNAME}/bootstrap
Appliance Statistics: https://${HOSTNAME}/stats

EOF
fi

/usr/sbin/agetty --reload