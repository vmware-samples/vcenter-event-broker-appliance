#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Event Processor

set -euo pipefail

echo -e "\e[92mCreating VMware namespace ..." > /dev/console
kubectl --kubeconfig /root/.kube/config create namespace vmware

kubectl --kubeconfig /root/.kube/config -n vmware create secret generic basic-auth \
        --from-literal=basic-auth-user=admin \
        --from-literal=basic-auth-password="${ROOT_PASSWORD}"

# Setup Event Processor Configuration File
EVENT_ROUTER_CONFIG=/root/event-router-config.json

if [ "${EVENT_PROCESSOR_TYPE}" == "AWS EventBridge" ]; then
    echo -e "\e[92mSetting up AWS Event Bridge Processor ..." > /dev/console
    cat > ${EVENT_ROUTER_CONFIG} << __AWS_EVENTBRIDGE_PROCESSOR__
[{
		"type": "stream",
		"provider": "vmware_vcenter",
		"address": "https://${VCENTER_SERVER}/sdk",
		"auth": {
			"method": "user_password",
			"secret": {
				"username": "${VCENTER_USERNAME}",
				"password": "${VCENTER_PASSWORD}"
			}
		},
		"options": {
			"insecure": "${VCENTER_DISABLE_TLS}"
		}
	},
	{
		"type": "processor",
		"provider": "aws_event_bridge",
		"auth": {
			"method": "access_key",
			"secret": {
				"aws_access_key_id": "${AWS_EVENTBRIDGE_ACCESS_KEY}",
				"aws_secret_access_key": "${AWS_EVENTBRIDGE_ACCESS_SECRET}"
			}
		},
		"options": {
			"aws_region": "${AWS_EVENTBRIDGE_REGION}",
			"aws_eventbridge_event_bus": "${AWS_EVENTBRIDGE_EVENT_BUS}",
			"aws_eventbridge_rule_arn": "${AWS_EVENTBRIDGE_RULE_ARN}"
		}
	},
	{
		"type": "metrics",
		"provider": "internal",
		"address": "0.0.0.0:8080",
		"auth": {
			"method": "basic_auth",
			"secret": {
				"username": "admin",
				"password": "${ROOT_PASSWORD}"
			}
		}
	}
]
__AWS_EVENTBRIDGE_PROCESSOR__
echo "Processor: EventBridge" >> /etc/veba-release
else
    # Setup OpenFaaS
    echo -e "\e[92mSetting up OpenFaas Processor ..." > /dev/console
    kubectl --kubeconfig /root/.kube/config create -f /root/download/faas-netes/namespaces.yml

    # Setup OpenFaaS Secret
    kubectl --kubeconfig /root/.kube/config -n openfaas create secret generic basic-auth \
        --from-literal=basic-auth-user=admin \
        --from-literal=basic-auth-password="${OPENFAAS_PASSWORD}"

    kubectl --kubeconfig /root/.kube/config create -f /root/download/faas-netes/yaml

    cat > ${EVENT_ROUTER_CONFIG} << __OPENFAAS_PROCESSOR__
[{
		"type": "stream",
		"provider": "vmware_vcenter",
		"address": "https://${VCENTER_SERVER}/sdk",
		"auth": {
			"method": "user_password",
			"secret": {
				"username": "${VCENTER_USERNAME}",
				"password": "${VCENTER_PASSWORD}"
			}
		},
		"options": {
			"insecure": "${VCENTER_DISABLE_TLS}"
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
			"method": "basic_auth",
			"secret": {
				"username": "admin",
				"password": "${ROOT_PASSWORD}"
			}
		}
	}
]
__OPENFAAS_PROCESSOR__
echo "Processor: OpenFaaS" >> /etc/veba-release
fi