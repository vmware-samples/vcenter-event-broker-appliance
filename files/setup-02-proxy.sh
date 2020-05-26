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
        a=($(printf '%s\n' "${HTTP_PROXY//\:\/\//$'\n'}"))
        if [ ${#a[*]} -eq 2 ]; then
            HTTP_PROXY_PROTOCOL=${a[0]}
            HTTP_PROXY_SERVER_PORT=${a[1]}
            if [ $YES_CREDS -eq 1 ]; then
                HTTP_PROXY_URL="${HTTP_PROXY_PROTOCOL}://${PROXY_USERNAME}:${PROXY_PASSWORD}@${HTTP_PROXY_SERVER_PORT}"
            else
                HTTP_PROXY_URL="${HTTP_PROXY_PROTOCOL}://${HTTP_PROXY_SERVER_PORT}"
            fi
            echo "HTTP_PROXY=\"${HTTP_PROXY_URL}\"" >> ${PROXY_CONF}
            cat > ${DOCKER_PROXY}/http-proxy.conf << __HTTP_DOCKER_PROXY__
[Service]
Environment="HTTP_PROXY=${HTTP_PROXY_URL}" "NO_PROXY=${NO_PROXY}"
__HTTP_DOCKER_PROXY__
        else
	    echo -e "\e[91mInvalid HTTP Proxy URL supplied" > /dev/console
        fi
    fi

    if [ ! -z "${HTTPS_PROXY}" ]; then
        a=($(printf '%s\n' "${HTTPS_PROXY//\:\/\//$'\n'}"))
        if [ ${#a[*]} -eq 2 ]; then
            HTTPS_PROXY_PROTOCOL=${a[0]}
            HTTPS_PROXY_SERVER_PORT=${a[1]}
            if [ $YES_CREDS -eq 1 ]; then
                HTTPS_PROXY_URL="${HTTPS_PROXY_PROTOCOL}://${PROXY_USERNAME}:${PROXY_PASSWORD}@${HTTPS_PROXY_SERVER_PORT}"
            else
                HTTPS_PROXY_URL="${HTTPS_PROXY_PROTOCOL}://${HTTPS_PROXY_SERVER_PORT}"
            fi
            echo "HTTPS_PROXY=\"${HTTPS_PROXY_URL}\"" >> ${PROXY_CONF}
            cat > ${DOCKER_PROXY}/https-proxy.conf << __HTTPS_DOCKER_PROXY__
[Service]
Environment="HTTPS_PROXY=${HTTPS_PROXY_URL}" "NO_PROXY=${NO_PROXY}"
__HTTPS_DOCKER_PROXY__
        else
	    echo -e "\e[91mInvalid HTTPS Proxy URL supplied" > /dev/console
        fi
    fi
fi
