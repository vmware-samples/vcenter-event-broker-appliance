#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail

# Extract all OVF Properties
VEBA_DEBUG=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.debug" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
HOSTNAME=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.hostname" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
IP_ADDRESS=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.ipaddress" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
NETMASK=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.netmask" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}' | awk -F ' ' '{print $1}')
GATEWAY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.gateway" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
DNS_SERVER=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.dns" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
DNS_DOMAIN=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.domain" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
NTP_SERVER=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.ntp" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
HTTP_PROXY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.http_proxy" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
HTTPS_PROXY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.https_proxy" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
PROXY_USERNAME=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.proxy_username" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
PROXY_PASSWORD=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.proxy_password" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
NO_PROXY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.no_proxy" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
ROOT_PASSWORD=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.root_password" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
ENABLE_SSH=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.enable_ssh" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}' | tr '[:upper:]' '[:lower:]')
VCENTER_SERVER=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.vcenter_server" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
VCENTER_USERNAME=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.vcenter_username" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
VCENTER_PASSWORD=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.vcenter_password" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
VCENTER_DISABLE_TLS=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.vcenter_disable_tls_verification" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}' | tr '[:upper:]' '[:lower:]')
EVENT_PROCESSOR_TYPE=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.event_processor_type" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
OPENFAAS_PASSWORD=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.openfaas_password" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
OPENFAAS_ADV_OPTION=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.openfaas_advanced_options" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
AWS_EVENTBRIDGE_ACCESS_KEY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.aws_eb_access_key" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
AWS_EVENTBRIDGE_ACCESS_SECRET=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.aws_eb_access_secret" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
AWS_EVENTBRIDGE_EVENT_BUS=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.aws_eb_event_bus" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
AWS_EVENTBRIDGE_REGION=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.aws_eb_region" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
AWS_EVENTBRIDGE_RULE_ARN=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.aws_eb_arn" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
AWS_EVENTBRIDGE_ADV_OPTION=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.aws_eb_advanced_options" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
DOCKER_NETWORK_CIDR=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.docker_network_cidr" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
POD_NETWORK_CIDR=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.pod_network_cidr" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')

if [ -e /root/ran_customization ]; then
    exit
else
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

	echo -e "\e[92mStarting Customization ..." > /dev/console

	echo -e "\e[92mStarting OS Configuration ..." > /dev/console
	. /root/setup/setup-01-os.sh

	echo -e "\e[92mStarting Network Proxy Configuration ..." > /dev/console
	. /root/setup/setup-02-proxy.sh

	echo -e "\e[92mStarting Network Configuration ..." > /dev/console
	. /root/setup/setup-03-network.sh

	echo -e "\e[92mStarting Kubernetes Configuration ..." > /dev/console
	. /root/setup/setup-04-kubernetes.sh

	echo -e "\e[92mStarting VMware Event Processor Configuration ..." > /dev/console
	. /root/setup/setup-05-event-processor.sh

	echo -e "\e[92mStarting VMware Event Router Configuration ..." > /dev/console
	. /root/setup/setup-06-event-router.sh

	echo -e "\e[92mStarting TinyWWW Configuration ..." > /dev/console
	. /root/setup/setup-07-tinywww.sh

	echo -e "\e[92mStarting Ingress Router Configuration ..." > /dev/console
	. /root/setup/setup-08-ingress.sh

	echo -e "\e[92mStarting OS Banner Configuration ..."> /dev/console
	. /root/setup/setup-09-banner.sh &

	echo -e "\e[92mCustomization Completed ..." > /dev/console

	# Clear guestinfo.ovfEnv
	vmtoolsd --cmd "info-set guestinfo.ovfEnv NULL"

	# Ensure we don't run customization again
	touch /root/ran_customization
fi