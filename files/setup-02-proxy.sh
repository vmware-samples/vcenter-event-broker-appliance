#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Network Proxy for both OS and Docker

set -euo pipefail

if [ -n "${HTTP_PROXY}" ] || [ -n "${HTTPS_PROXY}" ]; then
    PROXY_CONF=/etc/sysconfig/proxy
    DOCKER_PROXY=/etc/systemd/system/docker.service.d

    echo -e "\e[92mConfiguring Proxy ..." > /dev/console
    echo "PROXY_ENABLED=\"yes\"" > ${PROXY_CONF}
    mkdir -p ${DOCKER_PROXY}
    YES_CREDS=0
    if [ -n "${PROXY_USERNAME}" ] && [ -n "${PROXY_PASSWORD}" ]; then
        YES_CREDS=1
    fi

    if [ ! -z "${NO_PROXY}" ]; then
        echo "NO_PROXY=\"${NO_PROXY}\"" >> ${PROXY_CONF}
    fi

    if [ ! -z "${HTTP_PROXY}" ]; then
        if [ $YES_CREDS -eq 1 ]; then
            HTTP_PROXY_URL="http://${PROXY_USERNAME}:${PROXY_PASSWORD}@${HTTP_PROXY}"
        else
            HTTP_PROXY_URL="http://${HTTP_PROXY}"
        fi
        echo "HTTP_PROXY=\"${HTTP_PROXY_URL}\"" >> ${PROXY_CONF}
        cat > ${DOCKER_PROXY}/http-proxy.conf << __HTTP_DOCKER_PROXY__
[Service]
Environment="HTTP_PROXY=${HTTP_PROXY_URL}" "NO_PROXY=${NO_PROXY}"
__HTTP_DOCKER_PROXY__
    fi

    if [ ! -z "${HTTPS_PROXY}" ]; then
        if [ $YES_CREDS -eq 1 ]; then
            HTTPS_PROXY_URL="https://${PROXY_USERNAME}:${PROXY_PASSWORD}@${HTTPS_PROXY}"
        else
            HTTPS_PROXY_URL="https://${HTTPS_PROXY}"
        fi
        echo "HTTPS_PROXY=\"${HTTPS_PROXY_URL}\"" >> ${PROXY_CONF}
        cat > ${DOCKER_PROXY}/https-proxy.conf << __HTTPS_DOCKER_PROXY__
[Service]
Environment="HTTPS_PROXY=${HTTPS_PROXY_URL}" "NO_PROXY=${NO_PROXY}"
__HTTPS_DOCKER_PROXY__
    fi
fi