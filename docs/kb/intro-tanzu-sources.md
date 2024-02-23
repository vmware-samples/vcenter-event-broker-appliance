---
layout: docs
toc_id: tanzu-sources
title: VMware Tanzu Sources for Knative - Introduction
description: VMware Tanzu Sources for Knative Introduction
permalink: /kb/tanzu-sources
cta:
 title: Get Started
 description: Explore the capabilities that the VMware Tanzu Sources for Knative enables
 actions:
    - text: Install the [Appliance](install-veba) to extend your SDDC with our [community-sourced functions](/examples)
    - text: Learn more about the [Events in vCenter](vcenter-events) and how to find the right event for your use case
    - text: Learn more about Functions in this overview [here](functions).
---

# Introduction to VMware Tanzu Sources for Knative

VMware [Tanzu Sources for
Knative](https://github.com/vmware-tanzu/sources-for-knative) are designed to
facilitate event-driven architectures and streamline the development of event
sources for Knative.
At its core, Tanzu Sources for Knative aims to simplify the connection between
event producers and the Knative Eventing ecosystem.

The Tanzu Sources are the successor of the [VMware Event Router](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/development/vmware-event-router)
which originally got used by the VMware Event Broker Appliance project
as the core logic to forward [CloudEvents](https://cloudevents.io/),
e.g. from vSphere, to configurable event processors.

The VMware Tanzu Sources are also using the CloudEvents
standard to normalize events from the supported event `providers`.

This open-source project is officially maintained by VMware and used by the
VMware Event Broker Appliance as the core logic to source non-CloudEvent
conformant event-payloads from e.g. vSphere (`VSphereSource`) or
Horizon (`HorizonSource`) and to ultimately forward these
CloudEvent conformant to a configurable Knative
[Sink](https://knative.dev/docs/eventing/sinks/) (addressable destination).

## Technical Components

- **Source CRDs** (Custom Resource Definitions)**:** Tanzu Sources for Knative
  leverages Kubernetes CRDs to define custom `sources`, making them a native part
  of the Kubernetes API.
- **Controllers:** For each `source` type, there is a corresponding controller
  responsible for managing the lifecycle of that `source` and ensuring the
  delivery of events.
- **Event Delivery:** It uses Knative Eventing's robust delivery mechanisms,
  such as brokers, channels and subscriptions, to deliver events to the
  appropriate destinations.

<!-- omit in toc -->
## Table of Contents

- [Introduction to VMware Tanzu Sources for Knative](#introduction-to-vmware-tanzu-sources-for-knative)
  - [Technical Components](#technical-components)
  - [Available `Sources`](#available-sources)
  - [Event Delivery](#event-delivery)
    - [Event `Provider` Delivery Guarantees](#event-provider-delivery-guarantees)
      - [At-least-once Event Delivery](#at-least-once-event-delivery)
      - [At-most-once Event Delivery](#at-most-once-event-delivery)
    - [Current Limitations](#current-limitations)
  - [Installation Tanzu Sources for Knative](#installation-tanzu-sources-for-knative)
    - [Prerequisites](#prerequisites)
  - [Configuration Source Type VSphereSource](#configuration-source-type-vspheresource)
    - [Event Provider Type vCenter Server](#event-provider-type-vcenter-server)
    - [Create a new VSphereSource via CLI](#create-a-new-vspheresource-via-cli)
    - [Create a new VSphereSource via Manifest File](#create-a-new-vspheresource-via-manifest-file)
  - [Configuration Source Type HorizonSource](#configuration-source-type-horizonsource)
    - [Event Provider Type Horizon](#event-provider-type-horizon)
  - [Event Viewer Application Sockeye](#event-viewer-application-sockeye)
  - [Troubleshooting](#troubleshooting)
  - [Delete a VSphereSource](#delete-a-vspheresource)
  - [Delete a HorizonSource](#delete-a-horizonsource)

## Available `Sources`

- [VMware vCenter Server](https://www.vmware.com/products/vcenter-server.html)
  - VMware vCenter Server is an advanced server management software that provides a
centralized platform for controlling your VMware vSphere environments.
- [VMware Horizon](https://www.vmware.com/products/horizon.html)
  - VMware Horizon is a platform for delivering virtual desktops and apps
efficiently and securely across hybrid clouds.

## Event Delivery

Knative abstracts event receivers ([Sinks](https://knative.dev/docs/eventing/sinks/)) in a very flexible way via
*addressable* `Destinations`. The VMware Tanzu Sources for Knative technically
supports all Knative `Destinations`. However, Knative `Broker` or Knative
`Services` are likely the ones you will be using.

> **Note:** We have not done any testing for Knative `Channels`, since request
> response style interactions are out of scope for the Tanzu Sources.

### Event `Provider` Delivery Guarantees

#### At-least-once Event Delivery

- with the [vCenter event provider](#provider-type-vcenter) option `checkpoint:
     true`
- with the [Horizon event provider](#provider-type-horizon) option `checkpoint:
     true`

The `Source` controller will periodically checkpoint its progress in the vCenter
event stream ("history") by using a Kubernetes `ConfigMap` as storage backend.
The name of the `ConfigMap` is `<name_of_source>-configmap` (see example below).
By default, checkpoints will be created every `10 seconds`. The minimum
checkpoint frequency is `1s` but be aware of potential load on the Kubernetes
API this might cause.

Checkpointing is useful to guarantee **at-least-once** event delivery semantics,
e.g. to guard against lost events due to controller downtime (maintenance,
crash, etc.).

To influence the checkpointing logic, read up on the corresponding section on
Github - [Configuring Checkpoint and Event
Replay](https://github.com/vmware-tanzu/sources-for-knative#configuring-checkpoint-and-event-replay).

#### At-most-once Event Delivery

Checkpointing itself cannot be disabled and there will be exactly zero or one
checkpoint per controller. If **at-most-once** event delivery is desired, i.e.
no event replay upon controller start, simply set `maxAgeSeconds: 0`.

- with the [vCenter event provider](#provider-type-vcenter) option `maxAgeSeconds: 0`
- with the [Horizon event provider](#provider-type-horizon) option `maxAgeSeconds: 0`

> **Note:** Knative's built-in retry mechanisms is used so your function might
still be involved multiple times depending on its response code. However, if an
event `provider` crashes before sending an event to Knative or when the Knative
returns an error, the event is not retried and discarded.

### Current Limitations

- During the deployment of the VMware Event Broker Appliance (VM) to vSphere,
  only one vSphere source and one Horizon source can be configured at a time
  (see note below)
- At-least-once event delivery semantics are not guaranteed if a source crashes
  **within seconds** right after startup and having received *n* events but
  before creating the first valid checkpoint (current checkpoint interval is
  10s)
- If an event cannot be successfully delivered (retried) by Knative it is logged
  and discarded, i.e. there is currently no support for [dead letter
  queues](https://en.wikipedia.org/wiki/Dead_letter_queue) (see note below)

> **Note:** It is possible though to run **multiple instances** of a source
> (e.g. `VSphereSource`) with different configurations to address multi-vCenter
> scenarios. This decision was made for scalability and resource/tenancy
> isolation purposes.

> **Note:** Knative supports Dead Letter Queues when using `Broker` mode.

For more detailed information check the official [VMware Tanzu Sources for
Knative](https://github.com/vmware-tanzu/sources-for-knative) repository on
Github.

## Installation Tanzu Sources for Knative

Installing Tanzu Sources for Knative is a straightforward process that enables
you to integrate VMware event sources into your Knative environment.

### Prerequisites

The following prerequisites must be in place:
- A running Kubernetes cluster (> v1.26.0)
- Access to a Kubernetes cluster and the ability to execute `kubectl` commands
- A running [Knative installation](https://knative.dev/docs/install/) (Serving and Eventing)
- Knative CLI installed - [Installing the Knative
  CLI](https://knative.dev/docs/client/install-kn/)
- Knative CLI Plugin `knative-vsphere` installed
  - Download the binary for your OS from the [Tanzu Sources page](https://github.com/vmware-tanzu/sources-for-knative/releases/) on Github
  - `mv` the `kn-vsphere` binary into e.g. `/usr/local/bin/`

Install Tanzu Sources for Knative by applying the YAML manifest provided by
VMware. This manifest defines the necessary Kubernetes resources for Knative
event sources.

Use the `kubectl apply` command:

```console
kubectl apply -f https://github.com/vmware-tanzu/sources-for-knative/releases/latest/download/release.yaml
```

```console
namespace/vmware-sources created
serviceaccount/horizon-source-controller created
serviceaccount/horizon-source-webhook created
clusterrole.rbac.authorization.k8s.io/vsphere-receive-adapter-cm created
clusterrole.rbac.authorization.k8s.io/vmware-sources-admin created
clusterrole.rbac.authorization.k8s.io/vmware-sources-core created
clusterrole.rbac.authorization.k8s.io/podspecable-binding configured
clusterrole.rbac.authorization.k8s.io/builtin-podspecable-binding configured
serviceaccount/vsphere-controller created
clusterrole.rbac.authorization.k8s.io/horizon-source-controller created
clusterrole.rbac.authorization.k8s.io/horizon-source-observer created
clusterrolebinding.rbac.authorization.k8s.io/vmware-sources-controller-admin created
clusterrolebinding.rbac.authorization.k8s.io/vmware-sources-webhook-podspecable-binding created
clusterrolebinding.rbac.authorization.k8s.io/vmware-sources-webhook-addressable-resolver-binding created
clusterrolebinding.rbac.authorization.k8s.io/horizon-source-controller-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/horizon-source-webhook-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/horizon-source-controller-addressable-resolver created
clusterrole.rbac.authorization.k8s.io/horizon-source-webhook created
customresourcedefinition.apiextensions.k8s.io/horizonsources.sources.tanzu.vmware.com created
customresourcedefinition.apiextensions.k8s.io/vspherebindings.sources.tanzu.vmware.com created
customresourcedefinition.apiextensions.k8s.io/vspheresources.sources.tanzu.vmware.com created
service/horizon-source-controller-manager created
service/horizon-source-webhook created
service/vsphere-source-webhook created
deployment.apps/horizon-source-controller created
mutatingwebhookconfiguration.admissionregistration.k8s.io/defaulting.webhook.horizon.sources.tanzu.vmware.com created
validatingwebhookconfiguration.admissionregistration.k8s.io/validation.webhook.horizon.sources.tanzu.vmware.com created
validatingwebhookconfiguration.admissionregistration.k8s.io/config.webhook.horizon.sources.tanzu.vmware.com created
secret/webhook-certs created
deployment.apps/horizon-source-webhook created
mutatingwebhookconfiguration.admissionregistration.k8s.io/defaulting.webhook.vsphere.sources.tanzu.vmware.com created
validatingwebhookconfiguration.admissionregistration.k8s.io/validation.webhook.vsphere.sources.tanzu.vmware.com created
validatingwebhookconfiguration.admissionregistration.k8s.io/config.webhook.vsphere.sources.tanzu.vmware.com created
secret/vsphere-webhook-certs created
mutatingwebhookconfiguration.admissionregistration.k8s.io/vspherebindings.webhook.vsphere.sources.tanzu.vmware.com created
deployment.apps/vsphere-source-webhook created
configmap/config-logging created
configmap/config-observability created
```

## Configuration Source Type VSphereSource

The `VSphereSource` provides a simple mechanism to enable users to react to
vSphere events.

### Event Provider Type vCenter Server

VMware vCenter Server is an advanced server management software that provides a
centralized platform for controlling your VMware vSphere environments, allowing
you to automate and deliver a virtual infrastructure across the hybrid cloud
with confidence.

Since VMware vCenter Server event types are environment
specific (vSphere version, extensions), a list of events for vCenter as an event
source can be found on the following Github repository - [William Lam -
vcenter-event-mapping](https://github.com/lamw/vcenter-event-mapping).

A new `VSphereSource` can be configured in two ways:
1. Using the above mentioned `kn-vsphere` plugin
2. Applying a `VSphereSource` manifest file (yaml) using `kubectl`

### Create a new VSphereSource via CLI

The creation of a Kubernetes `basic-auth` secret is required in order to create
a new `VSphereSource`.

```console
export VCENTER_USERNAME='read-only-user@vsphere.local' \
export VCENTER_PASSWORD='my-secret-pwd' \
export VCENTER_HOSTNAME='vcsa.mydomain.com'
```

The new `kn-vsphere` plugin for the Knative CLI (`kn`) can be used for this
task.

Creating the new secret in the namespace `vmware-system`:

```console
kn vsphere auth create \
--namespace vmware-system \
--username $VCENTER_USERNAME \
--password $VCENTER_PASSWORD \
--name vcsa-ro-creds \
--verify-url https://$VCENTER_HOSTNAME \
--verify-insecure
```

The created secret stores your sensible data:

```console
kubectl -n vmware-system get secret vcsa-ro-creds -oyaml
```

```yaml
apiVersion: v1
data:
  password: <base64-encoded-data>
  username: <base64-encoded-data>
kind: Secret
metadata:
  name: vcsa-ro-creds
  namespace: vmware-system
type: kubernetes.io/basic-auth
```

Create the new `VSphereSource` by also using the `kn-vsphere` plugin. Make sure
to not only provide valid vCenter data but also the correct Knative Sink data.
If the Sink is a Broker, retrieve the correct URI via `kn broker list -A`:

```console
kn broker list -A

NAMESPACE          NAME      URL                                                                AGE    CONDITIONS   READY   REASON
vmware-functions   default   http://default-broker-ingress.vmware-functions.svc.cluster.local   167d   8 OK / 8     True
```

Create the new `VSphereSource` in the `vmware-system` namespace as well:

```console
kn vsphere source create \
--namespace vmware-system \
--name vcsa-source \
--vc-address https://$VCENTER_HOSTNAME \
--skip-tls-verify \
--secret-ref vcsa-ro-creds \
--sink-uri http://default-broker-ingress.vmware-functions.svc.cluster.local \
--encoding json

Created source
```

First indications of its functionality can be validated by checking the logs:

```console
kubectl -n vmware-system logs vcsa-source-adapter-58c744d8db-ttddd

{"level":"info","ts":"2023-12-01T10:28:30.955Z","logger":"vsphere-source-adapter","caller":"vsphere/adapter.go:312","msg":"setting begin of event stream","commit":"8fda92a-dirty","beginTimestamp":"2023-12-01 10:28:30.950583 +0000 UTC"}
{"level":"info","ts":"2023-12-01T10:33:30.920Z","logger":"vsphere-source-adapter","caller":"vsphere/client.go:115","msg":"Executing SOAP keep-alive handler","commit":"8fda92a-dirty","rpc":"keepalive"}
{"level":"info","ts":"2023-12-01T10:33:30.939Z","logger":"vsphere-source-adapter","caller":"vsphere/client.go:121","msg":"vCenter current time: 2023-12-01 10:33:30.930277 +0000 UTC","commit":"8fda92a-dirty","rpc":"keepalive"}
{"level":"info","ts":"2023-12-01T10:38:30.940Z","logger":"vsphere-source-adapter","caller":"vsphere/client.go:115","msg":"Executing SOAP keep-alive handler","commit":"8fda92a-dirty","rpc":"keepalive"}
{"level":"info","ts":"2023-12-01T10:38:30.968Z","logger":"vsphere-source-adapter","caller":"vsphere/client.go:121","msg":"vCenter current time: 2023-12-01 10:38:30.965905 +0000 UTC","commit":"8fda92a-dirty","rpc":"keepalive"}
```

### Create a new VSphereSource via Manifest File

The following table lists allowed and required fields for connecting to a
vCenter Server and the respective type values and examples for these fields.

| Field           | Type    | Description                                                                                           | Required | Example                          |
|-----------------|---------|-------------------------------------------------------------------------------------------------------|----------|----------------------------------|
| `address`       | String  | URI of the VMware vCenter Server                                                                      | true     | `https://$VCENTER_HOSTNAME`       |
| `skipTLSVerify`   | Boolean | Skip TSL verification                                                                                 | true     | `true` (i.e. ignore errors)      |
| `checkpointConfig`    | String | Configure checkpointing using options `maxAgeSeconds` and `periodSeconds`.                       | true     | Defaults are `maxAgeSeconds: 300` and `periodSeconds: 10`                          |
| `payloadEncoding` | String | Set the CloudEvent data encoding scheme to e.g. `json` or `xml` | true    | `payloadEncoding: application/json`         |
| `secretRef`        | Object  | `secretRef` holds the name of the Kubernetes secret which is used by the source to authenticate against it.  | true     | see section [Authentication with vSphere](https://github.com/vmware-tanzu/sources-for-knative/tree/main#authenticating-with-vsphere) |
| `sink`        | Object  | Where to send the events.  | true     | Default for VEBA is the adress of the RabbitMQ broker: `http://default-broker-ingress.vmware-functions.svc.cluster.local` |

The following `vspheresource.yaml` sample shows the configuration of the `VSphereSource` which got created above using the `kn vsphere` cli.

```yaml
apiVersion: sources.tanzu.vmware.com/v1alpha1
kind: VSphereSource
metadata:
  name: vcsa-source
  namespace: vmware-system
spec:
  address: https://vcsa.mydomain.com
  checkpointConfig:
    maxAgeSeconds: 300
    periodSeconds: 10
  payloadEncoding: application/json
  secretRef:
    name: vcsa-ro-creds
  sink:
    uri: http://default-broker-ingress.vmware-functions.svc.cluster.local
  skipTLSVerify: true
```

## Configuration Source Type HorizonSource

The `HorizonSource` provides a simple mechanism to enable users to react to
Horizon events.

### Event Provider Type Horizon

VMware Horizon is a platform for delivering virtual desktops and apps
efficiently and securely across hybrid cloud for the best end-user digital
workspace experience.

This provider supports all audit events (model `AuditEventSummary`) exposed
through the Horizon REST API starting with
[version](https://code.vmware.com/apis/1169/view-rest-api#/External/listAuditEvents)
`2106`.

The following table lists allowed and required fields for connecting to the
Horizon REST API and the respective type values and examples for these fields.

| Field         | Type    | Description                 | Required | Example                                |
|---------------|---------|-----------------------------|----------|----------------------------------------|
| `address`     | String  | URI of the Horizon REST API | true     | `https://myhorizon.corp.local`     |
| `skipTLSVerify` | Boolean | Skip TSL verification       | true     | `true` (i.e. ignore errors)            |
| `secretRef`      | Object  | secretRef holds the name of the Kubernetes secret which is used by the source to authenticate against it.  | true     | See `secret` example below. |
| `sink`        | Object  | Where to send the events.  | true     | Default for VEBA is the adress of the RabbitMQ broker: `http://default-broker-ingress.vmware-functions.svc.cluster.local` |

Create a Kubernetes `Secret` as per the name under `secretRef` in the
`HorizonSource` above which holds the required Horizon credentials. `domain`,
`username` and `password` are required fields. Replace the field values
accordingly.

```shell
kubectl create secret generic horizon-credentials --from-literal=domain="example.com" --from-literal=username="horizon-source-account" --from-literal=password='ReplaceMe'
```

Modify the `HorizonSource` example according to your environment.

`address` is the HTTPs endpoint of the Horizon API server. To skip TLS and
certificate verification, set `skipTLSVerify` to `true`.

Change the values under `sink` to match your Knative Eventing environment.

```yaml
apiVersion: sources.tanzu.vmware.com/v1alpha1
kind: HorizonSource
metadata:
  name: horizon-example
spec:
  sink:
    ref:
      apiVersion: eventing.knative.dev/v1
      kind: Broker
      name: example-broker
      namespace: default
  address: https://horizon.server.example.com
  skipTLSVerify: false
  secretRef:
    name: horizon-credentials
  serviceAccountName: horizon-source-sa
```

If the specified `serviceAccountName` does not exist, it will be created
automatically.

```console
kubectl get serviceaccount

NAME                SECRETS   AGE
default             0         5h7m
horizon-source-sa   0         152m
```

Verify the new created `HorizonSource`:

```console
kn source list

NAME              TYPE            RESOURCE                                  SINK                   READY
horizon-example   HorizonSource   horizonsources.sources.tanzu.vmware.com   broker:in-mem-broker   True
```

```console
kubectl get horizonsources

NAME              SOURCE                   SINK        READY                                                                            REASON
horizon-example   https://horizon.server.example.com   http://broker-ingress.knative-eventing.svc.cluster.local/default/in-mem-broker   True
```

The following example shows a Horizon `userloggedin` event displayed via the
event viewer application Sockeye (default in project VEBA).

```json
[...]

  specversion: 1.0
  type: com.vmware.horizon.vlsi_userloggedin_rest.v0
  source: https://horizon.server.example.com
  id: 3405990
  time: 2023-12-06T13:48:05Z
  datacontenttype: application/json
Extensions,
  knativearrivaltime: 2023-12-06T13:48:06.324666167Z
Data,
2023/12/06 13:48:06 Broadasting to 0 clients: {"data":{"id":3405990,"machine_dns_name":"horizon.server.example.com","message":"User example.com\\horizonedge has logged in to Horizon REST API","module":"Vlsi","severity":"AUDIT_SUCCESS","time":1701870485733,"type":"VLSI_USERLOGGEDIN_REST","user_id":"S-1-5-21-2404106297-684390199-568881881-4197"},"datacontenttype":"application/json","id":"3405990","knativearrivaltime":"2023-12-06T13:48:06.324666167Z","source":"https://horizon.server.example.com","specversion":"1.0","time":"2023-12-06T13:48:05Z","type":"com.vmware.horizon.vlsi_userloggedin_rest.v0"}
  {
    "id": 3405990,
    "machine_dns_name": "horizon.server.example.com",
    "message": "User example.com\\horizonedge has logged in to Horizon REST API",
    "module": "Vlsi",
    "severity": "AUDIT_SUCCESS",
    "time": 1701870485733,
    "type": "VLSI_USERLOGGEDIN_REST",
    "user_id": "S-1-5-21-2404106297-684390199-568881881-4197"
  }

[...]
```

## Event Viewer Application Sockeye

In order to provide users with the ability to display incoming events from a
source, the VEBA project is using the open-source event viewer application
[Sockeye](https://github.com/n3wscott/sockeye). When using the appliance
form factor the event viewer is accessible via `https://veba-fqdn/events`.

The manual installation of Sockeye will be explained in section [Deploy Tanzu Sources to KinD](./deploy-tanzu-sources-kind.md).


The following example shows a vSphere `DrsVmPoweredOn` event displayed via Sockeye.

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

## Troubleshooting

All components follow the Knative logging convention.
The log level (`debug`, `info`, `error`, etc. is configurable per component,
e.g. `vsphere-source-webhook`, `VSphereSource` adapter, etc.

The default logging level is `info`.

The log level for adapters, e.g. a particular `VSphereSource` `deployment` can
be changed at runtime via the `config-logging` `ConfigMap` which is created
when deploying the Tanzu Sources for Knative.

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
