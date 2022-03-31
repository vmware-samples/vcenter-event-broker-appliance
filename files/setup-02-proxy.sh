#!/bin/bash
# Copyright 2021 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Network Proxy for both OS and Containerd

set -euo pipefail

if [ -n "${HTTP_PROXY}" ] || [ -n "${HTTPS_PROXY}" ]; then
    PROXY_CONF=/etc/sysconfig/proxy
    CONTAINERD_CONF=/usr/lib/systemd/system/containerd.service

    echo -e "\e[92mConfiguring Proxy ..." > /dev/console
    echo "PROXY_ENABLED=\"yes\"" > ${PROXY_CONF}
    YES_CREDS=0
    if [ -n "${PROXY_USERNAME}" ] && [ -n "${PROXY_PASSWORD}" ]; then
        YES_CREDS=1
    fi

    if [ ! -z "${NO_PROXY}" ]; then
        echo "NO_PROXY=\"${NO_PROXY}\"" >> ${PROXY_CONF}
        sed -i "/^\[Install\]/i Environment=NO_PROXY=${NO_PROXY}" ${CONTAINERD_CONF}
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
            echo "HTTP_PROXY='${HTTP_PROXY_URL}'" >> ${PROXY_CONF}
            sed -i "/^\[Install\]/i Environment=HTTP_PROXY=${HTTP_PROXY_URL}" ${CONTAINERD_CONF}
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
            echo "HTTPS_PROXY='${HTTPS_PROXY_URL}'" >> ${PROXY_CONF}
            sed -i "/^\[Install\]/i Environment=HTTPS_PROXY=${HTTPS_PROXY_URL}" ${CONTAINERD_CONF}
        else
	    echo -e "\e[91mInvalid HTTPS Proxy URL supplied" > /dev/console
        fi
    fi
fi
