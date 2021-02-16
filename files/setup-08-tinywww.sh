#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Deploy TinyWWW Pod

set -euo pipefail

if [ ${VEBA_DEBUG} == "True" ]; then
    kubectl apply -f /root/config/tinywww-debug.yml
else
    kubectl apply -f /root/config/tinywww.yml
fi