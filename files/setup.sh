#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Bootstrap script to setup k8s, OpenFaaS & vCenter Connector

set -euo pipefail

if [ -e /root/ran_customization ]; then
    exit
else
    NETWORK_CONFIG_FILE=$(ls /etc/systemd/network | grep .network)

    VEBA_DEBUG_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.debug")
    VEBA_DEBUG=$(echo "${VEBA_DEBUG_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    VEBA_LOG_FILE=/var/log/bootstrap.log
    if [ ${VEBA_DEBUG} == "True" ]; then
        VEBA_LOG_FILE=/var/log/bootstrap-debug.log
        set -x
        exec 2> ${VEBA_LOG_FILE}
        echo
        echo "### WARNING -- DEBUG LOG CONTAINS ALL EXECUTED COMMANDS WHICH INCLUDES CREDENTIALS -- WARNING ###"
        echo "### WARNING --             PLEASE REMOVE CREDENTIALS BEFORE SHARING LOG            -- WARNING ###"
        echo
    fi

    HOSTNAME_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.hostname")
    IP_ADDRESS_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.ipaddress")
    NETMASK_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.netmask")
    GATEWAY_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.gateway")
    DNS_SERVER_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.dns")
    DNS_DOMAIN_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.domain")
    NTP_SERVER_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.ntp")
    ROOT_PASSWORD_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.root_password")
    OPENFAAS_PASSWORD_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.openfaas_password")
    VCENTER_SERVER_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.vcenter_server")
    VCENTER_USERNAME_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.vcenter_username")
    VCENTER_PASSWORD_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.vcenter_password")
    VCENTER_DISABLE_TLS_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.vcenter_disable_tls_verification")
    POD_NETWORK_CIDR_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep "guestinfo.pod_network_cidr")

    ##################################
    ### No User Input, assume DHCP ###
    ##################################
    if [ -z "${HOSTNAME_PROPERTY}" ]; then
        cat > /etc/systemd/network/${NETWORK_CONFIG_FILE} << __CUSTOMIZE_PHOTON__
[Match]
Name=e*

[Network]
DHCP=yes
IPv6AcceptRA=no
__CUSTOMIZE_PHOTON__
    #########################
    ### Static IP Address ###
    #########################
    else
        HOSTNAME=$(echo "${HOSTNAME_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
        IP_ADDRESS=$(echo "${IP_ADDRESS_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
        NETMASK=$(echo "${NETMASK_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
        GATEWAY=$(echo "${GATEWAY_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
        DNS_SERVER=$(echo "${DNS_SERVER_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
        DNS_DOMAIN=$(echo "${DNS_DOMAIN_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')

        echo -e "\e[92mConfiguring Static IP Address ..." > /dev/console
        cat > /etc/systemd/network/${NETWORK_CONFIG_FILE} << __CUSTOMIZE_PHOTON__
[Match]
Name=e*

[Network]
Address=${IP_ADDRESS}/${NETMASK}
Gateway=${GATEWAY}
DNS=${DNS_SERVER}
Domain=${DNS_DOMAIN}
__CUSTOMIZE_PHOTON__
    #########################
    ### NTP Settings      ###
    #########################
    NTP_SERVER=$(echo "${NTP_SERVER_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')

    echo -e "\e[92mConfiguring NTP ..." > /dev/console
    cat > /etc/systemd/timesyncd.conf << __CUSTOMIZE_PHOTON__

[Match]
Name=e*

[Time]
NTP=${NTP_SERVER}
__CUSTOMIZE_PHOTON__

    echo -e "\e[92mConfiguring hostname ..." > /dev/console
    hostnamectl set-hostname ${HOSTNAME}
    echo "${IP_ADDRESS} ${HOSTNAME}" >> /etc/hosts
    echo -e "\e[92mRestarting Network ..." > /dev/console
    systemctl restart systemd-networkd
    echo -e "\e[92mRestarting Timesync ..." > /dev/console
    systemctl restart systemd-timesyncd
    fi

    echo -e "\e[92mConfiguring root password ..." > /dev/console
    ROOT_PASSWORD=$(echo "${ROOT_PASSWORD_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    echo "root:${ROOT_PASSWORD}" | /usr/sbin/chpasswd

    echo -e "\e[92mRetrieving vSphere & OpenFaaS Variables ..." > /dev/console
    OPENFAAS_PASSWORD=$(echo "${OPENFAAS_PASSWORD_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    VCENTER_SERVER=$(echo "${VCENTER_SERVER_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    VCENTER_USERNAME=$(echo "${VCENTER_USERNAME_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    VCENTER_PASSWORD=$(echo "${VCENTER_PASSWORD_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    VCENTER_DISABLE_TLS=$(echo "${VCENTER_DISABLE_TLS_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    POD_NETWORK_CIDR=$(echo "${POD_NETWORK_CIDR_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')

    echo -e "\e[92mStarting Docker ..." > /dev/console
    systemctl start docker.service
    systemctl enable docker.service

    echo -e "\e[92mDisabling/Stopping IP Tables  ..." > /dev/console
    systemctl stop iptables
    systemctl disable iptables

    # Setup k8s
    echo -e "\e[92mSetting up k8s ..." > /dev/console
    HOME=/root
    kubeadm init --ignore-preflight-errors SystemVerification --skip-token-print --config /root/kubeconfig.yml
    mkdir -p $HOME/.kube
    cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
    chown $(id -u):$(id -g) $HOME/.kube/config
    echo -e "\e[92mDeloying kubeadm ..." > /dev/console

    # Customize the POD CIDR Network if provided or else default to 10.99.0.0/20
    if [ -z "${POD_NETWORK_CIDR}" ]; then
      POD_NETWORK_CIDR="10.99.0.0/20"
    fi

    sed -i "s#POD_NETWORK_CIDR#${POD_NETWORK_CIDR}#g" /root/weave.yaml

    kubectl --kubeconfig /root/.kube/config apply -f /root/weave.yaml
    kubectl --kubeconfig /root/.kube/config taint nodes --all node-role.kubernetes.io/master-
    echo -e "\e[92mStarting k8s ..." > /dev/console
    systemctl enable kubelet.service

    while [[ $(systemctl is-active kubelet.service) == "inactive" ]]
    do
        echo -e "\e[92mk8s service is still inactive, sleeping for 10secs" > /dev/console
        sleep 10
    done

    # Setup Contour
    echo -e "\e[92mDeploying Contour ..." > /dev/console
    kubectl --kubeconfig /root/.kube/config create -f /root/download/contour/examples/contour/

    # Setup OpenFaaS
    echo -e "\e[92mDeploying OpenFaas ..." > /dev/console
    kubectl --kubeconfig /root/.kube/config create -f /root/download/faas-netes/namespaces.yml

    # Setup OpenFaaS Secret
    kubectl --kubeconfig /root/.kube/config -n openfaas create secret generic basic-auth \
        --from-literal=basic-auth-user=admin \
        --from-literal=basic-auth-password="${OPENFAAS_PASSWORD}"

    kubectl --kubeconfig /root/.kube/config create -f /root/download/faas-netes/yaml

    ## Create SSL Certificate & Secret
    KEY_FILE=/root/openfaas-gw.key
    CERT_FILE=/root/openfaas-gw.crt
    CN_NAME=$(hostname)
    CERT_NAME=openfaas-gw-tls

    openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ${KEY_FILE} -out ${CERT_FILE} -subj "/CN=${CN_NAME}/O=${CN_NAME}"

    kubectl --kubeconfig /root/.kube/config -n openfaas create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}

    # Deploy Ingress Route Gateway
    cat << EOF > /root/ingressroute-gateway.yaml
apiVersion: contour.heptio.com/v1beta1
kind: IngressRoute
metadata:
  labels:
    app: openfaas
  name: ingressroute-gateway
  namespace: openfaas
spec:
  virtualhost:
    fqdn: ${HOSTNAME}
    tls:
      secretName: ${CERT_NAME}
      minimumProtocolVersion: "1.2"
  routes:
    - match: /status
      prefixRewrite: /status
      services:
      - name: tinywww
        port: 8100
    - match: /bootstrap
      prefixRewrite: /bootstrap
      services:
      - name: tinywww
        port: 8100
    - match: /
      services:
      - name: gateway
        port: 8080
EOF

    kubectl --kubeconfig /root/.kube/config create -f /root/ingressroute-gateway.yaml

    # Setup OpenFaaS vCenter Connector
    echo -e "\e[92mSetting up vCenter Connector ..." > /dev/console
    sed -i "s/http:\/\/vcsim.openfaas:8989/${VCENTER_SERVER}/g" /root/download/vcenter-connector/yaml/kubernetes/connector-dep.yml

    # Enable TLS verification for vCenter Server connection by default unless user specifies otherwise
    if [ ${VCENTER_DISABLE_TLS} != "True" ] ;then
      sed -i 's/"-insecure", //g' /root/download/vcenter-connector/yaml/kubernetes/connector-dep.yml
    fi

    # Setup OpenFaaS vCenter Connector Secrets
    kubectl --kubeconfig /root/.kube/config create secret generic vcenter-secrets \
        -n openfaas \
        --from-literal vcenter-username=${VCENTER_USERNAME} \
        --from-literal vcenter-password=${VCENTER_PASSWORD}

    echo -e "\e[92mDeploying vCenter Connector ..." > /dev/console
    kubectl --kubeconfig /root/.kube/config -n openfaas create -f /root/download/vcenter-connector/yaml/kubernetes/connector-dep.yml

    # Deploy TinyWWW Pod
    if [ ${VEBA_DEBUG} == "True" ]; then
      kubectl --kubeconfig /root/.kube/config apply -f /root/tinywww-debug.yml
    else
      kubectl --kubeconfig /root/.kube/config apply -f /root/tinywww.yml
    fi

    # Ensure we don't run customization again
    touch /root/ran_customization

    # Update /etc/issue with IP Address
    echo -e "\e[92mUpdating the Login Banner ..." > /dev/console
    /root/setup-banner.sh &

    # Disabling SSH
    systemctl disable sshd
    systemctl stop sshd
fi