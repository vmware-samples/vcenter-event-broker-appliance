---
layout: docs
toc_id: intro-event-router
title: VMware Event Router - Introduction
description: VMware Event Router Introduction
permalink: /kb/event-router
cta:
 title: Get Started
 description: Explore the capabilities that the VMware Event Router enables
 actions:
    - text: Install the [Appliance with Knative](install-knative) to extend your SDDC with our [community-sourced functions](/examples)
    - text: Learn more about the [Events in vCenter](vcenter-events) and how to find the right event for your use case
    - text: Learn more about Functions in this overview [here](functions).
---

# Introduction to VMware Event Router

The VMware Event Router is used to connect to various VMware event `providers`
(i.e. "sources") and forward these events to different event `processors` (i.e.
"sinks"). This project is currently used by the [_VMware Event Broker
Appliance_](https://www.vmweventbroker.io/) as the core logic to forward
[CloudEvents](https://cloudevents.io/), e.g. from vSphere, to configurable event
`processors` (see below).

**Supported Event `Providers`:**

- [VMware vCenter Server](https://www.vmware.com/products/vcenter-server.html)
- [VMware Horizon](https://www.vmware.com/products/horizon.html)
- Generic [CloudEvents](https://cloudevents.io/) Webhook
- vCenter Simulator [vcsim](https://github.com/vmware/govmomi/tree/master/vcsim)
  (deprecated, see note [below](#provider-type-vcsim))

**Supported Event `Processors`:**

- [Knative](https://knative.dev/)
- [OpenFaaS](https://www.openfaas.com/)
- [AWS EventBridge](https://aws.amazon.com/eventbridge/?nc1=h_ls)

The VMware Event Router uses the [CloudEvents](https://cloudevents.io/) standard
to normalize events from the supported event `providers`. See
[below](#example-event-structure) for an example.

**Event `Provider` Delivery Guarantees:**

- At-least-once event delivery
  -  with the [vCenter event provider](#provider-type-vcenter) option `checkpoint: true`
- At-most-once event delivery
  - with the [vCenter event provider](#provider-type-vcenter) option `checkpoint: false`
  - with the [Webhook event provider](#provider-type-webhook)
  - with the [Horizon event provider](#provider-type-horizon)
  - with the [vcsim event provider](#provider-type-vcsim)

> **Note:** All implemented event `processors` use built-in retry mechanisms so
your function might still be involved multiple times depending on its response
code. However, if an event `provider` crashes before sending an event to the
configured `processor` or when the `processor` returns an error, the event is
not retried and discarded.

**Current limitations:**

- Only one event `provider` and one event `processor` can be configured at a
  time (see note below)
- At-least-once event delivery semantics are not guaranteed if the event
  router crashes **within seconds** right after startup and having received *n* events but before creating the
  first valid checkpoint (current checkpoint interval is 5s)
- If an event cannot be successfully delivered (retried) by an event `processor` it is
  logged and discarded, i.e. there is currently no support for [dead letter
  queues](https://en.wikipedia.org/wiki/Dead_letter_queue) (see note below)
- Retries in the [OpenFaaS event processor](#processor-type-openfaas) are only
  supported when running in synchronous mode, i.e. `async: false` (see this
  OpenFaaS [issue](https://github.com/openfaas/nats-queue-worker/issues/84))

> **Note:** It is possible though to run **multiple instances** of the event
> router with different configurations to address multi-vCenter scenarios. This
> decision was made for scalability and resource/tenancy isolation purposes.

> **Note:** Event Processors, like Knative, support Dead Letter Queues when
> using `Broker` mode.

<!-- omit in toc -->
## Table of Contents
- [Introduction to VMware Event Router](#introduction-to-vmware-event-router)
- [Configuration](#configuration)
  - [Overview: Configuration File Structure (YAML)](#overview-configuration-file-structure-yaml)
  - [JSON Schema Validation](#json-schema-validation)
  - [API Version, Kind and Metadata](#api-version-kind-and-metadata)
  - [The `eventProvider` section](#the-eventprovider-section)
    - [Provider Type `vcenter`](#provider-type-vcenter)
    - [Provider Type `horizon`](#provider-type-horizon)
    - [Provider Type `webhook`](#provider-type-webhook)
    - [Provider Type `vcsim`](#provider-type-vcsim)
  - [The `eventProcessor` section](#the-eventprocessor-section)
    - [Processor Type `knative`](#processor-type-knative)
      - [Destination](#destination)
    - [Processor Type `openfaas`](#processor-type-openfaas)
    - [Processor Type `aws_event_bridge`](#processor-type-aws_event_bridge)
  - [The `auth` section](#the-auth-section)
    - [Type `basic_auth`](#type-basic_auth)
    - [Type `aws_access_key`](#type-aws_access_key)
    - [Type `aws_iam_role`](#type-aws_iam_role)
    - [Type `active_directory`](#type-active_directory)
  - [The `metricsProvider` section](#the-metricsprovider-section)
    - [Provider Type `default`](#provider-type-default)
- [Deployment](#deployment)
  - [Assisted Deployment](#assisted-deployment)
    - [Helm Deployment](#helm-deployment)
      - [Option 1: Configuration with Knative](#option-1-configuration-with-knative)
      - [Option 2: Configuration with OpenFaaS](#option-2-configuration-with-openfaas)
      - [Deploy the VMware Event Router Helm Chart](#deploy-the-vmware-event-router-helm-chart)
      - [Creating/Updating the Chart](#creatingupdating-the-chart)
    - [Manual Deployment](#manual-deployment)
      - [Create Knative ClusterRoleBinding (skip if not using Knative)](#create-knative-clusterrolebinding-skip-if-not-using-knative)
      - [Create the VMware Event Router Deployment](#create-the-vmware-event-router-deployment)
  - [CLI Flags](#cli-flags)

# Configuration

The VMware Event Router can be run standalone (statically linked binary) or
deployed as a Docker container, e.g. in a Kubernetes environment. See
[deployment](#deployment) for further instructions. The configuration of event
`providers` and `processors` and other internal components (such as metrics) is
done via a YAML file passed in via the `-config` command line flag.

```
 _    ____  ___                            ______                 __     ____              __
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/


Usage of ./vmware-event-router:

  -config string
        path to configuration file (default "/etc/vmware-event-router/config")
  -log-json
        print JSON-formatted logs
  -log-level string
        set log level (debug,info,warn,error) (default "info")

commit: <git_commit_sha>
version: <release_tag>
```

The following sections describe the layout of the configuration file (YAML) and
specific options for the supported event `providers`, `processors` and `metrics`
endpoint. Configuration examples are provided [here](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/vmware-event-router/deploy).

> **Note:** Currently only one event `provider` and one event `processor` can be
> configured at a time, e.g. one vCenter Server instance streaming events to
> OpenFaaS **or** AWS EventBridge. It is possible to run multiple instances of
> the event router with different configurations to address
> multi-provider/processor scenarios.

## Overview: Configuration File Structure (YAML)

The following file, using `vcenter` as the event `provider` and `knative` as
the `processor` shows an example of the configuration file syntax:

```yaml
apiVersion: event-router.vmware.com/v1alpha1
kind: RouterConfig
metadata:
  name: router-config-openfaas
  labels:
    key: value
eventProvider:
  type: vcenter
  name: veba-demo-vc-01
  vcenter:
    address: https://my-vcenter01.domain.local/sdk
    insecureSSL: false
    checkpoint: true
    auth:
      type: basic_auth
      basicAuth:
        username: administrator@vsphere.local
        password: ReplaceMe
eventProcessor:
  type: knative
  name: veba-demo-knative
  knative:
    encoding: binary
    insecureSSL: false
    destination:
      ref:
        apiVersion: eventing.knative.dev/v1
        kind: Broker
        name: rabbit
        namespace: default
metricsProvider:
  type: default
  name: veba-demo-metrics
  default:
    bindAddress: "0.0.0.0:8082"
```

## JSON Schema Validation

In order to simplify the configuration and validation of the YAML configuration
file a JSON schema [file](https://github.com/vmware-samples/vcenter-event-broker-appliance/blob/master/vmware-event-router/routerconfig.schema.json) is provided. Many editors/IDEs offer
support for registering a schema file, e.g.
[Jetbrains](https://www.jetbrains.com/help/rider/Settings_Languages_JSON_Schema.html)
and [VS
Code](https://code.visualstudio.com/docs/languages/json#_json-schemas-and-settings).

> **Note:** The schema file can be downloaded and provided via a local file
> location or (recommended) via a direct URL, e.g. Github
> [raw](https://help.data.world/hc/en-us/articles/115006300048-GitHub-how-to-find-the-sharable-download-URL-for-files-on-GitHub)
> URL pointing to the aforementioned JSON schema file.

## API Version, Kind and Metadata

The following table lists allowed and required fields with their respective type
values and examples for these fields.

| Field             | Type              | Description                                     | Required | Example                            |
|-------------------|-------------------|-------------------------------------------------|----------|------------------------------------|
| `apiVersion`      | String            | API Version used for this configuration file    | true     | `event-router.vmware.com/v1alpha1` |
| `kind`            | String            | Type of this API resource                       | true     | `RouterConfig`                     |
| `metadata`        | Object            | Additional metadata for this configuration file | true     |                                    |
| `metadata.name`   | String            | Name of this configuration file                 | true     | `config-vc-openfaas-PROD`          |
| `metadata.labels` | map[String]String | Optional key/value pairs                        | false    | `env: PROD`                        |

## The `eventProvider` section

The following table lists allowed and required fields with their respective type
values and examples for these fields.

| Field             | Type   | Description                            | Required | Example                                    |
|-------------------|--------|----------------------------------------|----------|--------------------------------------------|
| `type`            | String | Type of the event provider             | true     | `vcenter`                                  |
| `name`            | String | Name identifier for the event provider | true     | `vc-01-PROD`                               |
| `<provider_type>` | Object | Provider specific configuration        | true     | (see specific provider type section below) |

### Provider Type `vcenter`

VMware vCenter Server is advanced server management software that provides a
centralized platform for controlling your VMware vSphere environments, allowing
you to automate and deliver a virtual infrastructure across the hybrid cloud
with confidence. Since VMware vCenter Server event types are environment specific (vSphere version,
extensions), a list of events for vCenter as an event source can be generated as
described in this [blog post](https://www.williamlam.com/2019/12/listing-all-events-for-vcenter-server.html).

The following table lists allowed and required fields for connecting to a
vCenter Server and the respective type values and examples for these fields.

| Field           | Type    | Description                                                                                           | Required | Example                          |
|-----------------|---------|-------------------------------------------------------------------------------------------------------|----------|----------------------------------|
| `address`       | String  | URI of the VMware vCenter Server                                                                      | true     | `https://10.0.0.1:443/sdk`       |
| `insecureSSL`   | Boolean | Skip TSL verification                                                                                 | true     | `true` (i.e. ignore errors)      |
| `checkpoint`    | Boolean | Configure checkpointing via checkpoint file for event recovery/replay purposes                        | true     | `true`                           |
| `checkpointDir` | Boolean | **Optional:** Configure an alternative location for persisting checkpoints (default: `./checkpoints`) | false    | `/var/local/checkpoints`         |
| `<auth>`        | Object  | vCenter credentials                                                                                   | true     | (see `basic_auth` example below) |

### Provider Type `horizon`

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
| `address`     | String  | URI of the Horizon REST API | true     | `https://api.myhorizon.corp.local`     |
| `insecureSSL` | Boolean | Skip TSL verification       | true     | `true` (i.e. ignore errors)            |
| `<auth>`      | Object  | Horizon domain credentials  | true     | (see `active_directory` example below) |

### Provider Type `webhook`

The `webhook` event provider listens for incoming
[CloudEvents](https://cloudevents.io/) (binary or structured mode) on a
configurable HTTP server in the VMware Event Router. The HTTP method used to
send the CloudEvent must be `POST`.

The following table lists allowed and required fields for setting up a webhook
server.

| Field         | Type   | Description                                                                    | Required | Example                          |
|---------------|--------|--------------------------------------------------------------------------------|----------|----------------------------------|
| `bindAddress` | String | TCP/IP socket and port to listen on (**do not** add any URI scheme or slashes) | true     | `0.0.0.0:8080`                   |
| `path`        | String | Webhook endpoint path (must not be `/`)                                        | true     | `/webhook`                       |
| `<auth>`      | Object | Configure `basic_auth` for incoming requests                                   | false    | (see `basic_auth` example below) |

**Note:** When the VMware Event Router log level is `DEBUG` incoming webhook
requests (method, path, headers, remote address) will be logged.

### Provider Type `vcsim`

⚠️ This provider is **deprecated** and will be removed in future versions. The
`vcenter` provider will work correctly against a `vcsim` instance.

The following table lists allowed and required fields for connecting to the
govmomi vCenter Simulator
[vcsim](https://github.com/vmware/govmomi/tree/master/vcsim) and the respective
type values and examples for these fields.

| Field         | Type    | Description                           | Required | Example                          |
|---------------|---------|---------------------------------------|----------|----------------------------------|
| `address`     | String  | URI of the govmomi vCenter simulator  | true     | `https://127.0.0.1:8989/sdk`     |
| `insecureSSL` | Boolean | Skip TSL verification                 | true     | `true` (i.e. ignore errors)      |
| `<auth>`      | Object  | govmomi vCenter simulator credentials | true     | (see `basic_auth` example below) |

> **Note:** This event provider has some limitations and currently does not
> behave like a "real" vCenter Server event stream, e.g. see issue
> [#2134](https://github.com/vmware/govmomi/issues/2134). This provider is for
> prototyping/testing purposes only.

## The `eventProcessor` section

The following table lists allowed and required fields with their respective type
values and examples for these fields.

| Field              | Type   | Description                             | Required | Example                                     |
|--------------------|--------|-----------------------------------------|----------|---------------------------------------------|
| `type`             | String | Type of the event processor             | true     | `knative`, `openfaas` or `aws_event_bridge` |
| `name`             | String | Name identifier for the event processor | true     | `knative-broker-PROD`                       |
| `<processor_type>` | Object | Processor specific configuration        | true     | (see specific processor type section below) |

### Processor Type `knative`

Knative is a Kubernetes-based platform to deploy and manage modern serverless
workloads. Knative has two core building blocks, that is Serving (Knative
`Service`) and Eventing (`Broker`, `Channel`, etc.). 

The VMware Event Router can be configured to directly send events to any
*addressable* Knative resource ("reference"), e.g. a Knative `Broker` or
`Service`. `Broker` is the recommended deployment model for the VMware Event
Router. Please see the Knative documentation on
[Eventing](https://knative.dev/docs/eventing/) for details around brokers,
triggers, event filtering, etc.

Alternatively, the router can send events to a URI, e.g. an external HTTP endpoint accepting
CloudEvents.

The following table lists allowed and optional fields for using Knative as an
event `processor`.

| Field           | Type    | Description                                                                                                                                                | Required | Example                           |
|-----------------|---------|------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|-----------------------------------|
| `<destination>` | Object  | Knative *addressable* Destination to send events to.                                                                                                       | true     | (see `destination` section below) |
| `encoding`      | Boolean | Cloud Events message [encoding](https://github.com/cloudevents/spec/blob/v1.0/spec.md#message)                                                             | true     | `structured` or `binary`          |
| `insecureSSL`   | Boolean | Skip TSL verification                                                                                                                                      | true     | `true` (i.e. ignore errors)       |
| `<auth>`        | Object  | **Optional:** authentication data (see auth section below). Omit section if your Knative Service, URI (Ingress) or Broker does not enforce authentication. | false    | (see `basic_auth` example below)  |

> **Note:** When sending events to a Knative `Broker`, the Knative broker will always
> send `binary` encoded cloud events to the Knative sinks, e.g. triggered `Service`.

#### Destination

Knative abstracts event receivers ("sinks") in a very flexible way via
*addressable* `Destinations`. The VMware Event Router technically supports all
Knative `Destinations`. However, Knative `Broker` or Knative `Services` are
likely the ones you will be using.

> **Note:** We have not done any testing for Knative `Channels`, since
> request response style interactions are out of scope for the VMware Event
> Router.

The following table lists the available destination types for the `destination` section in
the router configuration when targeting a URI.

| Field        | Type   | Description                                                    | Required              | Example                       |
|--------------|--------|----------------------------------------------------------------|-----------------------|-------------------------------|
| `<uri>`      | Object | URI can be used to send events to an external HTTP(S) endpoint | one_of `uri` or `ref` |                               |
| `uri.scheme` | String | URI scheme                                                     | `true`                | `http` or `https`             |
| `uri.host`   | String | URI target host                                                | `true`                | `gateway-external.corp.local` |

The following table lists the available destination types for the `destination` section in
the router configuration when targeting a URI.

| Field            | Type   | Description                                                                           | Required              | Example                                               |
|------------------|--------|---------------------------------------------------------------------------------------|-----------------------|-------------------------------------------------------|
| `<ref>`          | Object | Ref can be used to send events to a *addressable* Kubernetes or Knative target (sink) | one_of `uri` or `ref` |                                                       |
| `ref.apiVersion` | String | Kubernetes API version of the target                                                  | `true`                | `eventing.knative.dev/v1` or `serving.knative.dev/v1` |
| `ref.kind`       | String | Kubernetes Kind of the target                                                         | `true`                | `Broker` or `Service`                                 |
| `ref.name`       | String | Kubernetes object name for the given `kind` and `apiVersion`                          | `true`                | `mybroker`                                            |
| `ref.namespace`  | String | Kubernetes namespace for the given object reference                                   | `true`                | `default`                                             |


> **Note:** Only one of `uri` or `ref` MUST be specified. The list of all
> available URI options is documented [here](https://pkg.go.dev/net/url#URL).

> **Note:** A vanilla Kubernetes `Service` can also be specified when using
> `Ref` as the destination type. Use `v1` in the `apiVersion` section.


### Processor Type `openfaas`

OpenFaaS functions can subscribe to the event stream via function `"topic"`
annotations in the function stack configuration (see OpenFaaS documentation for
details on authoring functions), e.g.:

```yaml
annotations:
  topic: "VmPoweredOnEvent,VmPoweredOffEvent"
```

> **Note:** One or more event categories can be specified, delimited via `","`.
> A list of event names (categories) and how to retrieve them can be found
> [here](https://github.com/lamw/vcenter-event-mapping/blob/master/vsphere-6.7-update-3.md).
> A simple "echo" function useful for testing is provided
> [here](https://github.com/embano1/of-echo/blob/master/echo.yml).

The following table lists allowed and optional fields for using OpenFaaS as an
event `processor`.

| Field     | Type    | Description                                                                                                       | Required | Example                                          |
|-----------|---------|-------------------------------------------------------------------------------------------------------------------|----------|--------------------------------------------------|
| `address` | String  | URI of the OpenFaaS gateway                                                                                       | true     | `http://gateway.openfaas:8080`                   |
| `async`   | Boolean | Specify how to invoke functions (synchronously or asynchronously)                                                 | true     | `false` (i.e. use sync function invocation mode) |
| `<auth>`  | Object  | **Optional:** authentication data (see auth section below). Omit section if OpenFaaS gateway auth is not enabled. | false    | (see `basic_auth` example below)                 |

### Processor Type `aws_event_bridge`

Amazon EventBridge is a serverless event bus that makes it easy to connect
applications together using data from your own applications, integrated
Software-as-a-Service (SaaS) applications, and AWS services. In order to reduce
bandwidth and costs (number of events ingested, see
[pricing](https://aws.amazon.com/eventbridge/pricing/)), VMware Event Router
only forwards events configured in the associated `rule` of an event bus. Rules
in AWS EventBridge use pattern matching
([docs](https://docs.aws.amazon.com/eventbridge/latest/userguide/filtering-examples-structure.html)).
Upon start, VMware Event Router contacts EventBridge (using the given IAM role)
to parse the configured rule ARN (see configuration option below).

The VMware Event Router uses the pattern match library which supports a subset
of the EventBridge pattern rules. You may only use these supported patterns in
your specified EventBridge rule. Refer to [this
page](https://github.com/timbray/quamina/blob/v0.2.0/PATTERNS.md) for the
currently supported patterns in `Quamina`.

> **Note:** EventBridge wraps each VMware Event Router event (CloudEvent) into
> an EventBridge message envelop. The `detail` field contains the JSON
> representation of the full CloudEvent as produced by the VMware Event Router.

The following examples show supported and useful patterns.

Example: Forward all CloudEvents containing one of the specified `subjects`:

```json
{
  "detail": {
    "subject": ["VmPoweredOnEvent", "VmPoweredOffEvent", "DrsVmPoweredOnEvent"]
  }
}
```

Example: Forward all CloudEvents containing a `subject` with the prefix `Vm`:

```json
{
  "detail": {
    "subject": [{
      "shellstyle": "Vm*"
    }]
  }
}
```

Example: Forward all CloudEvents containing virtual machines with the prefix
`Linux`:

```json
{
  "detail": {
    "data": {
      "Vm": {
        "Name": [{
          "shellstyle": "Linux*"
        }]
      }
    }
  }
}
```

> **Note:** A list of event names (categories) and how to retrieve them can be
> found
> [here](https://github.com/lamw/vcenter-event-mapping/blob/master/vsphere-6.7-update-3.md).

The following table lists allowed and optional fields for using AWS EventBridge
as an event `processor`.

| Field      | Type   | Description                                                                                                                             | Required | Example                                                                |
|------------|--------|-----------------------------------------------------------------------------------------------------------------------------------------|----------|------------------------------------------------------------------------|
| `region`   | String | AWS region to use, see [regions doc](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html). | true     | `us-west-1`                                                            |
| `eventBus` | String | Name of the event bus to use                                                                                                            | true     | `default` or `arn:aws:events:us-west-1:1234567890:event-bus/customBus` |
| `ruleARN`  | String | Rule ARN to use for event pattern matching                                                                                              | true     | `arn:aws:events:us-west-1:1234567890:rule/vmware-event-router`         |
| `<auth>`   | Object | AWS IAM role credentials                                                                                                                | true     | (see `aws_access_key` and `aws_iam_role` examples below)               |

## The `auth` section

The following table lists allowed and required fields with their respective type
values and examples for these fields. Since the various `processors` and
`providers` use different authentication mechanisms (or none at all) this
section describes the various options.

### Type `basic_auth`

Supported providers/processors:

- `vcenter` (required: `true`)
- `vcsim` (required: `true`)
- `openfaas` (required: `false`, i.e. optional)
- `default` metrics server (see below) (required: `false`, i.e. optional)

| Field                | Type   | Description                             | Required | Example      |
|----------------------|--------|-----------------------------------------|----------|--------------|
| `type`               | String | Authentication method to use            | true     | `basic_auth` |
| `basicAuth`          | Object | Use when `basic_auth` type is specified | true     |              |
| `basicAuth.username` | String | Username                                | true     | `admin`      |
| `basicAuth.password` | String | Password                                | true     | `P@ssw0rd`   |

### Type `aws_access_key`

Use an AWS IAM role with the provided access key ID and secret access key for
authentication.

Supported providers/processors:

- `aws_event_bridge`

| Field                        | Type   | Description                                 | Required | Example          |
|------------------------------|--------|---------------------------------------------|----------|------------------|
| `type`                       | String | Authentication method to use                | true     | `aws_access_key` |
| `awsAccessKeyAuth`           | Object | Use when `aws_access_key` type is specified | true     |                  |
| `awsAccessKeyAuth.accessKey` | String | Access Key ID for the IAM role used         | true     | `ABCDEFGHIJK`    |
| `awsAccessKeyAuth.secretKey` | String | Secret Access Key for the IAM role used     | true     | `ZYXWVUTSRQPO`   |

> **Note:** Please follow the EventBridge IAM [user
> guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/getting-set-up-eventbridge.html)
> before deploying the event router. Further information can also be found in
> the
> [authentication](https://docs.aws.amazon.com/eventbridge/latest/userguide/auth-and-access-control-eventbridge.html#authentication-eventbridge)
> section.

In addition to the recommendation in the AWS EventBridge user guide you might
want to lock down the IAM role for the VMware Event Router and scope it to these
permissions ("Action"):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "VisualEditor0",
      "Effect": "Allow",
      "Action": [
        "events:PutEvents",
        "events:ListRules",
        "events:TestEventPattern"
      ],
      "Resource": "*"
    }
  ]
}
```

### Type `aws_iam_role`

Use an AWS IAM role configured from the shared credentials file.

Supported providers/processors:

- `aws_event_bridge`

| Field  | Type   | Description                  | Required | Example        |
|--------|--------|------------------------------|----------|----------------|
| `type` | String | Authentication method to use | true     | `aws_iam_role` |

> **Note:** Please follow the EventBridge IAM [user
> guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/getting-set-up-eventbridge.html)
> before deploying the event router. Further information can also be found in
> the
> [authentication](https://docs.aws.amazon.com/eventbridge/latest/userguide/auth-and-access-control-eventbridge.html#authentication-eventbridge)
> section.

In addition to the recommendation in the AWS EventBridge user guide you might
want to lock down the IAM role for the VMware Event Router and scope it to these
permissions ("Action"):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "VisualEditor0",
      "Effect": "Allow",
      "Action": [
        "events:PutEvents",
        "events:ListRules",
        "events:TestEventPattern"
      ],
      "Resource": "*"
    }
  ]
}
```

### Type `active_directory`

Supported providers/processors:

- `horizon` (required: `true`)

| Field                          | Type   | Description                                   | Required | Example            |
|--------------------------------|--------|-----------------------------------------------|----------|--------------------|
| `type`                         | String | Authentication method to use                  | true     | `active_directory` |
| `activeDirectoryAuth`          | Object | Use when `active_directory` type is specified | true     |                    |
| `activeDirectoryAuth.domain`   | String | Domain                                        | true     | `corp`             |
| `activeDirectoryAuth.username` | String | Username                                      | true     | `administrator`    |
| `activeDirectoryAuth.password` | String | Password                                      | true     | `P@ssw0rd`         |

> **Note:** UPN authentication, e.g. `administrator@corp.local` as `username`,
> is not supported.

## The `metricsProvider` section

The VMware Event Router currently only exposes a default ("internal" or "embedded") metrics
endpoint. In the future, support for more providers is planned, e.g. Wavefront,
Prometheus, etc.

| Field             | Type   | Description                     | Required | Example                                 |
|-------------------|--------|---------------------------------|----------|-----------------------------------------|
| `type`            | String | Type of the metrics provider    | true     | `default`                               |
| `name`            | String | Name of the metrics provider    | true     | `metrics-server-veba`                   |
| `<provider_type>` | Object | Provider specific configuration | true     | See metrics provider type section below |

### Provider Type `default`

The VMware Event Router exposes metrics in JSON format on a configurable HTTP
listener, e.g. `http://<bindAddress>/stats`. The following table lists allowed
and optional fields for configuring the `default` metrics server.

| Field         | Type   | Description                                                                                 | Required | Example                    |
|---------------|--------|---------------------------------------------------------------------------------------------|----------|----------------------------|
| `bindAddress` | String | TCP/IP socket and port to listen on (**do not** add any URI scheme or slashes)              | true     | `"0.0.0.0:8082"`           |
| `<auth>`      | Object | **Optional:** authentication data (see auth section). Omit section if auth is not required. | false    | (see `basic_auth` example) |

# Deployment

VMware Event Router can be deployed and run as standalone binary (see
[below](#build-from-source)). However, it is designed (and recommended) to be
run in a Kubernetes cluster for increased availability and ease of scaling out.

> **Note:** Docker images are available
> [here](https://hub.docker.com/r/vmware/veba-event-router).

## Assisted Deployment

For your convenience we provide a Helm Chart which can be used to easily install
the VMware Event Router into an **existing** Knative or OpenFaaS ("faas-netes")
environment. 

⚠️ The OpenFaaS deployment method is unmaintained in the VEBA project and
will be deprecated in a future release. The recommended deployment method is
using the Knative backend.

### Helm Deployment

The Helm files are located in the [chart](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/vmware-event-router/chart) directory. The `values.yaml`
file contains the allowed parameters and parameter descriptions which map to the
VMware Event Router [configuration](#overview-configuration-file-structure-yaml)
file.

#### Option 1: Configuration with Knative

If you don't have a working Knative installation, follow the steps described in
the [official](https://knative.dev/docs/install/prerequisites/) documentation to
deploy Knative Serving **and** Eventing.

Now create a Helm `override.yaml` file with your environment specific settings,
e.g.:

```yaml
eventrouter:
  config:
    logLevel: debug
  vcenter:
    address: https://vcenter.corp.local
    username: administrator@vsphere.local
    password: replaceMe
    insecure: true # if required ignore TLS certs
  eventProcessor: knative
  knative:
    destination: # follows Knative convention for ref/uri
      ref:
        apiVersion: eventing.knative.dev/v1
        kind: Broker
        name: default
        namespace: default
```

> **Note:** Please ensure the correct formatting/indentation which follows the
> Helm `values.yaml` file.

#### Option 2: Configuration with OpenFaaS

The following steps can be used to quickly install OpenFaaS as a requirement for
the Helm installation instructions of the VMware Event Router below. Skip this
part if you already have an OpenFaaS environment set up.

```console
kubectl create ns openfaas && kubectl create ns openfaas-fn

helm repo add openfaas https://openfaas.github.io/faas-netes && \
    helm repo update \
    && helm upgrade openfaas --install openfaas/openfaas \
        --namespace openfaas \
        --set functionNamespace=openfaas-fn \
        --set generateBasicAuth=true

OF_PASS=$(echo $(kubectl -n openfaas get secret basic-auth -o jsonpath="{.data.basic-auth-password}" | base64 --decode))
```

Now create a Helm `override.yaml` file with your environment specific settings,
e.g.:

```yaml
eventrouter:
  config:
    logLevel: debug
  vcenter:
    address: https://vcenter.corp.local
    username: administrator@vsphere.local
    password: replaceMe
    insecure: true # if required ignore TLS certs
  eventProcessor: openfaas
  openfaas:
    address: http://gateway.openfaas:8080
    basicAuth: true
    username: admin
    password: ${OF_PASS} # variable from previous section

```

> **Note:** Please ensure the correct formatting/indentation which follows the
> Helm `values.yaml` file.

#### Deploy the VMware Event Router Helm Chart

Add the VMware Event Router Helm release to your Helm repository:

```console
# adds the veba chartrepo to the list of local registries with the repo name "veba"
$ helm repo add vmware-veba https://projects.registry.vmware.com/chartrepo/veba
```

To ensure new releases are pulled/updated and reflected locally update the repo index:

```console
$ helm repo update
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "vmware-veba" chart repository
Update Complete. ⎈ Happy Helming!⎈
```

The chart should now show up in the search:

```console
$ helm search repo event-router
NAME                      CHART VERSION   APP VERSION     DESCRIPTION
vmware-veba/event-router  v0.7.0          v0.7.0          The VMware Event Router is used to connect to v...
[snip]
```

> **Note:** To list/install development releases add the `--devel` flag to the Helm CLI.

Now install the chart using a Helm release name of your choice, e.g. `veba`,
using the configuration override file created above. The following command will
create a release from the chart in the namespace `vmware`, which will be
automatically created if it does not exist:

```console
$ helm install -n vmware --create-namespace veba vmware-veba/event-router -f override.yaml
NAME: veba
LAST DEPLOYED: Mon Jun  28 16:27:27 2020
NAMESPACE: vmware
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

Check the logs of the VMware Event Router to validate it is operating correctly:

```console
$ kubectl -n vmware logs deploy/router -f
```

If you run into issues, the logs should give you a hint, e.g.:

- configuration file not found -> file naming issue
- connection to vCenter/OpenFaaS cannot be established -> check values,
  credentials (if any) in the configuration file
- deployment/pod will not even come up -> check for resource issues, docker pull
  issues and other potential causes using the standard `kubectl` troubleshooting
  ways

To uninstall the release run:

```console
$ helm -n vmware uninstall veba
```

#### Creating/Updating the Chart

Before running the following commands make the appropriate changes to the chart,
e.g. bumping up `version` and/or `appVersion` in `Chart.yaml`.

```console
$ cd chart
$ helm package -d releases .
```

Then upload the created `.tgz` file inside `releases/` to your Helm chart repo.

### Manual Deployment

Create a namespace where the VMware Event Router will be deployed to:

```console
$ kubectl create namespace vmware
```

Use one of the configuration files provided [here](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/vmware-event-router/deploy) to configure the
router with **one** VMware vCenter Server `eventProvider` and **one** OpenFaaS
**or** AWS EventBridge `eventProcessor`. Change the values to match your
environment. The following example will use the OpenFaaS config sample.

> **Note:** Before continuing, make sure your environment is up and running,
> including Kubernetes and the configured event `processor`.

After you made your changes to the configuration file, save it as
`"event-router-config.yaml` in your current Git working directory.

> **Note:** If you have changed the port of the metrics server in the
> configuration file (default: 8080) make sure to also change that value in the
> YAML manifest (under the Kubernetes service entry).

Now, from your current Git working directory create a Kubernetes
[secret](https://kubernetes.io/docs/concepts/configuration/secret/) with the
configuration file as input:

```console
$ kubectl -n vmware create secret generic event-router-config --from-file=event-router-config.yaml
```

> **Note:** You might want to delete the (local) configuration file to not leave
> behind sensitive information on your local machine.

If you have configured `knative` as your event `processor` you must also create
a `ClusterRoleBinding` for the VMware Event Router so it can lookup Knative
`destinations`.

#### Create Knative ClusterRoleBinding (skip if not using Knative)

The following commands creates the `ClusterRoleBinding` assuming the VMware
Event Router will be deployed into the `vmware` namespace using the predefined
service account (deployment manifest). This requires a properly configured
Knative environment with `Serving` and `Eventing` CRDs installed.

```console
# only for Knative-based deployments
$ kubectl create clusterrolebinding veba-addressable-resolver --clusterrole=knative-serving-aggregated-addressable-resolver --serviceaccount=vmware:vmware-event-router
```

```console
$ kubectl describe clusterrole addressable-resolver
Name:         knative-serving-aggregated-addressable-resolver
Labels:       serving.knative.dev/release=v0.23.0
Annotations:  <none>
PolicyRule:
  Resources                                      Non-Resource URLs  Resource Names  Verbs
  ---------                                      -----------------  --------------  -----
  services                                       []                 []              [get list watch]
  brokers.eventing.knative.dev/status            []                 []              [get list watch]
  brokers.eventing.knative.dev                   []                 []              [get list watch]
  parallels.flows.knative.dev/status             []                 []              [get list watch]
  parallels.flows.knative.dev                    []                 []              [get list watch]
  sequences.flows.knative.dev/status             []                 []              [get list watch]
  sequences.flows.knative.dev                    []                 []              [get list watch]
  channels.messaging.knative.dev/status          []                 []              [get list watch]
  channels.messaging.knative.dev                 []                 []              [get list watch]
  inmemorychannels.messaging.knative.dev/status  []                 []              [get list watch]
  inmemorychannels.messaging.knative.dev         []                 []              [get list watch]
  parallels.messaging.knative.dev/status         []                 []              [get list watch]
  parallels.messaging.knative.dev                []                 []              [get list watch]
  sequences.messaging.knative.dev/status         []                 []              [get list watch]
  sequences.messaging.knative.dev                []                 []              [get list watch]
  routes.serving.knative.dev/status              []                 []              [get list watch]
  routes.serving.knative.dev                     []                 []              [get list watch]
  services.serving.knative.dev/status            []                 []              [get list watch]
  services.serving.knative.dev                   []                 []              [get list watch]
  channels.messaging.knative.dev/finalizers      []                 []              [update]
```

#### Create the VMware Event Router Deployment

Now we can deploy the VMware Event Router.

Download the latest deployment manifest (release.yaml) file from the Github
[release](https://github.com/vmware-samples/vcenter-event-broker-appliance/releases)
page. Then save the file under the same name (to follow along with the
commands).

Example Download with `curl`:

```console
curl -L -O https://github.com/vmware-samples/vcenter-event-broker-appliance/releases/latest/download/release.yaml
```

Deploy the VMware Event Router:

```console
$ kubectl -n vmware create -f release.yaml
```

Check the logs of the VMware Event Router to validate it started correctly:

```console
$ kubectl -n vmware logs deploy/vmware-event-router -f
```

If you run into issues, the logs should give you a hint, e.g.:

- configuration file not found -> file naming issue
- connection to vCenter/OpenFaaS cannot be established -> check values,
  credentials (if any) in the configuration file
- deployment/pod will not even come up -> check for resource issues, docker pull
  issues and other potential causes using the standard `kubectl` troubleshooting
  ways

To delete the deployment and secret simply delete the namespace we created
earlier:

```console
$ kubectl delete namespace vmware
```

## CLI Flags

By default the VMware Event Router binary will look for a YAML configuration
file named `/etc/vmware-event-router/config`. The default log level is `info`
and human readable colored console logs will be printed. This behavior can be
overridden via `log-json` to generate JSON logs and `log-level` to change the
log level. Stack traces are only generated in level `error` or higher. 

```console
$ ./vmware-event-router -h

 _    ____  ___                            ______                 __     ____              __
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/

Usage of dist/vmware-event-router:

  -config string
        path to configuration file (default "/etc/vmware-event-router/config")
  -log-json
        print JSON-formatted logs
  -log-level string
        set log level (debug,info,warn,error) (default "info")

```