#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# OS Specific Settings where ordering does not matter

set -euo pipefail

echo -e "\e[92mConfiguring OS Root password ..." > /dev/console
echo "root:${ROOT_PASSWORD}" | /usr/sbin/chpasswd