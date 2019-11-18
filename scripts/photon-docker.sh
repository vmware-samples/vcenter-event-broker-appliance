#!/bin/bash -eux
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

##
## Enable Docker
##

echo '> Enabling Docker...'

systemctl enable docker

echo '> Done'

