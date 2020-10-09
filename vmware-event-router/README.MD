<!-- omit in toc -->
# VMware Event Router

The VMware Event Router is used to connect to various VMware event `providers`
(i.e. "sources") and forward these events to different event `processors` (i.e.
"sinks"). This project is currently used by the [*VMware Event Broker
Appliance*](https://www.vmweventbroker.io/) as the core logic to forward vCenter
events to configurable event `processors` (see below).

**Supported event sources:**
- [VMware vCenter Server](https://www.vmware.com/products/vcenter-server.html)

**Supported event processors:**
- [OpenFaaS](https://www.openfaas.com/)
- [AWS EventBridge](https://aws.amazon.com/eventbridge/?nc1=h_ls)

The VMware Event Router uses the [CloudEvents](https://cloudevents.io/) standard
to normalize events from the supported event `providers`. See
[below](#example-event-structure) for an example.

**Event Delivery Guarantees:**
- At-least-once event delivery for the [vCenter event provider](#provider-type-vcenter) (using checkpoints)
- At-most-once event delivery ("stateless" mode)

**Current limitations:**

- Only one event `provider` and one event `processor` can be configured at a
  time
- At-least-once event delivery semantics cannot be guaranteed if the event
  router crashes **within seconds** right after having received some events but before creating the
  first valid checkpoint
- At-least-once event delivery semantics are only supported for the vCenter
  event provider
- At-least-once event delivery semantics currently only guard against crashes of
  the event router. If an event cannot be delivered to an event processor,
  currently no retries are performed/dead letter queues are not supported

> **Note:** It is possible though to run **multiple instances** of the event
> router with different configurations to address multi-vCenter scenarios

<!-- omit in toc -->
## Table of Contents
- [Configuration](#configuration)
  - [Overview: Configuration File Structure
    (YAML)](#overview-configuration-file-structure-yaml)
  - [JSON Schema Validation](#json-schema-validation)
  - [API Version, Kind and Metadata](#api-version-kind-and-metadata)
  - [The `eventProvider` section](#the-eventprovider-section)
    - [Provider Type `vcenter`](#provider-type-vcenter)
  - [The `eventProcessor` section](#the-eventprocessor-section)
    - [Processor Type `openfaas`](#processor-type-openfaas)
    - [Processor Type `aws_event_bridge`](#processor-type-aws_event_bridge)
  - [The `auth` section](#the-auth-section)
    - [Type `basic_auth`](#type-basic_auth)
    - [Type `aws_access_key`](#type-aws_access_key)
  - [The `metricsProvider` section](#the-metricsprovider-section)
    - [Provider Type `default`](#provider-type-default)
- [Deployment](#deployment)
  - [Assisted Deployment](#assisted-deployment)
  - [Manual Deployment](#manual-deployment)
- [Build from Source](#build-from-source)
- [Contribute](#contribute)
- [Example Event Structure](#example-event-structure)

## Configuration

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
  -verbose
        verbose log output (default false)

commit: <git_commit_sha>
version: <release_tag>
```

The following sections describe the layout of the configuration file (YAML) and
specific options for the supported event `providers`, `processors` and `metrics`
endpoint. Configuration examples are provided [here](deploy/).

> **Note:** Currently only one event `provider` and one event `processor` can be
> configured at a time, e.g. one vCenter Server instance streaming events to
> OpenFaaS **or** AWS EventBridge. It is possible to run multiple instances of
> the event router with different configurations to address
> multi-provider/processor scenarios.

### Overview: Configuration File Structure (YAML)

The following file, using `vcenter` as the event `provider` and `openfaas` as
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
  type: openfaas
  name: veba-demo-openfaas
  openfaas:
    address: http://gateway.openfaas:8080
    async: false
    # assuming basic_auth enabled for OpenFaaS
    auth:
      type: basic_auth
      basicAuth:
        username: admin
        password: ReplaceMe
metricsProvider:
  type: default
  name: veba-demo-metrics
  default:
    bindAddress: "0.0.0.0:8082"
```

### JSON Schema Validation

In order to simplify the configuration and validation of the YAML configuration
file a JSON schema [file](README.MD) is provided. Many editors/IDEs offer
support for registering a schema file, e.g.
[Jetbrains](https://www.jetbrains.com/help/rider/Settings_Languages_JSON_Schema.html)
and [VS
Code](https://code.visualstudio.com/docs/languages/json#_json-schemas-and-settings).

> **Note:** The schema file can be downloaded and provided via a local file
> location or (recommended) via a direct URL, e.g. Github
> [raw](https://help.data.world/hc/en-us/articles/115006300048-GitHub-how-to-find-the-sharable-download-URL-for-files-on-GitHub)
> URL pointing to the aforementioned JSON schema file.


### API Version, Kind and Metadata

The following table lists allowed and required fields with their respective type
values and examples for these fields.

| Field             | Type              | Description                                     | Required | Example                            |
| ----------------- | ----------------- | ----------------------------------------------- | -------- | ---------------------------------- |
| `apiVersion`      | String            | API Version used for this configuration file    | true     | `event-router.vmware.com/v1alpha1` |
| `kind`            | String            | Type of this API resource                       | true     | `RouterConfig`                     |
| `metadata`        | Object            | Additional metadata for this configuration file | true     |                                    |
| `metadata.name`   | String            | Name of this configuration file                 | true     | `config-vc-openfaas-PROD`          |
| `metadata.labels` | map[String]String | Optional key/value pairs                        | false    | `env: PROD`                        |


### The `eventProvider` section

The following table lists allowed and required fields with their respective type
values and examples for these fields.

| Field             | Type   | Description                            | Required | Example                      |
| ----------------- | ------ | -------------------------------------- | -------- | ---------------------------- |
| `type`            | String | Type of the event provider             | true     | `vcenter`                    |
| `name`            | String | Name identifier for the event provider | true     | `vc-01-PROD`                 |
| `<provider_type>` | Object | Provider specific configuration        | true     | (see provider section below) |


#### Provider Type `vcenter`

The following table lists allowed and required fields for connecting to a
vCenter Server and the respective type values and examples for these fields.

| Field           | Type    | Description                                                                                           | Required | Example                          |
| --------------- | ------- | ----------------------------------------------------------------------------------------------------- | -------- | -------------------------------- |
| `address`       | String  | URI of the VMware vCenter Server                                                                      | true     | `https://10.0.0.1:443/sdk`       |
| `insecureSSL`   | Boolean | Skip TSL verification                                                                                 | true     | `true` (i.e. ignore errors)      |
| `checkpoint`    | Boolean | Configure checkpointing via checkpoint file for event recovery/replay purposes                        | true     | `true`                           |
| `checkpointDir` | Boolean | **Optional:** Configure an alternative location for persisting checkpoints (default: `./checkpoints`) | false    | `/var/local/checkpoints`         |
| `<auth>`        | Object  | vCenter credentials                                                                                   | true     | (see `basic_auth` example below) |

### The `eventProcessor` section

The following table lists allowed and required fields with their respective type
values and examples for these fields.

| Field              | Type   | Description                             | Required | Example                          |
| ------------------ | ------ | --------------------------------------- | -------- | -------------------------------- |
| `type`             | String | Type of the event processor             | true     | `openfaas` or `aws_event_bridge` |
| `name`             | String | Name identifier for the event processor | true     | `openfaas-01-PROD`               |
| `<processor_type>` | Object | Processor specific configuration        | true     | (see processor section below)    |

#### Processor Type `openfaas`

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
| --------- | ------- | ----------------------------------------------------------------------------------------------------------------- | -------- | ------------------------------------------------ |
| `address` | String  | URI of the OpenFaaS gateway                                                                                       | true     | `http://gateway.openfaas:8080`                   |
| `async`   | Boolean | Specify how to invoke functions (synchronously or asynchronously)                                                 | true     | `false` (i.e. use sync function invocation mode) |
| `<auth>`  | Object  | **Optional:** authentication data (see auth section below). Omit section if OpenFaaS gateway auth is not enabled. | false    | (see `basic_auth` example below)                 |

#### Processor Type `aws_event_bridge`

Amazon EventBridge is a serverless event bus that makes it easy to connect
applications together using data from your own applications, integrated
Software-as-a-Service (SaaS) applications, and AWS services. In order to reduce
bandwidth and costs (number of events ingested, see
[pricing](https://aws.amazon.com/eventbridge/pricing/)), VMware Event Router
only forwards events configured in the associated `rule` of an event bus. Rules
in AWS EventBridge use pattern matching
([docs](https://docs.aws.amazon.com/eventbridge/latest/userguide/filtering-examples-structure.html)).
Upon start, VMware Event Router contacts EventBridge (using the given IAM role)
to parse and extract event categories from the configured rule ARN (see
configuration option below).  

The VMware Event Router uses the `"subject"` field in the event payload to store
the event category, e.g. `"VmPoweredOnEvent"`. Thus it is required that you use
a **specific pattern match** (`"detail->subject"`) that the VMware Event Router
can parse to retrieve the desired event (forwarding) categories. For example,
the following AWS EventBridge event pattern rule matches power on/off events
(including DRS-enabled clusters):

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

`"subject"` can contain one or more event categories. Wildcards (`"*"`) are not
supported. If one wants to modify the event pattern match rule **after**
deploying the VMware Event Router, its internal rules cache is periodically
synchronized with AWS EventBridge at a fixed interval of 5 minutes.

> **Note:** A list of event names (categories) and how to retrieve them can be
> found
> [here](https://github.com/lamw/vcenter-event-mapping/blob/master/vsphere-6.7-update-3.md).

The following table lists allowed and optional fields for using AWS EventBridge
as an event `processor`.

| Field      | Type   | Description                                                                                                                             | Required | Example                                                                |
| ---------- | ------ | --------------------------------------------------------------------------------------------------------------------------------------- | -------- | ---------------------------------------------------------------------- |
| `region`   | String | AWS region to use, see [regions doc](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html). | true     | `us-west-1`                                                            |
| `eventBus` | String | Name of the event bus to use                                                                                                            | true     | `default` or `arn:aws:events:us-west-1:1234567890:event-bus/customBus` |
| `ruleARN`  | String | Rule ARN to use for event pattern matching                                                                                              | true     | `arn:aws:events:us-west-1:1234567890:rule/vmware-event-router`         |
| `<auth>`   | Object | AWS IAM role credentials                                                                                                                | true     | (see `aws_access_key` example below)                                   |

### The `auth` section

The following table lists allowed and required fields with their respective type
values and examples for these fields. Since the various `processors` and
`providers` use different authentication mechanisms (or none at all) this
section describes the various options.

#### Type `basic_auth`

Supported providers/processors:

- `vcenter` (required: `true`)
- `openfaas` (required: `false`, i.e. optional)
- `default` metrics server (see below) (required: `false`, i.e. optional)


| Field                | Type   | Description                             | Required | Example      |
| -------------------- | ------ | --------------------------------------- | -------- | ------------ |
| `type`               | String | Authentication method to use            | true     | `basic_auth` |
| `basicAuth`          | Object | Use when `basic_auth` type is specified | true     |              |
| `basicAuth.username` | String | Username                                | true     | `admin`      |
| `basicAuth.password` | String | Password                                | true     | `P@ssw0rd`   |

#### Type `aws_access_key`

Supported providers/processors:

- `aws_event_bridge`

| Field                        | Type   | Description                                 | Required | Example          |
| ---------------------------- | ------ | ------------------------------------------- | -------- | ---------------- |
| `type`                       | String | Authentication method to use                | true     | `aws_access_key` |
| `awsAccessKeyAuth`           | Object | Use when `aws_access_key` type is specified | true     |                  |
| `awsAccessKeyAuth.accessKey` | String | Access Key ID for the IAM role used         | true     | `ABCDEFGHIJK`    |
| `awsAccessKeyAuth.secretKey` | String | Secret Access Key for the IAM role used     | true     | `ZYXWVUTSRQPO`   |


> **Note:** Currently only IAM user accounts with access key/secret are
> supported to authenticate against AWS EventBridge. Please follow the [user
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

### The `metricsProvider` section

The VMware Event Router currently only exposes a default ("internal" or "embedded") metrics
endpoint. In the future, support for more providers is planned, e.g. Wavefront,
Prometheus, etc.

| Field             | Type   | Description                     | Required | Example                                 |
| ----------------- | ------ | ------------------------------- | -------- | --------------------------------------- |
| `type`            | String | Type of the metrics provider    | true     | `default`                               |
| `name`            | String | Name of the metrics provider    | true     | `metrics-server-veba`                   |
| `<provider_type>` | Object | Provider specific configuration | true     | See metrics provider type section below |

#### Provider Type `default`

The VMware Event Router exposes metrics in JSON format on a configurable HTTP
listener, e.g. `http://<bindAddress>/stats`. The following table lists allowed
and optional fields for configuring the `default` metrics server.

| Field         | Type   | Description                                                                                 | Required | Example                    |
| ------------- | ------ | ------------------------------------------------------------------------------------------- | -------- | -------------------------- |
| `bindAddress` | String | TCP/IP socket and port to listen on (**do not** add any URI scheme or slashes)              | true     | `"0.0.0.0:8082"`           |
| `<auth>`      | Object | **Optional:** authentication data (see auth section). Omit section if auth is not required. | false    | (see `basic_auth` example) |

## Deployment
VMware Event Router can be deployed and run as standalone binary (see
[below](#build-from-source)). However, it is designed (and recommended) to be
run in a Kubernetes cluster for increased availability and ease of scaling out.
The following steps describe the deployment of the VMware Event Router in **a
Kubernetes cluster** for an **existing** OpenFaaS ("faas-netes") environment,
respectively AWS EventBridge.

> **Note:** Docker images are available
> [here](https://hub.docker.com/r/vmware/veba-event-router).

### Assisted Deployment

For your convenience we provide an install script [here](hack/README.md) and a
Helm Chart is WIP
[#186](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues/186).

### Manual Deployment

Create a namespace where the VMware Event Router will be deployed to:

```bash
kubectl create namespace vmware
```

Use one of the configuration files provided [here](deploy/) to configure the
router with **one** VMware vCenter Server `eventProvider` and **one** OpenFaaS
**or** AWS EventBridge `eventProcessor`. Change the values to match your
environment. The following example will use the OpenFaaS config sample.

> **Note:** Before continuing, make sure your environment is up and running,
> including Kubernetes and OpenFaaS.

After you made your changes to the configuration file, save it as
`"event-router-config.yaml` in your current Git working directory. 

> **Note:** If you have changed the port of the metrics server in the
> configuration file (default: 8080) make sure to also change that value in the
> YAML manifest (under the Kubernetes service entry).

Now, from your current Git working directory create a Kubernetes
[secret](https://kubernetes.io/docs/concepts/configuration/secret/) with the
configuration file as input:

```bash
kubectl -n vmware create secret generic event-router-config --from-file=event-router-config.yaml
```

> **Note:** You might want to delete the (local) configuration file to not leave
> behind sensitive information on your local machine.

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
- connection to vCenter/OpenFaaS cannot be established -> check values,
  credentials (if any) in the configuration file
- deployment/pod will not even come up -> check for resource issues, docker pull
  issues and other potential causes using the standard kubectl troubleshooting
  ways

To delete the deployment and secret simply delete the namespace we created
earlier:

```bash
kubectl delete namespace vmware
```

## Build from Source

Requirements: This project uses [Golang](https://golang.org/dl/) and Go
[modules](https://blog.golang.org/using-go-modules). For convenience a Makefile
and Dockerfile are provided requiring `make` and
[Docker](https://www.docker.com/) to be installed as well.

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/vmware-event-router

# for Go versions before v1.13
export GO111MODULE=on 

# defaults to build with Docker (use "make binary" for local executable instead)
make 
```

> **Note:** For `_test.go` files your editor (e.g. vscode) might show errors and
> not be able to resolve symbols. This is due to the use of build tags which
> `gopls` currently does [not
> support](https://github.com/golang/go/issues/29202#issuecomment-515170916). In
> vscode add this to your configuration:
>
> ```json
> "go.toolsEnvVars": {
>        "GOFLAGS": "-tags=integration,unit"
> }
> ``` 

## Contribute

If you would like to make modification/additions to this code base, please
follow our [CONTRIBUTION](https://vmweventbroker.io/community) guidelines first.

In the `Makefile` we provide `make` targets for building a binary, Docker image
and validating changes via unit tests (`make unit-test`). These tests will run
when a pull request is submitted, but in order to run them locally to verify
your changes you need to have the following bits installed:

`make unit-test`:

- `go` tool chain
- `make`
- `gofmt`

To run the integration tests without the need to create the testbed manually use
the following script:

`./hack/run_integration_tests.sh`:

- `go` tool chain
- `jq`
- `kind`
- `docker`


## Example Event Structure

The following example for a `VmPoweredOnEvent` shows the event structure and
payload:

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

> **Note:** If you use the AWS EventBridge stream `processor` the event is
> wrapped and accessible under `""detail": {}"` as a JSON-formatted string.
