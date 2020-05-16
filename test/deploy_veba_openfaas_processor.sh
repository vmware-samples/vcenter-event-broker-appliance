#!/bin/bash
# Copyright 2020 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Sample Shell Script to test deployment of VEBA w/OpenFaaS Processor

OVFTOOL_BIN_PATH="/Applications/VMware OVF Tool/ovftool"
VEBA_OVA="../output-vmware-iso/vCenter_Event_Broker_Appliance_0.4.0-beta.ova"

# vCenter
DEPLOYMENT_TARGET_ADDRESS="192.168.30.200"
DEPLOYMENT_TARGET_USERNAME="administrator@vsphere.local"
DEPLOYMENT_TARGET_PASSWORD="VMware1!"
DEPLOYMENT_TARGET_DATACENTER="Primp-Datacenter"
DEPLOYMNET_TARGET_CLUSTER="Supermicro-Cluster"

VEBA_NAME="VEBA-TEST-OPENFAAS-PROCESSOR"
VEBA_IP="192.168.130.170"
VEBA_HOSTNAME="veba.primp-industries.com"
VEBA_PREFIX="24 (255.255.255.0)"
VEBA_GW="192.168.30.1"
VEBA_DNS="192.168.30.1"
VEBA_DNS_DOMAIN="primp-industries.com"
VEBA_NTP="pool.ntp.org"
VEBA_OS_PASSWORD="VMware1!"
VEBA_ENABLE_SSH="True"
VEBA_NETWORK="VM Network"
VEBA_DATASTORE="sm-vsanDatastore"
VEBA_DEBUG="True"
VEBA_VCENTER_SERVER="192.168.30.200"
VEBA_VCENTER_USER="administrator@vsphere.local"
VEBA_VCENTER_PASS="VMware1!"
VEBA_VCENTER_DISABLE_TLS="True"
VEBA_OPENFAAS_PASS="VMware1!"
VEBA_DOCKER_NETWORK="172.26.0.1/16"
VEBA_HTTP_PROXY=""
VEBA_HTTPS_PROXY=""
VEBA_PROXY_USERNAME=""
VEBA_PROXY_PASSWORD=""
VEBA_NOPROXY=""

### DO NOT EDIT BEYOND HERE ###

"${OVFTOOL_BIN_PATH}" \
    --powerOn \
    --noSSLVerify \
    --sourceType=OVA \
    --allowExtraConfig \
    --diskMode=thin \
    --name="${VEBA_NAME}" \
    --net:"VM Network"="${VEBA_NETWORK}" \
    --datastore="${VEBA_DATASTORE}" \
    --prop:guestinfo.ipaddress=${VEBA_IP} \
    --prop:guestinfo.hostname=${VEBA_HOSTNAME} \
    --prop:guestinfo.netmask="${VEBA_PREFIX}" \
    --prop:guestinfo.gateway=${VEBA_GW} \
    --prop:guestinfo.dns=${VEBA_DNS} \
    --prop:guestinfo.domain=${VEBA_DNS_DOMAIN} \
    --prop:guestinfo.ntp=${VEBA_NTP} \
    --prop:guestinfo.http_proxy=${VEBA_HTTP_PROXY} \
    --prop:guestinfo.https_proxy=${VEBA_HTTPS_PROXY} \
    --prop:guestinfo.proxy_username=${VEBA_PROXY_USERNAME} \
    --prop:guestinfo.proxy_password=${VEBA_PROXY_PASSWORD} \
    --prop:guestinfo.no_proxy=${VEBA_NOPROXY} \
    --prop:guestinfo.root_password=${VEBA_OS_PASSWORD} \
    --prop:guestinfo.enable_ssh=${VEBA_ENABLE_SSH} \
    --prop:guestinfo.vcenter_server=${VEBA_VCENTER_SERVER} \
    --prop:guestinfo.vcenter_username=${VEBA_VCENTER_USER} \
    --prop:guestinfo.vcenter_password=${VEBA_VCENTER_PASS} \
    --prop:guestinfo.vcenter_disable_tls_verification=${VEBA_VCENTER_DISABLE_TLS} \
    --prop:guestinfo.event_processor_type="OpenFaaS" \
    --prop:guestinfo.openfaas_password=${VEBA_OPENFAAS_PASS} \
    --prop:guestinfo.debug=${VEBA_DEBUG} \
    --prop:guestinfo.docker_network_cidr=${VEBA_DOCKER_NETWORK} \
    "${VEBA_OVA}" \
    "vi://${DEPLOYMENT_TARGET_USERNAME}:${DEPLOYMENT_TARGET_PASSWORD}@${DEPLOYMENT_TARGET_ADDRESS}/${DEPLOYMENT_TARGET_DATACENTER}/host/${DEPLOYMNET_TARGET_CLUSTER}"
