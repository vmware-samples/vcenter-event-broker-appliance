#!/bin/bash
# Copyright 2020 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail

# Sample Shell Script to test deployment of VEBA w/Knative Processor with the following:
#    vSphere Sources         ✔️
#    Horizon Sources         x
#    vSphere VEBA UI Plugin  ✔️
#    Webhook Processor       ✔️

# To run this on Windows, you can use a terminal like Git BASH or WSL Ubuntu
# OVFTOOL_BIN_PATH="/C/Program Files/VMware/VMware OVF Tool/ovftool.exe"
OVFTOOL_BIN_PATH="/Applications/VMware OVF Tool/ovftool"
VEBA_OVA="../output-vmware-iso/VMware_Event_Broker_Appliance_v0.7.0.ova"

# vCenter
DEPLOYMENT_TARGET_ADDRESS="192.168.30.3"
DEPLOYMENT_TARGET_USERNAME="administrator@vsphere.local"
DEPLOYMENT_TARGET_PASSWORD="VMware1!"
DEPLOYMENT_TARGET_DATACENTER="Primp-Datacenter"
DEPLOYMNET_TARGET_CLUSTER="Supermicro-Cluster"

VEBA_NAME="VEBA-TEST-KNATIVE-PROCESSOR-WITH-VEBA-UI-AND-WEBHOOK"
VEBA_IP="192.168.30.9"
VEBA_HOSTNAME="veba.primp-industries.local"
VEBA_PREFIX="24 (255.255.255.0)"
VEBA_GW="192.168.30.1"
VEBA_DNS="192.168.30.2"
VEBA_DNS_DOMAIN="primp-industries.local"
VEBA_NTP="pool.ntp.org"
VEBA_OS_PASSWORD='VMware1!'
VEBA_ENDPOINT_USERNAME="admin"
VEBA_ENDPOINT_PASSWORD="VMware1!"
VEBA_ENABLE_SSH="True"
VEBA_NETWORK="VM Network"
VEBA_DATASTORE="sm-vsanDatastore"
VEBA_DEBUG="True"
VEBA_VCENTER_SERVER="vcsa.primp-industries.local"
VEBA_UI_USERNAME="veba-ui@vsphere.local"
VEBA_UI_PASSWORD='$VMware1!'
VEBA_VCENTER_USERNAME="veba@vsphere.local"
VEBA_VCENTER_PASSWORD="VMware1!"
VEBA_VCENTER_DISABLE_TLS="True"
VEBA_DOCKER_NETWORK="172.26.0.1/16"
VEBA_HTTP_PROXY=""
VEBA_HTTPS_PROXY=""
VEBA_PROXY_USERNAME=""
VEBA_PROXY_PASSWORD=""
VEBA_NOPROXY=""
VEBA_WEBHOOK="True"
VEBA_WEBHOOK_USERNAME="veba"
VEBA_WEBHOOK_PASSWORD="V3b@"
VEBA_TANZU_SOURCES_DEBUG="False"

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
    --prop:guestinfo.endpoint_username=${VEBA_ENDPOINT_USERNAME} \
    --prop:guestinfo.endpoint_password=${VEBA_ENDPOINT_PASSWORD} \
    --prop:guestinfo.vcenter_server=${VEBA_VCENTER_SERVER} \
    --prop:guestinfo.vcenter_username=${VEBA_VCENTER_USERNAME} \
    --prop:guestinfo.vcenter_password=${VEBA_VCENTER_PASSWORD} \
    --prop:guestinfo.vcenter_veba_ui_username=${VEBA_UI_USERNAME} \
    --prop:guestinfo.vcenter_veba_ui_password=${VEBA_UI_PASSWORD} \
    --prop:guestinfo.vcenter_disable_tls_verification=${VEBA_VCENTER_DISABLE_TLS} \
    --prop:guestinfo.webhook=${VEBA_WEBHOOK} \
    --prop:guestinfo.webhook_username=${VEBA_WEBHOOK_USERNAME} \
    --prop:guestinfo.webhook_password=${VEBA_WEBHOOK_PASSWORD} \
    --prop:guestinfo.debug=${VEBA_DEBUG} \
    --prop:guestinfo.tanzu_sources_debug=${VEBA_TANZU_SOURCES_DEBUG} \
    "${VEBA_OVA}" \
    "vi://${DEPLOYMENT_TARGET_USERNAME}:${DEPLOYMENT_TARGET_PASSWORD}@${DEPLOYMENT_TARGET_ADDRESS}/${DEPLOYMENT_TARGET_DATACENTER}/host/${DEPLOYMNET_TARGET_CLUSTER}"
