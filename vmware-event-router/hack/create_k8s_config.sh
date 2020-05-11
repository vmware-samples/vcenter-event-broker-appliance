#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail

black='\E[30;40m'
red='\E[31;40m'
green='\E[32;40m'
magenta='\E[35;40m'
cyan='\E[36;40m'
reset='\033[00m'

cecho () {
    local default_msg="No message passed."

    message=${1:-$default_msg}
    color=${2:-$black}

    echo -e "$color"
    echo -e "$message"
    tput sgr0

    return
}

cecho "Enter the following values for deployment:" $cyan

echo -e "${magenta}"
read -p "vCenter Server FQDN: " VCENTER_SERVER
if [[ -z "${VCENTER_SERVER}" ]]; then
    cecho "Please start over. No input entered." $red
    exit 1
fi

echo -e "${magenta}"
read -p "vCenter Server Username: " VCENTER_USERNAME
if [[ -z "${VCENTER_USERNAME}" ]]; then
    cecho "Please start over. No input entered." $red
    exit 1
fi

echo -e "${magenta}"
echo -n "vCenter Server Password: "
read -s VCENTER_PASSWORD
if [[ -z "${VCENTER_PASSWORD}" ]]; then
    cecho "Please start over. No input entered." $red
    exit 1
fi
echo ""

echo -e "${magenta}"
read -p "Deploy OpenFaaS: [y|n] " DEPLOY_OPENFAAS
if [[ -z "${DEPLOY_OPENFAAS}" ]]; then
    cecho "Please start over. No input entered." $red
    exit 1
fi

if [ ${DEPLOY_OPENFAAS} == "y" ]; then
    echo -e "${magenta}"
    echo -n "OpenFaaS Admin Password: "
    read -s OPENFAAS_PASSWORD
    if [[ -z "${OPENFAAS_PASSWORD}" ]]; then
        cecho "Please start over. No input entered." $red
        exit 1
    fi
fi
echo ""

cecho "Please confirm the following settings are correct:\n" $cyan
echo -e "\tVCENTER_SERVER=${VCENTER_SERVER}"
echo -e "\tVCENTER_USERNAME=${VCENTER_USERNAME}"
echo -e "\tVCENTER_PASSWORD=${VCENTER_PASSWORD}"
echo -e "\tDEPLOY_OPENFAAS=${DEPLOY_OPENFAAS}"
if [ ${DEPLOY_OPENFAAS} == "y" ]; then
    echo -e "\tOPENFAAS_PASSWORD=${OPENFAAS_PASSWORD}"
fi

echo -e "${cyan}"
read -p "Do you want to proceed with the VEBA K8s deployment [y]: " answer
case $answer in
    [Yy]* ) cecho "Starting Deployment ..." $green;;
    [Nn]* ) cecho "Exiting ..." $red; exit;;
    * ) cecho "Exiting ..." $red; exit;;
esac

cecho "Checking for K8s namespace creation permission ..." $green
kubectl auth can-i create ns -q
if [ $? -eq 1 ]; then
  cecho "You do not have permission to create a new K8s namespace" $red
  exit 1
fi

cecho "Checking for K8s deployments creation permission ..."  $green
kubectl auth can-i create deployments -q
if [ $? -eq 1 ]; then
  cecho "You do not have permission to create a new K8s deployments" $red
  exit 1
fi

cecho "Checking for K8s secrets creation permission ..."  $green
kubectl auth can-i create secrets -q
if [ $? -eq 1 ]; then
  cecho "You do not have permission to create a new K8s secrets" $red
  exit 1
fi

cecho "Creating vmware namespace ..." $green
echo -e "\tkubectl create namespace vmware\n"
kubectl create namespace vmware

if [ $DEPLOY_OPENFAAS == "y" ]; then
    cecho "Deploying OpenFaaS ..." $green
    echo -e "\tkubectl create -f faas-netes/namespaces.yml"
    echo -e "\tkubectl -n openfaas create secret generic basic-auth --from-literal=basic-auth-user=admin --from-literal=basic-auth-password=${OPENFAAS_PASSWORD}"
    echo -e "\tkubectl create -f faas-netes/yaml"
    kubectl create -f faas-netes/namespaces.yml
    kubectl -n openfaas create secret generic basic-auth --from-literal=basic-auth-user=admin --from-literal=basic-auth-password="${OPENFAAS_PASSWORD}"
    kubectl create -f faas-netes/yaml

    OPENFAAS_GATEWAY=$(kubectl -n openfaas describe pods $(kubectl -n openfaas get pods | grep "gateway-" | awk '{print $1}') | grep "^Node:" | awk -F "/" '{print $2}')
    while [ 1 ];
    do
        cecho "Waiting for OpenFaaS to be ready ..." $green
        OPENFAAS_GATEWAY=$(kubectl -n openfaas describe pods $(kubectl -n openfaas get pods | grep "gateway-" | awk '{print $1}') | grep "^Node:" | awk -F "/" '{print $2}')
        if [ ! -z ${OPENFAAS_GATEWAY} ]; then
            break
        fi
    done
fi

cecho "Creating VEBA deployment files..." $green
cat > event-router-config.json << EOF
[{
        "type": "stream",
        "provider": "vmware_vcenter",
        "address": "https://${VCENTER_SERVER}:443/sdk",
        "auth": {
            "method": "user_password",
            "secret": {
                "username": "${VCENTER_USERNAME}",
                "password": "${VCENTER_PASSWORD}"
            }
        },
        "options": {
            "insecure": "true"
        }
    },
    {
        "type": "processor",
        "provider": "openfaas",
        "address": "http://gateway.openfaas:8080",
        "auth": {
            "method": "basic_auth",
            "secret": {
                "username": "admin",
                "password": "${OPENFAAS_PASSWORD}"
            }
        },
        "options": {
            "async": "false"
        }
    },
    {
        "type": "metrics",
        "provider": "internal",
        "address": "0.0.0.0:8080",
        "auth": {
            "method": "none"
        }
    }
]
EOF

cecho "Deploying VMware Event Router ..." $green
echo -e "\tkubectl -n vmware create secret generic event-router-config --from-file=event-router-config.json"
echo -e "\tkubectl -n vmware create -f https://raw.githubusercontent.com/vmware-samples/vcenter-event-broker-appliance/master/vmware-event-router/deploy/event-router-k8s.yaml"
kubectl -n vmware create secret generic event-router-config --from-file=event-router-config.json
kubectl -n vmware create -f https://raw.githubusercontent.com/vmware-samples/vcenter-event-broker-appliance/master/vmware-event-router/deploy/event-router-k8s.yaml
