#!/bin/bash

# Note: the default function namespace is openfaas-fn
# you can change this namespace to your own
kubectl create secret generic vc-credentials -n openfaas-fn \ 
    --from-literal=vc-password='<vcenter_pass>' \
    --from-literal=vc-host='<vcenter_host_ip_fqdn>' \
    --from-literal=vc-user='<vcenter_username>'