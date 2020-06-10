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
EVENT_ROUTER_CONFIG=/root/config/event-router-config.json

# Slicing of escaped variables needed to properly handle the double quotation issue with constructing vCenter Server URL
ESCAPED_VCENTER_SERVER=$(echo -n ${VCENTER_SERVER} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')
ESCAPED_VCENTER_USERNAME=$(echo -n ${VCENTER_USERNAME} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')
ESCAPED_VCENTER_PASSWORD=$(echo -n ${VCENTER_PASSWORD} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')
ESCAPED_ROOT_PASSWORD=$(echo -n ${ROOT_PASSWORD} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')

if [ "${EVENT_PROCESSOR_TYPE}" == "AWS EventBridge" ]; then
    echo -e "\e[92mSetting up AWS Event Bridge Processor ..." > /dev/console

	ESCAPED_AWS_EVENTBRIDGE_ACCESS_KEY=$(echo -n ${AWS_EVENTBRIDGE_ACCESS_KEY} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')
	ESCAPED_AWS_EVENTBRIDGE_ACCESS_SECRET=$(echo -n ${AWS_EVENTBRIDGE_ACCESS_SECRET} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')
	ESCAPED_AWS_EVENTBRIDGE_EVENT_BUS=$(echo -n ${AWS_EVENTBRIDGE_EVENT_BUS} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')
	ESCAPED_AWS_EVENTBRIDGE_RULE_ARN=$(echo -n ${AWS_EVENTBRIDGE_RULE_ARN} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')

    cat > ${EVENT_ROUTER_CONFIG} << __AWS_EVENTBRIDGE_PROCESSOR__
[{
		"type": "stream",
		"provider": "vmware_vcenter",
		"address": "https://${ESCAPED_VCENTER_SERVER}/sdk",
		"auth": {
			"method": "user_password",
			"secret": {
				"username": "${ESCAPED_VCENTER_USERNAME}",
				"password": "${ESCAPED_VCENTER_PASSWORD}"
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
				"aws_access_key_id": "${ESCAPED_AWS_EVENTBRIDGE_ACCESS_KEY}",
				"aws_secret_access_key": "${ESCAPED_AWS_EVENTBRIDGE_ACCESS_SECRET}"
			}
		},
		"options": {
			"aws_region": "${AWS_EVENTBRIDGE_REGION}",
			"aws_eventbridge_event_bus": "${ESCAPED_AWS_EVENTBRIDGE_EVENT_BUS}",
			"aws_eventbridge_rule_arn": "${ESCAPED_AWS_EVENTBRIDGE_RULE_ARN}"
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
				"password": "${ESCAPED_ROOT_PASSWORD}"
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

	ESCAPED_OPENFAAS_PASSWORD=$(echo -n ${OPENFAAS_PASSWORD} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')

    cat > ${EVENT_ROUTER_CONFIG} << __OPENFAAS_PROCESSOR__
[{
		"type": "stream",
		"provider": "vmware_vcenter",
		"address": "https://${ESCAPED_VCENTER_SERVER}/sdk",
		"auth": {
			"method": "user_password",
			"secret": {
				"username": "${ESCAPED_VCENTER_USERNAME}",
				"password": "${ESCAPED_VCENTER_PASSWORD}"
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
				"password": "${ESCAPED_OPENFAAS_PASSWORD}"
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
				"password": "${ESCAPED_ROOT_PASSWORD}"
			}
		}
	}
]
__OPENFAAS_PROCESSOR__
echo "Processor: OpenFaaS" >> /etc/veba-release
fi