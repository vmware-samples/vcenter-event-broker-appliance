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

## Knative Function Troubleshooting

If a function is not behaving as expected, you can look at the logs to troubleshoot. You can either perform this operation remotely by copying the `/root/.kube/config` onto your local desktop and you can interact with VEBA using your local kubectl client or SSH to the appliance using the local kubectl client.

List out the pods.

```bash
kubectl get pods -A
```

This is an example output:

```
NAMESPACE            NAME                                                READY   STATUS      RESTARTS   AGE
contour-external     contour-5869594b-b5tqk                              1/1     Running     0          22h
contour-external     contour-5869594b-q94vd                              1/1     Running     0          22h
contour-external     contour-certgen-v1.10.0-hm65q                       0/1     Completed   0          22h
contour-external     envoy-6p2bv                                         2/2     Running     0          22h
contour-internal     contour-5d47766fd8-5skt7                            1/1     Running     0          22h
contour-internal     contour-5d47766fd8-6r5g4                            1/1     Running     0          22h
contour-internal     contour-certgen-v1.10.0-rdd6z                       0/1     Completed   0          22h
contour-internal     envoy-wwxct                                         2/2     Running     0          22h
knative-eventing     eventing-controller-658f454d9d-mnqjb                1/1     Running     0          22h
knative-eventing     eventing-webhook-69fdcdf8d4-mljtb                   1/1     Running     0          22h
knative-eventing     rabbitmq-broker-controller-88fc96b44-bbvnb          1/1     Running     0          22h
knative-serving      activator-85cd6f6f9-wl7rf                           1/1     Running     0          22h
knative-serving      autoscaler-7959969587-z9rtq                         1/1     Running     0          22h
knative-serving      contour-ingress-controller-6d5777577c-f5qzn         1/1     Running     0          22h
knative-serving      controller-577558f799-mdjpt                         1/1     Running     0          22h
knative-serving      webhook-78f446786-bn7xm                             1/1     Running     0          22h
kube-system          antrea-agent-vbr5d                                  2/2     Running     0          22h
kube-system          antrea-controller-85c944dc84-jc28b                  1/1     Running     0          22h
kube-system          coredns-74ff55c5b-rqm7c                             1/1     Running     0          22h
kube-system          coredns-74ff55c5b-vq827                             1/1     Running     0          22h
kube-system          etcd-sjc-veba-01.tshirts.inc                        1/1     Running     0          22h
kube-system          kube-apiserver-sjc-veba-01.tshirts.inc              1/1     Running     0          22h
kube-system          kube-controller-manager-sjc-veba-01.tshirts.inc     1/1     Running     0          22h
kube-system          kube-proxy-mwpxs                                    1/1     Running     0          22h
kube-system          kube-scheduler-sjc-veba-01.tshirts.inc              1/1     Running     0          22h
local-path-storage   local-path-provisioner-5696dbb894-7x626             1/1     Running     0          22h
rabbitmq-system      rabbitmq-cluster-operator-7bbbb8d559-dqd85          1/1     Running     0          22h
vmware-functions     default-broker-ingress-5c98bf68bc-2zpc6             1/1     Running     0          22h
vmware-functions     kn-pcli-tag-00001-deployment-c845447d4-lnmrq        2/2     Running     0          7h41m
vmware-functions     sockeye-65697bdfc4-cmfxc                            1/1     Running     0          22h
vmware-functions     sockeye-trigger-dispatcher-7f4dbd7f78-n589p         1/1     Running     0          22h
vmware-functions     veba-pcli-tag-trigger-dispatcher-7b477dd84d-zl2vm   1/1     Running     0          7h41m
vmware-system        cadvisor-sk4j9                                      1/1     Running     0          22h
vmware-system        tinywww-dd88dc7db-dqnnc                             1/1     Running     0          22h
vmware-system        veba-rabbit-server-0                                1/1     Running     0          22h
vmware-system        veba-ui-54967b4bf4-lpjrn                            1/1     Running     0          22h
vmware-system        vmware-event-router-vcenter-6b76959df5-6mrb4        1/1     Running     3          22h
vmware-system        vmware-event-router-webhook-6b48cc5b8c-sjzx8        1/1     Running     0          22h
```

First, we want to see if the event router is capturing events and forwarding them on to a function.

Use this command to follow the live Event Router log.

```bash
kubectl logs -n vmware-system deploy/vmware-event-router-vcenter
```

For this sample troubleshooting, we have the sample [PowerCLI Tagging function](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/examples/knative/powercli/kn-pcli-tag) running which will react to a VM powered on Event (`DrsVmPoweredOnEvent`). To see if the appliance is properly handling the event, create a test VM and power it on before proceeding with the next steps.

When we look at the log output, we see various entries regarding `DrsVmPoweredOnEvent`, ending with the following:

```
2021-09-24T13:57:05.015Z	INFO	[KNATIVE]	knative/knative.go:181	sending event	{"eventID": "c6d61d55-8100-459e-a1e7-7a936bac6e43", "subject": "DrsVmPoweredOnEvent"}
2021-09-24T13:57:05.015Z	INFO	[KNATIVE]	knative/knative.go:193	successfully sent event	{"eventID": "c6d61d55-8100-459e-a1e7-7a936bac6e43"}
2021-09-24T13:57:06.017Z	INFO	[VCENTER]	vcenter/vcenter.go:343	invoking processor	{"eventID": "87af3e86-6516-4377-9a9b-86c4c9b00b05"}
```

This lets us know that the function was invoked. If we still don't see the expected result, we need to look at the function logs.

Each Knative function will have its own pod running in the vmware-functions namespace. If you have deployed the provided tagging function example from the VEBA examples, you can examine the logs with the following command.


```bash
kubectl logs -n vmware-functions deployment/kn-pcli-tag-00001-deployment user-container
```

> **Note:** Replace the name of the deployment in the examples with the name within your environment.

We don't need the `--follow` switch because we are just trying to look at recent logs, but `--follow` would work too.

This command will show you the last 5 minutes worth of logs.

```bash
kubectl logs -n vmware-functions deployment/kn-pcli-tag-00001-deployment user-container --since=5m
```

This command will show you the last 20 lines of logs.

```bash
kubectl logs -n vmware-functions deployment/kn-pcli-tag-00001-deployment user-container --tail=20
```

Log output showing a successful function invocation:

```
09/24/2021 13:57:05 - Applying vSphere Tag "VEBA" to My-VEBA-Test-VM ...

09/24/2021 13:57:09 - vSphere Tag Operation complete ...

09/24/2021 13:57:09 - Handler Processing Completed ...
```
