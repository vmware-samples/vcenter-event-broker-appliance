#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail

if [ -e /root/ran_customization ]; then
    exit
fi

# Extract all OVF Properties
# TODO - Shouldn't need to pass all this - can just parse all OVF properties?
PROPS=(
veba_debug
hostname
ip_address
netmask
gateway
dns_server
dns_domain
ntp_server
http_proxy
https_proxy
proxy_username
proxy_password
no_proxy
root_password
enable_ssh
vcenter_server
vcenter_username
vcenter_password
vcenter_username_for_veba_ui
vcenter_password_for_veba_ui
vcenter_disable_tls
horizon_enabled
horizon_server
horizon_domain
horizon_username
horizon_password
horizon_disable_tls
webhook_enabled
webhook_username
webhook_password
custom_veba_tls_private_key
custom_veba_tls_ca_cert
pod_network_cidr
syslog_server_hostname
syslog_server_port
syslog_server_protocol
syslog_server_format)
/root/setup/getOvfProperties.py ${PROPS[@]}

source /root/config/shell_env
rm /root/config/shell_env

HOSTNAME=$(echo "$HOSTNAME" | tr '[:upper:]' '[:lower:]')
NETMASK=$(echo "$NETMASK" | awk -F ' ' '{print $1}')
ENABLE_SSH=$(echo "$ENABLE_SSH" | tr '[:upper:]' '[:lower:]')

KUBECTL_WAIT="10m"
LOCAL_STORAGE_DISK="/dev/sdb"
LOCAL_STOARGE_VOLUME_PATH="/data"
export KUBECONFIG="/root/.kube/config"

VEBA_LOG_FILE=/var/log/bootstrap.log
if [ ${VEBA_DEBUG} == "True" ]; then
	VEBA_LOG_FILE=/var/log/bootstrap-debug.log
	set -x
	exec 2>> ${VEBA_LOG_FILE}
	echo
	echo "### WARNING -- DEBUG LOG CONTAINS ALL EXECUTED COMMANDS WHICH INCLUDES CREDENTIALS -- WARNING ###"
	echo "### WARNING --             PLEASE REMOVE CREDENTIALS BEFORE SHARING LOG            -- WARNING ###"
	echo
fi

# Determine Event Providers
EVENT_PROVIDERS=("vcenter")

if [ ${WEBHOOK_ENABLED} == "True" ]; then
	EVENT_PROVIDERS+=("webhook")
fi

if [ ${HORIZON_ENABLED} == "True" ]; then
	EVENT_PROVIDERS+=("horizon")
fi

# Customize the POD CIDR Network if provided or else default to 10.10.0.0/16
if [ -z "${POD_NETWORK_CIDR}" ]; then
	POD_NETWORK_CIDR="10.16.0.0/16"
fi

echo -e "\e[92mStarting Customization ..." > /dev/console

echo -e "\e[92mStarting OS Configuration ..." > /dev/console
. /root/setup/setup-01-os.sh

echo -e "\e[92mStarting Network Proxy Configuration ..." > /dev/console
. /root/setup/setup-02-proxy.sh

echo -e "\e[92mStarting Network Configuration ..." > /dev/console
. /root/setup/setup-03-network.sh

echo -e "\e[92mStarting Kubernetes Configuration ..." > /dev/console
. /root/setup/setup-04-kubernetes.sh

echo -e "\e[92mStarting Knative Configuration ..." > /dev/console
. /root/setup/setup-05-knative.sh

echo -e "\e[92mStarting VMware Event Processor Configuration ..." > /dev/console
. /root/setup/setup-06-event-processor.sh

echo -e "\e[92mStarting VMware Event Router Configuration ..." > /dev/console
. /root/setup/setup-07-event-router.sh

echo -e "\e[92mStarting TinyWWW Configuration ..." > /dev/console
. /root/setup/setup-08-tinywww.sh

echo -e "\e[92mStarting Ingress Router Configuration ..." > /dev/console
. /root/setup/setup-09-ingress.sh

if [[ ! -z ${VCENTER_USERNAME_FOR_VEBA_UI} ]] && [[ ! -z ${VCENTER_PASSWORD_FOR_VEBA_UI} ]]; then
	echo -e "\e[92mStarting Knative UI Configuration ..." > /dev/console
	. /root/setup/setup-010-veba-ui.sh
fi

if [ -n "${SYSLOG_SERVER_HOSTNAME}" ]; then
	echo -e "\e[92mStarting FluentBit Configuration ..." > /dev/console
	. /root/setup/setup-011-fluentbit.sh
fi

echo -e "\e[92mStarting cAdvisor Configuration ..." > /dev/console
. /root/setup/setup-012-cadvisor.sh

echo -e "\e[92mStarting VEBA Endpoint File Configuration ..." > /dev/console
. /root/setup/setup-098-dcui-endpoints.sh

echo -e "\e[92mStarting OS Banner Configuration ..."> /dev/console
. /root/setup/setup-099-banner.sh &

echo -e "\e[92mCustomization Completed ..." > /dev/console

# Clear guestinfo.ovfEnv
if [ ${VEBA_DEBUG} == "False" ]; then
	vmtoolsd --cmd "info-set guestinfo.ovfEnv NULL"
fi

sleep .1
echo -e "\e[92mDone!\e[0m" > /dev/console

# Ensure we don't run customization again
touch /root/ran_customization