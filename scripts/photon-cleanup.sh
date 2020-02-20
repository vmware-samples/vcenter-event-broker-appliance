#!/bin/bash -eux
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

##
## cleanup everything we can to make the OVA as small as possible
##

# Clear tdnf cache

echo '> Clearing tdnf cache...'
tdnf clean all


# Cleanup log files
echo '> Removing Log files...'
cat /dev/null > /var/log/wtmp 2>/dev/null
logrotate -f /etc/logrotate.conf 2>/dev/null
find /var/log -type f -delete
rm -rf /var/log/journal/*
rm -f /var/lib/dhcp/*

# Zero out the free space to save space in the final image, blocking 'til
# written otherwise, the disk image won't be zeroed, and/or Packer will try to
# kill the box while the disk is still full and that's bad.  The dd will run
# 'til failure, so (due to the 'set -e' above), ignore that failure.  Also,
# really make certain that both the zeros and the file removal really sync; the
# extra sleep 1 and sync shouldn't be necessary, but...)
echo '> Zeroing device to make space...'
dd if=/dev/zero of=/EMPTY bs=1M || true; sync; sleep 1; sync
rm -f /EMPTY; sync; sleep 1; sync

echo '> Setting random root password...'
RANDOM_PASSWORD=$(< /dev/urandom tr -dc _A-Z-a-z-0-9 | head -c${1:-32};echo;)
echo "root:${RANDOM_PASSWORD}" | /usr/sbin/chpasswd

echo '> Disabling SSH ...'
systemctl disable sshd
systemctl stop sshd

echo '> Clearing history ...'
unset HISTFILE && history -c && rm -fr /root/.bash_history

echo '> Done'
