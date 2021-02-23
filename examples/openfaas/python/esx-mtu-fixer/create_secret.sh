#!/bin/bash

set -exo pipefail

# Note: the default function namespace is openfaas-fn
# you can change this namespace to your own
kubectl -n openfaas-fn create secret generic vc-credentials --from-literal=vc-password='<vcenter_pass>' \
    --from-literal=vc-host='<vcenter_host_ip_fqdn>' \
    --from-literal=vc-user='<vcenter_username>'