---
layout: docs
toc_id: troubleshoot-functions
title: VMware Event Broker Function Troubleshooting
description: Troubleshooting guide for general function issues
permalink: /kb/troubleshoot-functions
cta:
 title: Still having trouble? 
 description: Please submit bug reports and feature requests by using our GitHub [Issues](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues){:target="_blank"} page or Join us on slack [#vcenter-event-broker-appliance](https://vmwarecode.slack.com/archives/CQLT9B5AA){:target="_blank"} on vmwarecode.slack.com.
---

## OpenFaaS Function Troubleshooting

If a function is not behaving as expected, you can look at the logs to troubleshoot. First, SSH or console to the appliance as shown in the Requirements section.

List out the pods.

```bash
kubectl get pods -A
```

This is the function output:

```
NAMESPACE        NAME                                   READY   STATUS      RESTARTS   AGE
kube-system      coredns-584795fc57-4bp2s               1/1     Running     1          6d4h
kube-system      coredns-584795fc57-76pwr               1/1     Running     1          6d4h
kube-system      etcd-veba01                            1/1     Running     2          6d4h
kube-system      kube-apiserver-veba01                  1/1     Running     2          6d4h
kube-system      kube-controller-manager-veba01         1/1     Running     3          6d4h
kube-system      kube-proxy-fvf2n                       1/1     Running     2          6d4h
kube-system      kube-scheduler-veba01                  1/1     Running     2          6d4h
kube-system      weave-net-v9jss                        2/2     Running     6          6d4h
openfaas-fn      powercli-entermaint-d84fd8d85-sjdgl    1/1     Running     1          6d4h
openfaas         alertmanager-58f8d787d9-nqwm8          1/1     Running     1          6d4h
openfaas         basic-auth-plugin-dd49cd66b-rv6n7      1/1     Running     1          6d4h
openfaas         faas-idler-59ff9778fd-84szz            1/1     Running     4          6d4h
openfaas         gateway-74f6f9489b-btgz8               2/2     Running     5          6d4h
openfaas         nats-6dfbf45d77-9swph                  1/1     Running     1          6d4h
openfaas         prometheus-5f5494b54f-srs2d            1/1     Running     1          6d4h
openfaas         queue-worker-59b67bf4-wqhm5            1/1     Running     4          6d4h
projectcontour   contour-5cddfc8f6-hpzn8                1/1     Running     1          6d4h
projectcontour   contour-5cddfc8f6-tdv2r                1/1     Running     2          6d4h
projectcontour   contour-certgen-wrgnb                  0/1     Completed   0          6d4h
projectcontour   envoy-8mdhb                            1/1     Running     2          6d4h
vmware           tinywww-7fcfc6fb94-v7ncj               1/1     Running     1          6d4h
vmware           vmware-event-router-5dd9c8f858-9g44h   1/1     Running     4          6d4h
```

First, we want to see if the event router is capturing events and forwarding them on to a function. 

Use this command to follow the live Event Router log.

```bash
kubectl logs -n vmware vmware-event-router-5dd9c8f858-9g44h --follow
```

For this sample troubleshooting, we have the sample hostmaintenance alarms function running. To see if the appliance is properly handling the event, we put a host into maintenance mode. 

When we look at the log output, we see various entries regarding EnteredMaintenanceModeEvent, ending with the following:

```
[OpenFaaS] 2020/03/11 22:15:09 invoking function(s) on topic: EnteredMaintenanceModeEvent
[OpenFaaS] 2020/03/11 22:15:09 successfully invoked function powercli-entermaint for topic EnteredMaintenanceModeEvent
```

This lets us know that the function was invoked. If we still don't see the expected result, we need to look at the function logs.

Each OpenFaaS function will have its own pod running in the openfaas-fn namespace. We can examine the logs with the following command.

```bash
kubectl logs -n openfaas-fn powercli-entermaint-d84fd8d85-sjdgl
```

We don't need the --follow switch because we are just trying to look at recent logs, but --follow would work too.  
Some other useful switches are `--since` and `--tail`. 

This command will show you the last 5 minutes worth of logs.

```bash
kubectl logs -n openfaas-fn powercli-entermaint-d84fd8d85-sjdgl --since=5m
```

This command will show you the last 20 lines of logs.

```bash
kubectl logs -n openfaas-fn powercli-entermaint-d84fd8d85-sjdgl --tail=20
```

Log output showing a succesful function invocation: 

```
Connecting to vCenter Server ...

Disabling alarm actions on host: esx01.labad.int
Disconnecting from vCenter Server ...

2020/03/11 22:15:15 Duration: 6.085448 seconds
```

An alternative way to troubleshoot OpenFaaS logs is to use `faas-cli`.

This faas-cli command will show all available functions in the appliance. ```--tls-no-verify``` bypasses SSL certificate validation

```bash
faas-cli list --tls-no-verify
```

The command output is:

```
Function                        Invocations     Replicas
powercli-entermaint             3               1
```

We can  look at the logs with this command.

```bash
faas-cli logs powercli-entermaint --tls-no-verify
```

The logs are the same: 

```
2020-03-11T22:15:15Z Connecting to vCenter Server ...
2020-03-11T22:15:15Z
2020-03-11T22:15:15Z Disabling alarm actions on host: esx01.labad.int
2020-03-11T22:15:15Z Disconnecting from vCenter Server ...
2020-03-11T22:15:15Z
2020-03-11T22:15:15Z 2020/03/11 22:15:15 Duration: 6.085448 seconds
```

All of the same switches shown in the kubectl commands such as `--tail` and `--since` work with `faas-cli`.

> **Note:** `faas-cli` will stop tailing the log after a fixed period.

## OpenFaaS Gateway not available

### Self-signed certificate errors

The most common issue with OpenFaaS is related to certificates. The appliance certificate is self-signed. Attempting to connect to it without ignoring certificate errors results in an error, shown below

```bash
faas-cli secret list
```

The command output is:
```bash
Cannot connect to OpenFaaS on URL: https://veba02.lab.int
```

Adding the switch `--tls-no-verify` allows you to bypass SSL errors
```bash
faas-cli secret list --tls-no-verify
```

The command output is:
```bash
NAME
vc-hostmaint-config
```
Alternatively, you can replace the self-signed certificate with a signed certificate

### Requirements
* Root access to the appliance
* A public/private key pair copied to a folder on the appliance filesystem

Run the following commands

```bash
cd /path/to/your/cert/files
CERT_NAME=eventrouter-tls 
KEY_FILE=yourkeyfile.pem
CERT_FILE=yourcertfile.cer

#recreate the tls secret
kubectl --kubeconfig /root/.kube/config -n vmware delete secret ${CERT_NAME}
kubectl --kubeconfig /root/.kube/config -n vmware create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}

#reapply the config to take the new certificate
kubectl --kubeconfig /root/.kube/config apply -f /root/ingressroute-gateway.yaml
```
