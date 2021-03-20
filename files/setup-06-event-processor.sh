#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Event Processor

set -euo pipefail

kubectl -n vmware-system create secret generic basic-auth \
        --from-literal=basic-auth-user=admin \
        --from-literal=basic-auth-password="${ROOT_PASSWORD}"

# Setup Event Processor Configuration File
EVENT_ROUTER_CONFIG=/root/config/event-router-config.yaml

# Slicing of escaped variables needed to properly handle the double quotation issue with constructing vCenter Server URL
ESCAPED_VCENTER_SERVER=$(echo -n ${VCENTER_SERVER} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')
ESCAPED_VCENTER_USERNAME=$(echo -n ${VCENTER_USERNAME} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)')
ESCAPED_VCENTER_PASSWORD=$(echo -n ${VCENTER_PASSWORD} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)')
ESCAPED_ROOT_PASSWORD=$(echo -n ${ROOT_PASSWORD} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)')

if [ "${EVENT_PROCESSOR_TYPE}" == "Knative" ]; then
    echo -e "\e[92mSetting up Knative Processor ..." > /dev/console

    # External Knative Broker
    if [ "${KNATIVE_DEPLOYMENT_TYPE}" == "external" ]; then
      cat > ${EVENT_ROUTER_CONFIG} << __KNATIVE_PROCESSOR__
apiVersion: event-router.vmware.com/v1alpha1
kind: RouterConfig
metadata:
  name: router-config-knative
eventProcessor:
  name: veba-knative
  type: knative
  knative:
    insecureSSL: ${KNATIVE_DISABLE_TLS}
    encoding: binary
    destination:
      uri:
        host: ${KNATIVE_HOST}
        scheme: ${KNATIVE_SCHEME}
        path: ${KNATIVE_PATH}
eventProvider:
  name: veba-vc-01
  type: vcenter
  vcenter:
    address: https://${ESCAPED_VCENTER_SERVER}/sdk
    auth:
      basicAuth:
        password: ${ESCAPED_VCENTER_PASSWORD}
        username: ${ESCAPED_VCENTER_USERNAME}
      type: basic_auth
    insecureSSL: ${VCENTER_DISABLE_TLS}
    checkpoint: false
metricsProvider:
  default:
    bindAddress: 0.0.0.0:8082
  name: veba-metrics
  type: default
__KNATIVE_PROCESSOR__
    else
      # Embedded Knative Broker
      cat > ${EVENT_ROUTER_CONFIG} << __KNATIVE_PROCESSOR__
apiVersion: event-router.vmware.com/v1alpha1
kind: RouterConfig
metadata:
  name: router-config-knative
eventProcessor:
  name: veba-knative
  type: knative
  knative:
    insecureSSL: false
    encoding: binary
    destination:
      ref:
        apiVersion: eventing.knative.dev/v1
        kind: Broker
        name: rabbit
        namespace: default
eventProvider:
  name: veba-vc-01
  type: vcenter
  vcenter:
    address: https://${ESCAPED_VCENTER_SERVER}/sdk
    auth:
      basicAuth:
        password: ${ESCAPED_VCENTER_PASSWORD}
        username: ${ESCAPED_VCENTER_USERNAME}
      type: basic_auth
    insecureSSL: ${VCENTER_DISABLE_TLS}
    checkpoint: false
metricsProvider:
  default:
    bindAddress: 0.0.0.0:8082
  name: veba-metrics
  type: default
__KNATIVE_PROCESSOR__
    fi
echo "Processor: Knative" >> /etc/veba-release
elif [ "${EVENT_PROCESSOR_TYPE}" == "AWS EventBridge" ]; then
    echo -e "\e[92mSetting up AWS Event Bridge Processor ..." > /dev/console

	ESCAPED_AWS_EVENTBRIDGE_ACCESS_KEY=$(echo -n ${AWS_EVENTBRIDGE_ACCESS_KEY} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)')
	ESCAPED_AWS_EVENTBRIDGE_ACCESS_SECRET=$(echo -n ${AWS_EVENTBRIDGE_ACCESS_SECRET} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)')
	ESCAPED_AWS_EVENTBRIDGE_EVENT_BUS=$(echo -n ${AWS_EVENTBRIDGE_EVENT_BUS} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)')
	ESCAPED_AWS_EVENTBRIDGE_RULE_ARN=$(echo -n ${AWS_EVENTBRIDGE_RULE_ARN} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)')

    cat > ${EVENT_ROUTER_CONFIG} << __AWS_EVENTBRIDGE_PROCESSOR__
apiVersion: event-router.vmware.com/v1alpha1
eventProcessor:
  awsEventBridge:
    auth:
      awsAccessKeyAuth:
        accessKey: ${ESCAPED_AWS_EVENTBRIDGE_ACCESS_KEY}
        secretKey: ${ESCAPED_AWS_EVENTBRIDGE_ACCESS_SECRET}
      type: aws_access_key
    eventBus: ${ESCAPED_AWS_EVENTBRIDGE_EVENT_BUS}
    region: ${AWS_EVENTBRIDGE_REGION}
    ruleARN: ${ESCAPED_AWS_EVENTBRIDGE_RULE_ARN}
  name: veba-aws
  type: awsEventBridge
eventProvider:
  name: veba-vc-01
  type: vcenter
  vcenter:
    address: https://${ESCAPED_VCENTER_SERVER}/sdk
    auth:
      basicAuth:
        password: ${ESCAPED_VCENTER_PASSWORD}
        username: ${ESCAPED_VCENTER_USERNAME}
      type: basic_auth
    insecureSSL: ${VCENTER_DISABLE_TLS}
    checkpoint: false
kind: RouterConfig
metadata:
  labels:
    key: value
  name: router-config-aws
metricsProvider:
  default:
    bindAddress: 0.0.0.0:8082
  name: veba-metrics
  type: default
__AWS_EVENTBRIDGE_PROCESSOR__
echo "Processor: EventBridge" >> /etc/veba-release
else
    # Setup OpenFaaS
    echo -e "\e[92mSetting up OpenFaas Processor ..." > /dev/console
    kubectl create -f /root/download/faas-netes/namespaces.yml

    # Setup OpenFaaS Secret
    kubectl -n openfaas create secret generic basic-auth \
        --from-literal=basic-auth-user=admin \
        --from-literal=basic-auth-password="${OPENFAAS_PASSWORD}"

    kubectl apply -f /root/download/faas-netes/yaml

	ESCAPED_OPENFAAS_PASSWORD=$(echo -n ${OPENFAAS_PASSWORD} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)')

    cat > ${EVENT_ROUTER_CONFIG} << __OPENFAAS_PROCESSOR__
apiVersion: event-router.vmware.com/v1alpha1
eventProcessor:
  name: veba-openfaas
  openfaas:
    address: http://gateway.openfaas:8080
    async: false
    auth:
      basicAuth:
        password: ${ESCAPED_OPENFAAS_PASSWORD}
        username: admin
      type: basic_auth
  type: openfaas
eventProvider:
  name: veba-vc-01
  type: vcenter
  vcenter:
    address: https://${ESCAPED_VCENTER_SERVER}/sdk
    auth:
      basicAuth:
        password: ${ESCAPED_VCENTER_PASSWORD}
        username: ${ESCAPED_VCENTER_USERNAME}
      type: basic_auth
    insecureSSL: ${VCENTER_DISABLE_TLS}
    checkpoint: false
kind: RouterConfig
metadata:
  labels:
    key: value
  name: router-config-openfaas
metricsProvider:
  default:
    bindAddress: 0.0.0.0:8082
  name: veba-metrics
  type: default
__OPENFAAS_PROCESSOR__
echo "Processor: OpenFaaS" >> /etc/veba-release
fi