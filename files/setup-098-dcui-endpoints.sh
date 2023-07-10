#!/bin/bash
# Copyright 2023 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail

DCUI_ENDPOINTS_FILE=/etc/veba-endpoints

# The initial endpoints file containing common endpoints for all deployments
# Max of only 7 can be rendered in DCUI
cat > ${DCUI_ENDPOINTS_FILE} <<EOF
Appliance Configuration,Install Logs,/bootstrap
Appliance Configuration,Resource Utilization,/top
Appliance Configuration,Events,/events
EOF

# For Webhook Provider
if [ ${WEBHOOK_ENABLED} == "True" ]; then
    echo "Appliance Configuration,Webhook,/webhook" >> ${DCUI_ENDPOINTS_FILE}
fi

# For Webhook Provider
if [ ${WEBHOOK_ENABLED} == "True" ]; then
    echo "Appliance Provider Stats,Webhook,/stats/webhook" >> ${DCUI_ENDPOINTS_FILE}
fi
