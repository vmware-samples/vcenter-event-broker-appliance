---
layout: docs
toc_id: deploy-tanzu-sources-kind
title: Deploy Tanzu Sources with KinD
description: Deploy Tanzu Sources with KinD
permalink: /kb/deploy-tanzu-sources-kind
---

# Deploy Tanzu Sources with KinD

It is possible to mimic VEBA's base functionalities on your local computer by
using KinD. KinD, short for "Kubernetes in Docker," is an open-source project
that provides a lightweight and easy way to create Kubernetes clusters for
development and testing purposes.

<!-- omit in toc -->
## Table of Contents

- [Deploy Tanzu Sources with KinD](#deploy-tanzu-sources-with-kind)
  - [Prerequisites](#prerequisites)
  - [Install the Tanzu Sources on KinD](#install-the-tanzu-sources-on-kind)
  - [Install Event Viewer Application Sockeye](#install-event-viewer-application-sockeye)
  - [Provider Type vcsim](#provider-type-vcsim)
    - [Run the vCenter Simulator](#run-the-vcenter-simulator)
    - [Create a vcsim VSphereSource](#create-a-vcsim-vspheresource)
    - [Install the govc CLI](#install-the-govc-cli)
  - [Troubleshooting](#troubleshooting)
  - [Delete a VSphereSource](#delete-a-vspheresource)
  - [Delete a HorizonSource](#delete-a-horizonsource)
  - [Delete a KinD Cluster](#delete-a-kind-cluster)

## Prerequisites

The following prerequisites must be in place:

- [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/) installed
- Kubernetes CLI
  [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl) installed
- Knative CLI [`kn`](https://knative.dev/docs/install/quickstart-install/)
  installed
- Knative CLI plugin
  [quickstart](https://knative.dev/docs/install/quickstart-install/#install-the-knative-quickstart-plugin)
  installed

The  Knative CLI plugin `quickstart` uses `kind` to create a new Kubernetes cluster locally.
Additionally, it also installs the two Knative subprojects `Serving` and `Eventing`.

**Example:**

```console
kn quickstart kind --registry

Running Knative Quickstart using Kind
✅ Checking dependencies...
    Kind version is: 0.20.0
💽 Installing local registry...
☸ Creating Kind cluster...
Creating cluster "knative" ...
 ✓ Ensuring node image (kindest/node:v1.26.6) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
 ✓ Waiting ≤ 2m0s for control-plane = Ready ⏳
 • Ready after 15s 💚
Set kubectl context to "kind-knative"
You can now use your cluster with:

kubectl cluster-info --context kind-knative

Have a nice day! 👋

🍿 Installing Knative Serving v1.12.0 ...
    CRDs installed...
    Core installed...
    Finished installing Knative Serving
🕸️ Installing Kourier networking layer v1.12.0 ...
    Kourier installed...
    Ingress patched...
    Finished installing Kourier Networking layer
🕸 Configuring Kourier for Kind...
    Kourier service installed...
    Domain DNS set up...
    Finished configuring Kourier
🔥 Installing Knative Eventing v1.12.0 ...
    CRDs installed...
    Core installed...
    In-memory channel installed...
    Mt-channel broker installed...
    Example broker installed...
    Finished installing Knative Eventing
🚀 Knative install took: 2m23s
🎉 Now have some fun with Serverless and Event Driven Apps!
```

The above provided output outlines the successful installation of a new
Kubernetes cluster, with having Knative `Serving` and `Eventing` installed.

Validate the installation:

```console
kubectl get deploy,po -A

NAMESPACE            NAME                                     READY   UP-TO-DATE   AVAILABLE   AGE
knative-eventing     deployment.apps/eventing-controller      1/1     1            1           63s
knative-eventing     deployment.apps/eventing-webhook         1/1     1            1           63s
knative-eventing     deployment.apps/imc-controller           1/1     1            1           31s
knative-eventing     deployment.apps/imc-dispatcher           1/1     1            1           31s
knative-eventing     deployment.apps/mt-broker-controller     1/1     1            1           21s
knative-eventing     deployment.apps/mt-broker-filter         1/1     1            1           21s
knative-eventing     deployment.apps/mt-broker-ingress        1/1     1            1           21s
knative-eventing     deployment.apps/pingsource-mt-adapter    0/0     0            0           63s
knative-serving      deployment.apps/activator                1/1     1            1           108s
knative-serving      deployment.apps/autoscaler               1/1     1            1           108s
knative-serving      deployment.apps/controller               1/1     1            1           108s
knative-serving      deployment.apps/net-kourier-controller   1/1     1            1           91s
knative-serving      deployment.apps/webhook                  1/1     1            1           108s
kourier-system       deployment.apps/3scale-kourier-gateway   1/1     1            1           90s
kube-system          deployment.apps/coredns                  2/2     2            2           2m11s
local-path-storage   deployment.apps/local-path-provisioner   1/1     1            1           2m10s

NAMESPACE            NAME                                                READY   STATUS    RESTARTS   AGE
knative-eventing     pod/eventing-controller-79547fd7f4-x48pg            1/1     Running   0          63s
knative-eventing     pod/eventing-webhook-8458d5898c-sclpj               1/1     Running   0          63s
knative-eventing     pod/imc-controller-6d55d956f-dlj78                  1/1     Running   0          31s
knative-eventing     pod/imc-dispatcher-895bfd847-c8hnj                  1/1     Running   0          31s
knative-eventing     pod/mt-broker-controller-6754559b7c-29jvc           1/1     Running   0          21s
knative-eventing     pod/mt-broker-filter-7475984f8-gjm44                1/1     Running   0          21s
knative-eventing     pod/mt-broker-ingress-6786db9bfd-8j67c              1/1     Running   0          21s
knative-serving      pod/activator-8c964665f-wzw5t                       1/1     Running   0          108s
knative-serving      pod/autoscaler-5fc869cc5-x545x                      1/1     Running   0          108s
knative-serving      pod/controller-5946d56bc-shcsz                      1/1     Running   0          108s
knative-serving      pod/net-kourier-controller-d46684575-xdscv          1/1     Running   0          91s
knative-serving      pod/webhook-75d84c68b9-bfmrx                        1/1     Running   0          108s
kourier-system       pod/3scale-kourier-gateway-6f84654dc4-klbfc         1/1     Running   0          90s
kube-system          pod/coredns-787d4945fb-79smc                        1/1     Running   0          117s
kube-system          pod/coredns-787d4945fb-978th                        1/1     Running   0          117s
kube-system          pod/etcd-knative-control-plane                      1/1     Running   0          2m11s
kube-system          pod/kindnet-hnhml                                   1/1     Running   0          117s
kube-system          pod/kube-apiserver-knative-control-plane            1/1     Running   0          2m11s
kube-system          pod/kube-controller-manager-knative-control-plane   1/1     Running   0          2m12s
kube-system          pod/kube-proxy-j8qwh                                1/1     Running   0          117s
kube-system          pod/kube-scheduler-knative-control-plane            1/1     Running   0          2m11s
local-path-storage   pod/local-path-provisioner-6bd6454576-v46rk         1/1     Running   0          117s
```

## Install the Tanzu Sources on KinD

```console
kubectl apply -f https://github.com/vmware-tanzu/sources-for-knative/releases/latest/download/release.yaml
```

By default the [Channel based
Broker](https://knative.dev/docs/eventing/brokers/broker-types/channel-based-broker/)
is shipped with Knative Eventing. The installation of the KinD cluster also
includes the creation of an
[InMemoryChannel](https://github.com/knative/eventing/blob/release-1.12/config/channels/in-memory-channel/README.md)
broker for event routing.

The broker got installed into the `default` namespace:

```console
kn broker list

NAME             URL                                                                               AGE   CONDITIONS   READY   REASON
example-broker   http://broker-ingress.knative-eventing.svc.cluster.local/default/example-broker   14m   6 OK / 6     True
```

```console
kubectl get broker

NAME             URL                                                                               AGE   READY   REASON
example-broker   http://broker-ingress.knative-eventing.svc.cluster.local/default/example-broker   14m   True
```

If you would like to receive events from an existing vSphere environment, create a new `VShereSource` like described in section [Create a new VSphereSource via CLI](#create-a-new-vspheresource-via-cli) above.

## Install Event Viewer Application Sockeye

Sockeye lets you view incoming events in the browser, which can be helpful with
troubleshooting as well as when creating new functions.

Install Sockeye by simply executing `kubectl apply -f https://github.com/n3wscott/sockeye/releases/download/v0.7.0/release.yaml`
or by applying the following manifest file:

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: sockeye
  namespace: default
spec:
  template:
    spec:
      containerConcurrency: 0
      containers:
      - image: n3wscott/sockeye:v0.7.0
```

If not adjusted, the new Knative Service (`ksvc`) will be created in the `default` namespace:

```console
kn service list

NAME      URL                                         LATEST          AGE   CONDITIONS   READY   REASON
sockeye   http://sockeye.default.127.0.0.1.sslip.io   sockeye-00001   15m   3 OK / 3     True
```

Update the `ksvc` Sockeye to be not automatically scaled to 0 by Knative:

```console
kn service update --scale 1 sockeye
```

This command will set the values for `autoscaling.knative.dev/max-scale` as well
as for `.../min-scale` to `1`.

In order to ultimately receive events from a `broker`, a `trigger` for Sockeye
must be created:

```console
kn trigger create sockeye --broker example-broker --sink ksvc:sockeye
```

Validate the conditions of the trigger:

```console
kn trigger list

NAME      BROKER           SINK           AGE   CONDITIONS   READY   REASON
sockeye   example-broker   ksvc:sockeye   19m   7 OK / 7     True
```

You should see incoming events from the vCenter server now. Use the log output
for this:

```console
kubectl logs sockeye-00004-deployment-759dc8cffc-p6ttk
```

**Example:**

```json
[...]
Context Attributes,
  specversion: 1.0
  type: com.vmware.vsphere.DrsVmPoweredOnEvent.v0
  source: https://vcsa.mydomain.com/sdk
  id: 16957152
  time: 2023-11-27T10:00:08.715999Z
  datacontenttype: application/json
Extensions,
  eventclass: event
  knativearrivaltime: 2023-11-27T10:00:12.697365304Z
  vsphereapiversion: 8.0.2.0
Data,
  {
    "Key": 16957152,
    "ChainId": 16957145,
    "CreatedTime": "2023-11-27T10:00:08.715999Z",
    "UserName": "mydomain.com\\Administrator",
    "Datacenter": {
      "Name": "DC1",
      "Datacenter": {
        "Type": "Datacenter",
        "Value": "datacenter-1"
      }
    },
    "ComputeResource": {
      "Name": "Cluster1",
      "ComputeResource": {
        "Type": "ClusterComputeResource",
        "Value": "domain-c8"
      }
    },
    "Host": {
      "Name": "esx01.mydomain.com",
      "Host": {
        "Type": "HostSystem",
        "Value": "host-15"
      }
    },
    "Vm": {
      "Name": "vm1",
      "Vm": {
        "Type": "VirtualMachine",
        "Value": "vm-81"
      }
    },
    "Ds": null,
    "Net": null,
    "Dvs": null,
    "FullFormattedMessage": "DRS powered on vm1 on esx01.mydomain.com in DC1",
    "ChangeTag": "",
    "Template": false
  }
[...]
```

## Provider Type vcsim

The [vcsim](https://github.com/vmware/govmomi/tree/main/vcsim) provider is a Go
application to simulate VMware vCenter Server environments.

⚠️ This provider is for experimental usage only! It is limited in its
functionalities and not comparable with a real VMware vCenter Server
environment.

### Run the vCenter Simulator

Run the simulator e.g. as a pod on your Kubernetes cluster in a dedicated namespace
named `ns-vcsim` for example.

Create the new namespace:

```console
kubectl create ns ns-vcsim

namespace/ns-vcsim created
```

Instantiate the `vcsim` pod:

```console
kubectl -n ns-vcsim run vcsim --image=vmware/vcsim:v0.33.0 --port=8989 --image-pull-policy=Always

pod/vcsim created
```

Create a new Kubernetes `Service` to expose the pod on the cluster:

```console
kubectl -n ns-vcsim expose pod vcsim

service/vcsim exposed
```

### Create a vcsim VSphereSource

Create a new `VSphereSource` to receive events from `vcsim`. The source will be
created within the same namespace `ns-vcsim`.

Begin with the `auth` part like describe before in section [Create a new VSphereSource via CLI](#create-a-new-vspheresource-via-cli).

```console
kn vsphere auth create \
--namespace ns-vcsim \
--username user \
--password pass \
--name vcsim-creds \
--verify-url https://127.0.0.1:8989/sdk \
--verify-insecure
```

Create the new `VSphereSource`:

```console
kn vsphere source create \
--namespace ns-vcsim \
--name vcsim-source \
--vc-address https://vcsim.ns-vcsim:8989/sdk \
--skip-tls-verify \
--secret-ref vcsim-creds \
--sink-uri http://broker-ingress.knative-eventing.svc.cluster.local/default/example-broker  \
--encoding json
```

Validate the condition of the source:

```console
kn vsphere source list -n ns-vcsim

NAME           VCENTER                           INSECURE   CREDENTIALS   AGE     CONDITIONS   READY   REASON
vcsim-source   https://vcsim.ns-vcsim:8989/sdk   true       vcsim-creds   5m53s   3 OK / 3     True
```

Additionally, check the logs:

```console
kubectl -n ns-vcsim logs vcsim-source-adapter-6576ff85fb-s6bbw
{"level":"warn","ts":"2023-12-05T14:41:45.035Z","logger":"vsphere-source-adapter","caller":"v2/config.go:198","msg":"Tracing configuration is invalid, using the no-op default{error 26 0  empty json tracing config}","commit":"bd08a1c-dirty"}
{"level":"warn","ts":"2023-12-05T14:41:45.035Z","logger":"vsphere-source-adapter","caller":"v2/config.go:191","msg":"Sink timeout configuration is invalid, default to -1 (no timeout)","commit":"bd08a1c-dirty"}
{"level":"info","ts":"2023-12-05T14:41:45.051Z","logger":"vsphere-source-adapter","caller":"kvstore/kvstore_cm.go:54","msg":"Initializing configMapKVStore...","commit":"bd08a1c-dirty"}
{"level":"info","ts":"2023-12-05T14:41:45.057Z","logger":"vsphere-source-adapter","caller":"vsphere/adapter.go:92","msg":"configuring checkpointing","commit":"bd08a1c-dirty","ReplayWindow":"5m0s","Period":"10s"}
{"level":"warn","ts":"2023-12-05T14:41:45.057Z","logger":"vsphere-source-adapter","caller":"vsphere/adapter.go:131","msg":"could not retrieve checkpoint configuration","commit":"bd08a1c-dirty","error":"key checkpoint does not exist"}
{"level":"info","ts":"2023-12-05T14:41:45.057Z","logger":"vsphere-source-adapter","caller":"vsphere/adapter.go:311","msg":"no valid checkpoint found","commit":"bd08a1c-dirty"}
{"level":"info","ts":"2023-12-05T14:41:45.057Z","logger":"vsphere-source-adapter","caller":"vsphere/adapter.go:312","msg":"setting begin of event stream","commit":"bd08a1c-dirty","beginTimestamp":"2023-12-05 14:41:45.057729228 +0000 UTC"}
```

### Install the govc CLI

[`govc`](https://github.com/vmware/govmomi/tree/master/govc) is used to perform
operations against the (simulated) vCenter, e.g. powering off a virtual machine
which will trigger a corresponding event.

```console
brew install govc

govc about
govc: specify an ESX or vCenter URL
```

In a separate terminal create a Kubernetes port-forwarding so we can use `govc`
to connect to `vcsim` running inside Kubernetes:

```console
kubectl -n ns-vcsim port-forward pod/vcsim 8989:8989

Forwarding from 127.0.0.1:8989 -> 8989
Forwarding from [::1]:8989 -> 8989
```

Open the logs of the instantiated container:

```console
kubectl -n ns-vcsim logs vcsim

export GOVC_URL=https://user:pass@10.244.0.73:8989/sdk GOVC_SIM_PID=1
```

If the information above isn't displayed anymore, just replace the pod IP
address with the one you'll receive from `kubectl -n ns-vcsim get pods -o wide`.

Open another terminal to trigger an event. First, set `govc` environment
variables (connection):

```console
# ignore self-signed certificate warnings
export GOVC_INSECURE=1

# use default credentials and local port-forwarding address
export GOVC_URL=https://user:pass@127.0.0.1:8989/sdk GOVC_SIM_PID=1
```

List all available resources of the simulated vSphere environment.

```console
govc ls

/DC0/vm
/DC0/host
/DC0/datastore
/DC0/network
```

Trigger an event and observe the output in `Sockeye`:

```console
govc vm.power -off /DC0/vm/DC0_H0_VM0

Powering off VirtualMachine:vm-55... OK
```

`Sockeye` should show a `com.vmware.vsphere.VmStoppingEvent.v0` event followed by a `com.vmware.vsphere.VmPoweredOffEvent.v0` event.
.

```json
[...]

got Validation: valid
Context Attributes,
  specversion: 1.0
  type: com.vmware.vsphere.VmStoppingEvent.v0
  source: https://vcsim.ns-vcsim:8989/sdk
  id: 40
  time: 2023-12-05T13:06:12.871792908Z
  datacontenttype: application/json
Extensions,
  eventclass: event
  knativearrivaltime: 2023-12-05T13:06:13.764523682Z
  vsphereapiversion: 6.5
Data,
  {
    "Key": 40,
    "ChainId": 40,
    "CreatedTime": "2023-12-05T13:06:12.871792908Z",
    "UserName": "user",
    "Datacenter": {
      "Name": "DC0",
      "Datacenter": {
        "Type": "Datacenter",
        "Value": "datacenter-2"
      }
    },
    "ComputeResource": {
      "Name": "DC0_H0",
      "ComputeResource": {
        "Type": "ComputeResource",
        "Value": "computeresource-23"
      }
    },
    "Host": {
      "Name": "DC0_H0",
      "Host": {
        "Type": "HostSystem",
        "Value": "host-21"
      }
    },
    "Vm": {
      "Name": "DC0_H0_VM0",
      "Vm": {
        "Type": "VirtualMachine",
        "Value": "vm-55"
      }
    },
    "Ds": {
      "Name": "LocalDS_0",
      "Datastore": {
        "Type": "Datastore",
        "Value": "datastore-52"
      }
    },
    "Net": null,
    "Dvs": null,
    "FullFormattedMessage": "DC0_H0_VM0 on host DC0_H0 in DC0 is stopping",
    "ChangeTag": "",
    "Template": false
  }

[...]
```

If you don't see any output, make sure you followed all steps above, with
correct naming and that all resources (`broker`, `trigger`, `service`, `router`,
etc.) are in a `READY` state.

## Troubleshooting

All components follow the Knative logging convention.
The log level (`debug`, `info`, `error`, etc. is configurable per component,
e.g. `vsphere-source-webhook`, `VSphereSource` adapter, etc.

The default logging level is `info`.

The log level for adapters, e.g. a particular `VSphereSource` `deployment` can
be changed at runtime via the `config-logging` `ConfigMap` which is
[created](./config/config-logging.yaml) when deploying the Tanzu Sources for
Knative manifests in this repository.

⚠️ **Note:** These settings will affect **all adapter** (created sources) deployments.
Changes to a particular adapter deployment are currently not possible.

```console
kubectl -n vmware-sources edit cm config-logging
```

An interactive editor opens. Change the settings in the JSON object under the
`zap-logger-config` key. For example, to change the log level from `info` to
`debug` use this configuration in the editor:

```yaml
apiVersion: v1
data:
  # details omitted
  zap-logger-config: |
    {
      "level": "debug"
      "development": false,
      "outputPaths": ["stdout"],
      "errorOutputPaths": ["stderr"],
      "encoding": "json",
      "encoderConfig": {
        "timeKey": "ts",
        "levelKey": "level",
        "nameKey": "logger",
        "callerKey": "caller",
        "messageKey": "msg",
        "stacktraceKey": "stacktrace",
        "lineEnding": "",
        "levelEncoder": "",
        "timeEncoder": "iso8601",
        "durationEncoder": "",
        "callerEncoder": ""
      }
    }
```

Save and leave the interactive editor to apply the `ConfigMap` changes.
Kubernetes will validate and confirm the changes:

```console
configmap/config-logging edited
```

To verify that the `Source` adapter owners (e.g. `vsphere-source-webhook` for a
`VSphereSource`) have noticed the desired change, inspect the log messages of
the owner (here: `vsphere-source-webhook`) `Pod`:

```console
vsphere-source-webhook-f7d8ffbc9-4xfwl vsphere-source-webhook {"level":"info","ts":"2022-03-29T12:25:20.622Z","logger":"vsphere-source-webhook","caller":"vspheresource/vsphere.go:250","msg":"update from logging ConfigMap{snip...}
```

⚠️ **Note:** To avoid unwanted disruption during event retrieval/delivery, the
changes are **not applied** automatically to deployed adapters, i.e.
`VSphereSource` adapter, etc. The operator is in full control over the lifecycle
(downtime) of the affected `Deployment(s)`.

To make the changes take affect for existing adapter `Deployment`, an operator
needs to manually perform a rolling upgrade. The existing adapter `Pod` will be
terminated and a new instanced created with the desired log level changes.

```console
kubectl get vspheresource

NAME                SOURCE                     SINK                                                                              READY   REASON
example-vc-source   https://my-vc.corp.local   http://broker-ingress.knative-eventing.svc.cluster.local/default/example-broker   True
```

```console
kubectl rollout restart deployment/example-vc-source-adapter

deployment.apps/example-vc-source-adapter restarted
```

⚠️ **Note:** To avoid losing events due to this (brief) downtime, consider
enabling the [Checkpointing](#configuring-checkpoint-and-event-replay)
capability.


More details can be found at the Tanzu Sources for Knative repository on Github - [Changing Log Levels](https://github.com/vmware-tanzu/sources-for-knative/tree/main#changing-log-levels).

## Delete a VSphereSource

```console
kn vsphere source delete --name vcsim-source --namespace ns-vcsim
```

## Delete a HorizonSource

```console
kn source delete --name horizon-example --namespace vmware-system
```

## Delete a KinD Cluster

```code
kind delete cluster --name knative

Deleting cluster "knative" ...
Deleted nodes: ["knative-control-plane"]
```
