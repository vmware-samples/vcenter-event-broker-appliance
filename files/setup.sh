#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail

# Extract all OVF Properties
VEBA_DEBUG=$(/root/setup/getOvfProperty.py "guestinfo.debug")
HOSTNAME=$(/root/setup/getOvfProperty.py "guestinfo.hostname")
IP_ADDRESS=$(/root/setup/getOvfProperty.py "guestinfo.ipaddress")
NETMASK=$(/root/setup/getOvfProperty.py "guestinfo.netmask" | awk -F ' ' '{print $1}')
GATEWAY=$(/root/setup/getOvfProperty.py "guestinfo.gateway")
DNS_SERVER=$(/root/setup/getOvfProperty.py "guestinfo.dns")
DNS_DOMAIN=$(/root/setup/getOvfProperty.py "guestinfo.domain")
NTP_SERVER=$(/root/setup/getOvfProperty.py "guestinfo.ntp")
HTTP_PROXY=$(/root/setup/getOvfProperty.py "guestinfo.http_proxy")
HTTPS_PROXY=$(/root/setup/getOvfProperty.py "guestinfo.https_proxy")
PROXY_USERNAME=$(/root/setup/getOvfProperty.py "guestinfo.proxy_username")
PROXY_PASSWORD=$(/root/setup/getOvfProperty.py "guestinfo.proxy_password")
NO_PROXY=$(/root/setup/getOvfProperty.py "guestinfo.no_proxy")
ROOT_PASSWORD=$(/root/setup/getOvfProperty.py "guestinfo.root_password")
ENABLE_SSH=$(/root/setup/getOvfProperty.py "guestinfo.enable_ssh" | tr '[:upper:]' '[:lower:]')
VCENTER_SERVER=$(/root/setup/getOvfProperty.py "guestinfo.vcenter_server")
VCENTER_USERNAME=$(/root/setup/getOvfProperty.py "guestinfo.vcenter_username")
VCENTER_PASSWORD=$(/root/setup/getOvfProperty.py "guestinfo.vcenter_password")
VCENTER_USERNAME_FOR_VEBA_UI=$(/root/setup/getOvfProperty.py "guestinfo.vcenter_veba_ui_username")
VCENTER_PASSWORD_FOR_VEBA_UI=$(/root/setup/getOvfProperty.py "guestinfo.vcenter_veba_ui_password")
VCENTER_DISABLE_TLS=$(/root/setup/getOvfProperty.py "guestinfo.vcenter_disable_tls_verification")
EVENT_PROCESSOR_TYPE=$(/root/setup/getOvfProperty.py "guestinfo.event_processor_type")
OPENFAAS_PASSWORD=$(/root/setup/getOvfProperty.py "guestinfo.openfaas_password")
OPENFAAS_ADV_OPTION=$(/root/setup/getOvfProperty.py "guestinfo.openfaas_advanced_options")
KNATIVE_HOST=$(/root/setup/getOvfProperty.py "guestinfo.knative_host")
KNATIVE_SCHEME=$(/root/setup/getOvfProperty.py "guestinfo.knative_scheme" | tr [:upper:] [:lower:])
KNATIVE_DISABLE_TLS=$(/root/setup/getOvfProperty.py "guestinfo.knative_disable_tls_verification")
KNATIVE_PATH=$(/root/setup/getOvfProperty.py "guestinfo.knative_path")
AWS_EVENTBRIDGE_ACCESS_KEY=$(/root/setup/getOvfProperty.py "guestinfo.aws_eb_access_key")
AWS_EVENTBRIDGE_ACCESS_SECRET=$(/root/setup/getOvfProperty.py "guestinfo.aws_eb_access_secret")
AWS_EVENTBRIDGE_EVENT_BUS=$(/root/setup/getOvfProperty.py "guestinfo.aws_eb_event_bus")
AWS_EVENTBRIDGE_REGION=$(/root/setup/getOvfProperty.py "guestinfo.aws_eb_region")
AWS_EVENTBRIDGE_RULE_ARN=$(/root/setup/getOvfProperty.py "guestinfo.aws_eb_arn")
AWS_EVENTBRIDGE_ADV_OPTION=$(/root/setup/getOvfProperty.py "guestinfo.aws_eb_advanced_options")
CUSTOM_VEBA_TLS_PRIVATE_KEY=$(/root/setup/getOvfProperty.py "guestinfo.custom_tls_private_key")
CUSTOM_VEBA_TLS_CA_CERT=$(/root/setup/getOvfProperty.py "guestinfo.custom_tls_ca_cert")
DOCKER_NETWORK_CIDR=$(/root/setup/getOvfProperty.py "guestinfo.docker_network_cidr")
POD_NETWORK_CIDR=$(/root/setup/getOvfProperty.py "guestinfo.pod_network_cidr")
SYSLOG_SERVER_HOSTNAME=$(/root/setup/getOvfProperty.py "guestinfo.syslog_server_hostname")
SYSLOG_SERVER_PORT=$(/root/setup/getOvfProperty.py "guestinfo.syslog_server_port")
SYSLOG_SERVER_PROTOCOL=$(/root/setup/getOvfProperty.py "guestinfo.syslog_server_protocol")
SYSLOG_SERVER_FORMAT=$(/root/setup/getOvfProperty.py "guestinfo.syslog_server_format")
LOCAL_STORAGE_DISK="/dev/sdb"
LOCAL_STOARGE_VOLUME_PATH="/data"
export KUBECONFIG="/root/.kube/config"

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

	# Determine Knative deployment model
	if [ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]; then
		if [ ! -z ${KNATIVE_HOST} ]; then
			KNATIVE_DEPLOYMENT_TYPE="external"
		else
			KNATIVE_DEPLOYMENT_TYPE="embedded"
		fi
	else
		KNATIVE_DEPLOYMENT_TYPE="na"
	fi

	# Customize the POD CIDR Network if provided or else default to 10.10.0.0/16
	if [ -z "${POD_NETWORK_CIDR}" ]; then
		POD_NETWORK_CIDR="10.16.0.0/16"
	fi

	# Slicing of escaped variables needed to properly handle the double quotation issue
	ESCAPED_VCENTER_SERVER=$(eval echo -n ${VCENTER_SERVER} | jq -Rs .)
	ESCAPED_VCENTER_USERNAME=$(eval echo -n ${VCENTER_USERNAME} | jq -Rs .)
	ESCAPED_VCENTER_PASSWORD=$(eval echo -n ${VCENTER_PASSWORD} | jq -Rs .)
	ESCAPED_ROOT_PASSWORD=$(eval echo -n ${ROOT_PASSWORD} | jq -Rs .)

	ESCAPED_VCENTER_USERNAME_FOR_VEBA_UI=$(eval echo -n ${VCENTER_USERNAME_FOR_VEBA_UI} | jq -Rs .)
	ESCAPED_VCENTER_PASSWORD_FOR_VEBA_UI=$(eval echo -n ${VCENTER_PASSWORD_FOR_VEBA_UI} | jq -Rs .)

	ESCAPED_AWS_EVENTBRIDGE_ACCESS_KEY=$(eval echo -n ${AWS_EVENTBRIDGE_ACCESS_KEY} | jq -Rs .)
	ESCAPED_AWS_EVENTBRIDGE_ACCESS_SECRET=$(eval echo -n ${AWS_EVENTBRIDGE_ACCESS_SECRET} | jq -Rs .)
	ESCAPED_AWS_EVENTBRIDGE_EVENT_BUS=$(eval echo -n ${AWS_EVENTBRIDGE_EVENT_BUS} | jq -Rs .)
	ESCAPED_AWS_EVENTBRIDGE_RULE_ARN=$(eval echo -n ${AWS_EVENTBRIDGE_RULE_ARN} | jq -Rs .)

	ESCAPED_OPENFAAS_PASSWORD=$(eval echo -n ${OPENFAAS_PASSWORD} | jq -Rs .)

	cat > /root/config/veba-config.json <<EOF
{
	"VEBA_DEBUG": "${VEBA_DEBUG}",
	"HOSTNAME": "${HOSTNAME}",
	"IP_ADDRESS": "${IP_ADDRESS}",
	"NETMASK": "${NETMASK}",
	"GATEWAY": "${GATEWAY}",
	"DNS_SERVER": "${DNS_SERVER}",
	"DNS_DOMAIN": "${DNS_DOMAIN}",
	"NTP_SERVER": "${NTP_SERVER}",
	"HTTP_PROXY": "${HTTP_PROXY}",
	"HTTPS_PROXY": "${HTTPS_PROXY}",
	"PROXY_USERNAME": "${PROXY_USERNAME}",
	"PROXY_PASSWORD": "${PROXY_PASSWORD}",
	"NO_PROXY": "${NO_PROXY}",
	"ESCAPED_ROOT_PASSWORD": ${ESCAPED_ROOT_PASSWORD},
	"ENABLE_SSH": "${ENABLE_SSH}",
	"ESCAPED_VCENTER_SERVER": ${ESCAPED_VCENTER_SERVER},
	"ESCAPED_VCENTER_USERNAME": ${ESCAPED_VCENTER_USERNAME},
	"ESCAPED_VCENTER_PASSWORD": ${ESCAPED_VCENTER_PASSWORD},
	"ESCAPED_VCENTER_USERNAME_FOR_VEBA_UI": ${ESCAPED_VCENTER_USERNAME_FOR_VEBA_UI},
	"ESCAPED_VCENTER_PASSWORD_FOR_VEBA_UI": ${ESCAPED_VCENTER_PASSWORD_FOR_VEBA_UI},
	"VCENTER_DISABLE_TLS": "${VCENTER_DISABLE_TLS}",
	"EVENT_PROCESSOR_TYPE": "${EVENT_PROCESSOR_TYPE}",
	"KNATIVE_DEPLOYMENT_TYPE": "${KNATIVE_DEPLOYMENT_TYPE}",
	"ESCAPED_OPENFAAS_PASSWORD": ${ESCAPED_OPENFAAS_PASSWORD},
	"OPENFAAS_ADV_OPTION": "${OPENFAAS_ADV_OPTION}",
	"KNATIVE_HOST": "${KNATIVE_HOST}",
	"KNATIVE_SCHEME": "${KNATIVE_SCHEME}",
	"KNATIVE_DISABLE_TLS": "${KNATIVE_DISABLE_TLS}",
	"KNATIVE_PATH": "${KNATIVE_PATH}",
	"ESCAPED_AWS_EVENTBRIDGE_ACCESS_KEY": ${ESCAPED_AWS_EVENTBRIDGE_ACCESS_KEY},
	"ESCAPED_AWS_EVENTBRIDGE_ACCESS_SECRET": ${ESCAPED_AWS_EVENTBRIDGE_ACCESS_SECRET},
	"ESCAPED_AWS_EVENTBRIDGE_EVENT_BUS": ${ESCAPED_AWS_EVENTBRIDGE_EVENT_BUS},
	"AWS_EVENTBRIDGE_REGION": "${AWS_EVENTBRIDGE_REGION}",
	"ESCAPED_AWS_EVENTBRIDGE_RULE_ARN": ${ESCAPED_AWS_EVENTBRIDGE_RULE_ARN},
	"AWS_EVENTBRIDGE_ADV_OPTION": "${AWS_EVENTBRIDGE_ADV_OPTION}",
	"CUSTOM_VEBA_TLS_PRIVATE_KEY": "${CUSTOM_VEBA_TLS_PRIVATE_KEY}",
	"CUSTOM_VEBA_TLS_CA_CERT": "${CUSTOM_VEBA_TLS_CA_CERT}",
	"DOCKER_NETWORK_CIDR": "${DOCKER_NETWORK_CIDR}",
	"POD_NETWORK_CIDR": "${POD_NETWORK_CIDR}",
	"SYSLOG_SERVER_HOSTNAME": "${SYSLOG_SERVER_HOSTNAME}",
	"SYSLOG_SERVER_PORT": "${SYSLOG_SERVER_PORT}",
	"SYSLOG_SERVER_PROTOCOL": "${SYSLOG_SERVER_PROTOCOL}",
	"SYSLOG_SERVER_FORMAT": "${SYSLOG_SERVER_FORMAT}"
}
EOF

	echo -e "\e[92mStarting Customization ..." > /dev/console

	echo -e "\e[92mStarting OS Configuration ..." > /dev/console
	. /root/setup/setup-01-os.sh

	echo -e "\e[92mStarting Network Proxy Configuration ..." > /dev/console
	. /root/setup/setup-02-proxy.sh

	echo -e "\e[92mStarting Network Configuration ..." > /dev/console
	. /root/setup/setup-03-network.sh

	echo -e "\e[92mStarting Kubernetes Configuration ..." > /dev/console
	. /root/setup/setup-04-kubernetes.sh

	if [ "${KNATIVE_DEPLOYMENT_TYPE}" == "embedded" ]; then
		echo -e "\e[92mStarting Knative Configuration ..." > /dev/console
		. /root/setup/setup-05-knative.sh
	fi

	echo -e "\e[92mStarting VMware Event Processor Configuration ..." > /dev/console
	. /root/setup/setup-06-event-processor.sh

	echo -e "\e[92mStarting VMware Event Router Configuration ..." > /dev/console
	. /root/setup/setup-07-event-router.sh

	echo -e "\e[92mStarting TinyWWW Configuration ..." > /dev/console
	. /root/setup/setup-08-tinywww.sh

	echo -e "\e[92mStarting Ingress Router Configuration ..." > /dev/console
	. /root/setup/setup-09-ingress.sh

	if [[ "${KNATIVE_DEPLOYMENT_TYPE}" == "embedded" ]] && [[ ! -z ${VCENTER_USERNAME_FOR_VEBA_UI} ]] && [[ ! -z ${VCENTER_PASSWORD_FOR_VEBA_UI} ]]; then
		echo -e "\e[92mStarting Knative UI Configuration ..." > /dev/console
		. /root/setup/setup-010-veba-ui.sh
	fi

	if [ -n "${SYSLOG_SERVER_HOSTNAME}" ]; then
		echo -e "\e[92mStarting FluentBit Configuration ..." > /dev/console
		. /root/setup/setup-011-fluentbit.sh
	fi

	echo -e "\e[92mStarting cAdvisor Configuration ..." > /dev/console
	. /root/setup/setup-012-cadvisor.sh

	echo -e "\e[92mStarting OS Banner Configuration ..."> /dev/console
	. /root/setup/setup-099-banner.sh &

	echo -e "\e[92mCustomization Completed ..." > /dev/console

	# Clear guestinfo.ovfEnv
	vmtoolsd --cmd "info-set guestinfo.ovfEnv NULL"

	# Ensure we don't run customization again
	touch /root/ran_customization
fi