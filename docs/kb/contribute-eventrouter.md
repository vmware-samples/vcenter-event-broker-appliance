---
layout: docs
toc_id: contribute-eventrouter
title: Building the Event Router
description: Building the Event Router
permalink: /kb/contribute-eventrouter
cta:
 title: Have a question? 
 description: Please check our [Frequently Asked Questions](/faq) first.
---

# Build VMware Event Router from Source

Requirements: This project uses [Golang](https://golang.org/dl/) and Go [modules](https://blog.golang.org/using-go-modules){:target="_blank"}. For convenience a Makefile and Dockerfile are provided requiring `make` and [Docker](https://www.docker.com/){:target="_blank"} to be installed as well.

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/vmware-event-router

# for Go versions before v1.13
export GO111MODULE=on 

# defaults to build with Docker (use make binary for local executable instead)
make 
```

<!-- omit in toc -->
# VMware Event Router

The VMware Event Router is used to connect to various VMware event `streams` (i.e. "sources") and forward these events to different `processors` (i.e. "sinks"). This project is currently used by the [*VMware Event Broker Appliance*](https://github.com/vmware-samples/vcenter-event-broker-appliance){:target="_blank"} as the core logic to forward vCenter events to configurable event `processors` (see below).

**Supported event sources:**
- [VMware vCenter Server](https://www.vmware.com/products/vcenter-server.html){:target="_blank"}

**Supported event processors:**
- [OpenFaaS](https://www.openfaas.com/){:target="_blank"}
- [AWS EventBridge](https://aws.amazon.com/eventbridge/?nc1=h_ls){:target="_blank"}

The VMware Event Router uses the [CloudEvents](https://cloudevents.io/){:target="_blank"} standard to format events from the supported `stream` providers in JSON. See [below](#example-event-structure) for an example.

**Current limitations:**

- Only one event `stream` and one event `processor` can be configured at a time
  - It is possible though to run **multiple instances** of the event router
- At-most-once delivery semantics are provided
  - See [this FAQ](https://github.com/vmware-samples/vcenter-event-broker-appliance/blob/development/FAQ.md) for a deeper understanding of messaging semantics

<!-- omit in toc -->
## Table of Contents
- [Usage and Configuration](#usage-and-configuration)
  - [Event Stream Provider and Processor Configuration Options](#event-stream-provider-and-processor-configuration-options)
  - [Stream Provider: Configuration Details for VMware vCenter Server](#stream-provider-configuration-details-for-vmware-vcenter-server)
  - [Stream Processor: Configuration Details for OpenFaaS](#stream-processor-configuration-details-for-openfaas)
  - [Stream Processor: Configuration Details for AWS EventBridge](#stream-processor-configuration-details-for-aws-eventbridge)
  - [Metrics Server: Configuration Details](#metrics-server-configuration-details)
  - [Deployment](#deployment)
- [Build from Source](#build-from-source)
- [Example Event Structure](#example-event-structure)

## Usage and Configuration

The VMware Event Router can be run standalone (statically linked binary) or deployed as a Docker container, e.g. in a Kubernetes environment. See [deployment](#deployment) for further instructions. The configuration of event `stream` providers and `processors` and other internal components (such as metrics) is done via a JSON file passed in via the `"-config"` command line flag.

```
 _    ____  ___                            ______                 __     ____              __
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/


Usage of ./vmware-event-router:

  -config string
        path to configuration file for metrics, stream source and processor (default "/etc/vmware-event-router/config")
  -verbose
        print event handling information

commit: <git_commit_sha>
version: <release_tag:master>
```

The following sections describe the layout of the configuration file (JSON) and specific options for the event `stream` provider, `processor` and `metrics` server. A correct configuration file requires `stream`, `processor` and `metrics` to be defined. Configuration examples are provided [here](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/vmware-event-router/deploy){:target="_blank"}.

> **Note:** Currently only one event `stream` (i.e. one vCenter Server) and one event `processor` can be configured at a time, e.g. one vCenter Server instance streaming events to OpenFaaS **or** AWS EventBridge. Specifying multiple instances of the same provider will lead to unintended behavior.

### Event Stream Provider and Processor Configuration Options

The following table lists allowed fields with their respective value types in the JSON configuration file. Detailed instructions for the specific event `stream` providers, `processors` and `metrics` are described in dedicated sections further below.

| Field    | Value             | Description                                    | Example                                                                                           |
|----------|-------------------|------------------------------------------------|---------------------------------------------------------------------------------------------------|
| type     | string            | event stream, processor or internal            | "type": "stream"                                                                                  |
| provider | string            | identifier of stream, processor or metrics     | "provider": "vmware_vcenter"                                                                      |
| address  | string            | URI of the provider (when required)            | "address": "https://10.0.0.1:443/sdk"                                                             |
| auth     | map[string]string | authentication options for the type provider   | "auth": { "method":"user_password","secret": {...}} **Note: see provider specific options below** |
| options  | map[string]string | provider specific options (see sections below) | "options":{"insecure": "true"}                                                                    |

> **Note:** Besides event `stream` providers and `processors` the configuration file is also used for router-internal components, such as metrics (and likely others in the future). The `type: internal` is reserved for these use cases.

### Stream Provider: Configuration Details for VMware vCenter Server

The following table lists allowed and optional fields for using VMware vCenter Server as an event `stream` provider.

| Field                | Value                         | Description                                                                                                                    |
|----------------------|-------------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| type                 | "stream"                      | VMware vCenter is an event **stream** provider.                                                                                |
| provider             | "vmware_vcenter"              | Use this exact value to use VMware vCenter Server as a provider.                                                               |
| address              | "https://10.0.0.1:443/sdk"    | URI of the VMware vCenter Server (IP or FQDN incl. "<:PORT>/sdk").                                                             |
| auth.method          | "user_password"               | Use this exact value. Only username/password are supported to authenticate against VMware vCenter Server.                      |
| auth.secret.username | "administrator@vsphere.local" | Replace with user/service account to use for connecting to this vCenter event stream.                                          |
| auth.secret.password | "REPLACE_ME"                  | Replace with password for the given user/service account to use for connecting to this vCenter event stream.                   |
| options.insecure     | "true"                        | Ignore TLS certificate warnings. **Note:** must use quotes around this value (is of type string). Default: "false". (optional) |

Example of the configuration section for VMware vCenter Server: 

```json
{
  "type": "stream",
  "provider": "vmware_vcenter",
  "address": "https://10.0.0.1:443/sdk",
  "auth": {
    "method": "user_password",
    "secret": {
      "username": "administrator@vsphere.local",
      "password": "REPLACE_ME"
    }
  },
  "options": {
    "insecure": "true"
  }
}
```

> **Note:** The JSON configuration file is an array of maps, ie. "[{<stream_provider>},{<stream_processor>}]". The snippet above is trimmed for readability. The examples provided [here](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/vmware-event-router/deploy){:target="_blank"} are properly formatted.

### Stream Processor: Configuration Details for OpenFaaS

OpenFaaS functions can subscribe to the event stream via function `"topic"` annotations in the function stack configuration (see OpenFaaS documentation for details on authoring functions), e.g.:

```yaml
annotations:
      topic: "VmPoweredOnEvent,VmPoweredOffEvent"
```

> **Note:** One or more event categories can be specified, delimited via `","`. A list of event names (categories) and how to retrieve them can be found [here](https://github.com/lamw/vcenter-event-mapping/blob/master/vsphere-6.7-update-3.md){:target="_blank"}. A simple "echo" function useful for testing is provided [here](https://github.com/embano1/of-echo/blob/master/echo.yml){:target="_blank"}.

The following table lists allowed and optional fields for using OpenFaaS as an event stream `processor`.

| Field                | Value                          | Description                                                                                                                                                 |
|----------------------|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------|
| type                 | "processor"                    | OpenFaaS is an event stream **processor**.                                                                                                                  |
| provider             | "openfaas"                     | Use this exact value to use OpenFaaS as a provider.                                                                                                         |
| address              | "http://gateway.openfaas:8080" | URI of the OpenFaaS gateway (IP or FQDN incl. "<:PORT>").                                                                                                   |
| auth.method          | "basic_auth"                   | Use this exact value. Only `"basic_auth"` is supported to authenticate against OpenFaaS (must use authentication).                                             |
| auth.secret.username | "admin"                        | Replace with OpenFaaS gateway admin user name (is "admin" unless changed during gateway deployment).                                                        |
| auth.secret.password | "REPLACE_ME"                   | Replace with password for the given admin account to use for connecting to the OpenFaaS gateway.                                                            |
| options.async        | "true"                         | Use `"async"` function invocation against the OpenFaaS gateway. **Note:** must use quotes around this value (is of type string). Default: "false". (optional) |

Example of the configuration section for OpenFaaS: 

```json
{
  "type": "processor",
  "provider": "openfaas",
  "address": "http://gateway.openfaas:8080",
  "auth": {
    "method": "basic_auth",
    "secret": {
      "username": "admin",
      "password": "REPLACE_ME"
    }
  },
  "options": {
    "async": "false"
  }
}
```

> **Note:** The JSON configuration file is an array of maps, ie. "[{<stream_provider>},{<stream_processor>}]". The snippet above is trimmed for readability. The examples provided [here](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/vmware-event-router/deploy){:target="_blank"} are properly formatted.

### Stream Processor: Configuration Details for AWS EventBridge

Amazon EventBridge is a serverless event bus that makes it easy to connect applications together using data from your own applications, integrated Software-as-a-Service (SaaS) applications, and AWS services. In order to reduce bandwidth and costs (number of events ingested, see [pricing](https://aws.amazon.com/eventbridge/pricing/){:target="_blank"}), VMware Event Router only forwards events configured in the associated `rule` of an event bus. Rules in AWS EventBridge use pattern matching ([docs](https://docs.aws.amazon.com/eventbridge/latest/userguide/filtering-examples-structure.html){:target="_blank"}). Upon start, VMware Event Router contacts EventBridge (using the given IAM role) to parse and extract event categories from the configured rule ARN (see configuration option below). 

The VMware Event Router uses the `"subject"` field in the event payload to store the event category, e.g. `"VmPoweredOnEvent"`. Thus it is required that you use a **specific pattern match** (`"detail->subject"`) that the VMware Event Router can parse to retrieve the desired event (forwarding) categories. For example, the following AWS EventBridge event pattern rule matches power on/off events (including DRS-enabled clusters):

```json
{
  "detail": {
    "subject": [
      "VmPoweredOnEvent",
      "VmPoweredOffEvent",
      "DrsVmPoweredOnEvent"
    ]
  }
}
```

`"subject"` can contain one or more event categories. Wildcards (`"*"`) are not supported. If one wants to modify the event pattern match rule **after** deploying the VMware Event Router, its internal rules cache is periodically synchronized with AWS EventBridge at a fixed interval of 5 minutes.

> **Note:** A list of event names (categories) and how to retrieve them can be found [here](https://github.com/lamw/vcenter-event-mapping/blob/master/vsphere-6.7-update-3.md){:target="_blank"}.

The following table lists allowed and optional fields for using AWS EventBridge as an event stream `processor`.

| Field                             | Value                                                             | Description                                                                                                                                |
|-----------------------------------|-------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| type                              | "processor"                                                       | AWS EventBridge is an event stream **processor**.                                                                                          |
| provider                          | "aws_event_bridge"                                                | Use this exact value to use AWS EventBridge as a provider.                                                                                 |
| auth.method                       | "access_key"                                                      | Use this exact value. Only `"access_key"` is supported to authenticate against AWS EventBridge.                                            |
| auth.secret.aws_access_key_id     | "ABCDEFGHIJK"                                                     | Access Key ID for the IAM role used.                                                                                                       |
| auth.secret.aws_secret_access_key | "ZYXWVUTSRQPO"                                                    | Secret Access Key for the IAM role used.                                                                                                   |
| options.aws_region                | "eu-central-1"                                                    | AWS region to use, see region [overview](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html){:target="_blank"} |
| options.aws_eventbridge_event_bus | "default"                                                         | Name of the event bus to use. Default: "default" (optional)                                                                                |
| options.aws_eventbridge_rule_arn  | "arn:aws:events:eu-central-1:1234567890:rule/vmware-event-router" | Rule ARN to use for event pattern matching.                                                                                                |

> **Note:** Currently only IAM user accounts with access key/secret are supported to authenticate against AWS EventBridge. Please follow the [user guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/getting-set-up-eventbridge.html){:target="_blank"} before deploying the event router. Further information can also be found in the [authentication](https://docs.aws.amazon.com/eventbridge/latest/userguide/auth-and-access-control-eventbridge.html#authentication-eventbridge){:target="_blank"} section. 

In addition to the recommendation in the AWS EventBridge user guide you might want to lock down the IAM role for the VMware Event Router and scope it to these permissions ("Action"):

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

Example of the configuration section for AWS EventBridge: 

```json
{
  "type": "processor",
  "provider": "aws_event_bridge",
  "auth": {
    "method": "access_key",
    "secret": {
      "aws_access_key_id": "ABCDEFGHIJK",
      "aws_secret_access_key": "ZYXWVUTSRQPO"
    }
  },
  "options": {
    "aws_region": "eu-central-1",
    "aws_eventbridge_event_bus": "default",
    "aws_eventbridge_rule_arn": "arn:aws:events:eu-central-1:1234567890:rule/vmware-event-router"
  }
}
```

> **Note:** The JSON configuration file is an array of maps, ie. "[{<stream_provider>},{<stream_processor>}]". The snippet above is trimmed for readability. The examples provided [here](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/vmware-event-router/deploy){:target="_blank"} are properly formatted.

### Metrics Server: Configuration Details

The VMware Event Router exposes metrics (JSON format) on the (currently hardcoded) HTTP endpoint `"http://IP:PORT/stats". The following table lists allowed and optional fields for configuring the metrics server.

| Field                | Value          | Description                                                                                                     |
|----------------------|----------------|-----------------------------------------------------------------------------------------------------------------|
| type                 | "internal"     | Metrics server is of type `"internal"`                                                                          |
| provider             | "metrics"      | Use this exact value to configure the metrics server.                                                           |
| address              | "0.0.0.0:8080" | Bind address for the http server to listen on.                                                                  |
| auth.method          | "basic_auth"   | `"basic_auth"` or `"none"` (disabled) is supported to configure authentication of the  metrics server endpoint. |
| auth.secret.username | "admin"        | Only required when `"basic_auth"` is configured.                                                                |
| auth.secret.password | "REPLACE_ME"   | Only required when `"basic_auth"` is configured.                                                                |

Example of the configuration section for the metrics server: 

```json
{
  "type": "metrics",
  "provider": "internal",
  "address": "0.0.0.0:8080",
  "auth": {
      "method": "none"
  }
}
```

### Deployment
VMware Event Router can be deployed and run as standalone binary (see [below](#build-from-source)). However, it is designed to be run in a Kubernetes cluster for increased availability and ease of scaling out. The following steps describe the deployment of the VMware Event Router in **a Kubernetes cluster** for an existing OpenFaaS ("faas-netes") environment, respectively AWS EventBridge.

> **Note:** Docker images are available [here](https://hub.docker.com/r/vmware/veba-event-router).

Create a namespace where the VMware Event Router will be deployed to:

```bash
kubectl create namespace vmware
```

Use one of the configuration files provided [here](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/vmware-event-router/deploy){:target="_blank"} to configure the router for **one** VMware vCenter Server event `stream` and **one** OpenFaaS **or** AWS EventBridge event stream `processor`. Change the values to match your environment. The following example will use the OpenFaaS config sample.

> **Note:** Make sure your environment is up and running, i.e. Kubernetes and OpenFaaS (incl. a function for testing) up and running or AWS EventBridge correctly configured (IAM Role, event bus and pattern rule).

After you made your changes to the configuration file, save it as `"event-router-config.json` in your current Git working directory. 

> **Note:** If you have changed the port of the metrics server in the configuration file (default: 8080) make sure to also change that value in the YAML manifest (under the Kubernetes service entry).

Now, from your current Git working directory create a Kubernetes [secret](https://kubernetes.io/docs/concepts/configuration/secret/){:target="_blank"} from the configuration file:

```bash
kubectl -n vmware create secret generic event-router-config --from-file=event-router-config.json
```

> **Note:** You might want to delete the (local) configuration file to not leave behind sensitive information on your local machine.

Now we can deploy the VMware Event Router:

```bash
kubectl -n vmware create -f deploy/event-router-k8s.yaml
```

Check the logs of the VMware Event Router to validate it started correctly:

```bash
kubectl -n vmware logs deploy/vmware-event-router -f
```

If you run into issues, the logs should give you a hint, e.g.:

- configuration file not found -> file naming issue
- connection to vCenter/OpenFaaS cannot be established -> check values in the configuration file
- deployment/pod will not even come up -> check for resource issues, docker pull issues and other potential causes using the standard kubectl troubleshooting ways

To delete the deployment and secret simply delete the namespace we created earlier:

```bash
kubectl delete namespace vmware
```
## Build from Source

Requirements: This project uses [Golang](https://golang.org/dl/) and Go [modules](https://blog.golang.org/using-go-modules). For convenience a Makefile and Dockerfile are provided requiring `make` and [Docker](https://www.docker.com/) to be installed as well.

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/vmware-event-router

# for Go versions before v1.13
export GO111MODULE=on 

# defaults to build with Docker (use make binary for local executable instead)
make 
```

## Example Event Structure

The following example for a `VmPoweredOnEvent` shows the event structure and payload:

```json
{
  "id": "08179137-b8e0-4973-b05f-8f212bf5003b",
  "source": "https://10.0.0.1:443/sdk",
  "specversion": "1.0",
  "type": "com.vmware.event.router/event",
  "subject": "VmPoweredOffEvent",
  "time": "2020-02-11T21:29:54.9052539Z",
  "data": {
    "Key": 9902,
    "ChainId": 9895,
    "CreatedTime": "2020-02-11T21:28:23.677595Z",
    "UserName": "VSPHERE.LOCAL\\Administrator",
    "Datacenter": {
      "Name": "testDC",
      "Datacenter": {
        "Type": "Datacenter",
        "Value": "datacenter-2"
      }
    },
    "ComputeResource": {
      "Name": "cls",
      "ComputeResource": {
        "Type": "ClusterComputeResource",
        "Value": "domain-c7"
      }
    },
    "Host": {
      "Name": "10.185.22.74",
      "Host": {
        "Type": "HostSystem",
        "Value": "host-21"
      }
    },
    "Vm": {
      "Name": "test-01",
      "Vm": {
        "Type": "VirtualMachine",
        "Value": "vm-56"
      }
    },
    "Ds": null,
    "Net": null,
    "Dvs": null,
    "FullFormattedMessage": "test-01 on  10.0.0.1 in testDC is powered off",
    "ChangeTag": "",
    "Template": false
  },
  "datacontenttype": "application/json"
}
```

> **Note:** If you use the AWS EventBridge stream `processor` the event is wrapped and accessible under `""detail": {}"` as a JSON-formatted string.
