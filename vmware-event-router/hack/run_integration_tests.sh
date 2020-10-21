#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

set -euo pipefail

#################################
### variables
TIMEOUT_CMD="timeout"
BOM_FILE="../veba-bom.json"
JQBIN="jq -r"

# colors
RED='\033[31m'
GREEN='\033[32m'
RESET='\033[0m'

# OpenFaaS
OF_VERSION=$(${JQBIN} '.openfaas.gitRepoTag' ${BOM_FILE})                # faas-netes version to use
OF_TIMEOUT=3m                                                         # exit if OpenFaaS is not ready within this time
CLI_VERSION=$(${JQBIN} '.openfaas."faas-cli"["version"]' ${BOM_FILE}) # faas-cli version
PORT_FWD="deploy/gateway 8080:8080"                                   # local/remote ports to use for port-forwarding to OpenFaaS gateway
OKFN=of-echo                                                          # OpenFaaS function
FAILFN=of-fail                                                        # OpenFaaS function
FN_TIMEOUT=1m                                                         # exit if OpenFaaS functions are not ready within this time

# AWS EventBridge
AWS_SECRET=secret_aws.json # when running AWS integration tests, secret file holding AWS config at index [1]

# Kubernetes (kind)
K8S_VERSION=$(${JQBIN} '.kubernetes.gitRepoTag' ${BOM_FILE})
KINDBIN="kind"
KIND_TIMEOUT=5m # exit if kind cluster creation does not complete within this time
KIND_WAIT=2m    # wait for control plane to show ready
KIND_CLUSTER="veba-integration"
# using custom image built with kind since not all versions are pushed to kindtest
KIND_IMAGE="embano1/node:${K8S_VERSION}"
#################################

tmp_dir=$(mktemp -d -t ci-XXXXXXXXXX)

cecho() {
    local default_msg="No message passed."
    message=${1:-$default_msg}
    color=${2:-$RESET}

    counter=$((${counter:-0} + 1))
    echo -e ${color}"\n"\[ ${counter} \] ${message} ${RESET}
    return
}

function cleanup() {
    # capture previous return val, e.g. error
    rv=$?
    # don't run cleanup in CI env, e.g. Github Action sets CI=true
    if [ "${CI:-no}" != true ]; then
        cecho "Running cleanup and then exiting with code ${rv}" ${GREEN}
        rm -rf ${tmp_dir} && rm -rf faas-cli-* && rm -rf template && ${KINDBIN} delete cluster --name ${KIND_CLUSTER}
    fi
    exit $rv
}
trap cleanup INT TERM EXIT

# create kind cluster
cecho "---> Creating Kubernetes cluster (${K8S_VERSION}) with max wait time: ${KIND_TIMEOUT}" ${GREEN}
${TIMEOUT_CMD} ${KIND_TIMEOUT} ${KINDBIN} create cluster --name ${KIND_CLUSTER} --wait ${KIND_WAIT} --image ${KIND_IMAGE}

# generate password used by OpenFaaS for basic_auth
export OF_PASSWORD=$(head -c 12 /dev/urandom | shasum | cut -d' ' -f1)

# if running as Github Actions make it available to other steps
if [ "${CI:-no}" = true ]; then
    echo "OF_PASSWORD=$OF_PASSWORD" >> $GITHUB_ENV
fi

# deploy faas-netes
cecho "---> Deploying OpenFaaS (${OF_VERSION})" ${GREEN}
git clone https://github.com/openfaas/faas-netes ${tmp_dir} && git -C ${tmp_dir} checkout ${OF_VERSION}
kubectl create -f ${tmp_dir}/namespaces.yml
kubectl -n openfaas create secret generic basic-auth --from-literal=basic-auth-user=admin --from-literal=basic-auth-password="$OF_PASSWORD"
kubectl create -f ${tmp_dir}/yaml

cecho "---> Waiting up to ${OF_TIMEOUT} for OpenFaaS gateway to become ready" ${GREEN}
${TIMEOUT_CMD} ${OF_TIMEOUT} /bin/bash -c 'while [[ $READY -lt 1 ]]; do READY=$(kubectl -n openfaas get deploy gateway -o json | '"${JQBIN}"' .status.readyReplicas); sleep 1; done'

# start kubectl port-forwarding
cecho "---> Starting port-forwarding to OpenFaaS gateway" ${GREEN}
kubectl -n openfaas port-forward ${PORT_FWD} >/dev/null &

# get faas-cli
cecho "---> Downloading faas-cli (${CLI_VERSION})" ${GREEN}

case "${OSTYPE}" in
darwin*) FAAS_CLI="faas-cli-darwin" ;;
msys*) FAAS_CLI="faas-cli.exe" ;;
*) FAAS_CLI="faas-cli" ;;
esac

curl --silent --show-error --fail -LO https://github.com/openfaas/faas-cli/releases/download/${CLI_VERSION}/${FAAS_CLI}
chmod +x ./${FAAS_CLI}
echo ${OF_PASSWORD} | ./${FAAS_CLI} login --password-stdin

# deploy functions used by integration tests
cecho "---> Downloading templates and installing functions" ${GREEN}
./${FAAS_CLI} template pull
./${FAAS_CLI} template pull --overwrite https://github.com/openfaas-incubator/python-flask-template
./${FAAS_CLI} deploy -f https://raw.githubusercontent.com/embano1/${FAILFN}/master/stack.yml
./${FAAS_CLI} deploy -f https://raw.githubusercontent.com/embano1/${OKFN}/master/stack.yml

cecho "---> Waiting up to ${FN_TIMEOUT} for ${FAILFN} and ${OKFN} to become ready" ${GREEN}
${TIMEOUT_CMD} ${FN_TIMEOUT} /bin/bash -c 'while [[ $READY -lt 1 ]]; do READY=$(kubectl -n openfaas-fn get deploy '"${FAILFN}"' -o json | '"${JQBIN}"' .status.readyReplicas); sleep 1; done'
${TIMEOUT_CMD} ${FN_TIMEOUT} /bin/bash -c 'while [[ $READY -lt 1 ]]; do READY=$(kubectl -n openfaas-fn get deploy '"${OKFN}"' -o json | '"${JQBIN}"' .status.readyReplicas); sleep 1; done'

# run OpenFaaS integration tests
cecho "---> Running OpenFaaS integration tests" ${GREEN}
make integration-test

# optionally run AWS EventBridge integration tests
# note: won't run as Github Actions (${CI} = true)
if [ "${CI:-false}" = false ]; then
    if [ "${TEST_AWS:-false}" = true ]; then
        cecho "---> Running AWS integration tests" ${GREEN}
        AWS_ACCESS_KEY=$(${JQBIN} '.aws_access_key_id' ${AWS_SECRET}) \
        AWS_SECRET_KEY=$(${JQBIN} '.aws_secret_access_key' ${AWS_SECRET}) \
        AWS_REGION=$(${JQBIN} '.aws_region' ${AWS_SECRET}) \
        AWS_EVENT_BUS=$(${JQBIN} '.aws_eventbridge_event_bus' ${AWS_SECRET}) \
        AWS_RULE_ARN=$(${JQBIN} '.aws_eventbridge_rule_arn' ${AWS_SECRET}) \
            go test ./internal/integration/... -count 1 -race --tags=integration,aws -v -ginkgo.v
    fi
fi
