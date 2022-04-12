# kn-go-nsx-tag-sync

Example Knative Go function for synchronizing vSphere virtual machine tags to
NSX-T based on vSphere tagging events.

⚠️ This guide assumes that you have stood up a working Knative environment using
the vCenter Event Broker Appliance (VEBA).

# How the Synchronization works

When a vSphere tag is `attached` or `detached` to/from a virtual machine, a
corresponding vSphere event is generated. This is also true when batched tagging
operations are performed, i.e. every tag generates one event.

The `kn-go-nsx-tag-sync` function reacts to these events and performs the
following steps:

1) Validate that the received event is a valid `CloudEvent`
1) Validate that the received `CloudEvent` payload (`data`) contains a valid
   tagging event
1) Retrieve the virtual machine `ManagedObjectReference` for the `Value` in the `Arguments[].Key == "Object"` field in the payload
1) If the `Object` does not resolve to a virtual machine, e.g. when a tag is attached to a folder or cluster, a warning is logged and the event is discarded
1) Retrieve the instance ID of the virtual machine
1) Retrieve all attached `category:tag` associations for the virtual machine
1) Update the tags in NSX-T for the virtual machine via
`api/v1/fabric/virtual-machines?action=update_tags`
[API](https://vdc-download.vmware.com/vmwb-repository/dcr-public/ce4128ae-8334-4f91-871b-ecce254cf69e/488f1280-204c-441d-8520-8279ac33d54b/api_includes/method_TagVirtualMachine.html)

In case of temporary errors, e.g. HTTP timeouts or any HTTP 5xx code from NSX-T,
the VEBA `default` broker retries to deliver the event to the function. The
function will also log any errors during execution.

## Performance and Scalability

The function is configured to process incoming tagging events serially in a FIFO
(first-in, first-out) order (one event in-flight, maximum one function instance
running). 

This guarantees that concurrent or batched tagging operations on the same object
are **not interleaved, producing determinstic synchronization** results (safety
guarantee), eventually converging to the desired state.

However, in larger environments with lots of objects and tags, the function can
fall behind processing from the VEBA `default` broker, causing delays and stale
tag states (views) in NSX-T. 

⚠️ FIFO execution can impact your network security, e.g. when a virtual machine is
supposed to be removed from a security group or firewall rule.

To increase throughput and reduce latency (staleness) several tuning knobs exist
to **parallize** the processing of tagging events (see the [advanced
settings](#advanced-settings) section). In this case, when a function instance
(`pod`) receives multiple events **for the same object** (virtual machine),
events are deduplicated and only one operation to synchronize the tag(s) is
performed.

⚠️ With parallel processing, **FIFO order is not guaranteed**. Depending on the
environment (size, tagging activities, etc.), there is a chance of interleaving
tagging operations **for the same object** which can lead to inconsistent
synchronisation results. The function will log cases were deduplication was
performed so an administrator can manually inspect the outcome.

⚠️ When function autoscaling is also enabled (`maxScale > 1`), concurrent
operations without FIFO order can only be detected by inspecting the logs of all
active instances of the `kn-go-nsx-tag-sync` function.

# Step 1 - Build

⚠️ This step is only required if you made code changes to any of the \*.go
files. To directly deploy the function jump to [Step 3](#step-3---deploy).

Requirement: If you make changes to the Go code, the
[ko](https://github.com/google/ko) tool is required to create the artifacts. 

Set the destination to push the function container image with an environment
variable.

```bash
export KO_DOCKER_REPO=docker.io/my-user
export KO_COMMIT=$(git rev-parse --short=8 HEAD)
export KO_TAG=1.0
```

The following command will build and push the image to the specified
`KO_DOCKER_REPO` repository.

```bash
# for docker.io
ko publish --bare -t $KO_TAG .

# for GCR
ko publish -B -t $KO_TAG .
```

⚠️ Using the above example, the resulting image would be
`docker.io/myuser/kn-go-nsx-tag-sync:1.0`.


# Step 2 - Test

Run unit tests using the following command:

```bash
go test -v -race -count 1 ./...
```

# Step 3 - Deploy

⚠️ The following steps assume a working Knative environment using the `default`
 Rabbit `broker`. The Knative `service` and `triggers` will be installed in the
 `vmware-functions` Kubernetes namespace, assuming that the `broker` is also
 available there.

## Create vSphere and NSX Credentials Secrets

Create a secret holding the username (role) and password needed to access
vCenter Server. The role must have at least **read-only** access to (the desired
subset of) virtual machines, e.g. cluster/datacenter, and tags in the inventory.

```bash
kubectl create secret generic vsphere-credentials \
--type=kubernetes.io/basic-auth \
--from-literal=username='ro-user@vsphere.local' \
--from-literal=password='ReplaceMe' \
--namespace vmware-functions

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret vsphere-credentials app=veba-ui
```

Create a secret holding the username (role) and password needed to access NSX
Manager. The role must have **write** permissions to manage virtual machine tags
(`Inventory > VM > Create & Assign Tags`).

```bash
kubectl create secret generic nsx-credentials \
--type=kubernetes.io/basic-auth \
--from-literal=username='tag-admin@nsx.local' \
--from-literal=password='ReplaceMe' \
--namespace vmware-functions

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret nsx-credentials app=veba-ui
```

## Update Environment Settings

Update environment specific settings under `env:` in the `function.yaml` file.

Please see the table below for a description of the available (and **required**)
settings.


| Configuration         | Description                                                                                       | Example Values                    | Required |
|-----------------------|---------------------------------------------------------------------------------------------------|-----------------------------------|----------|
| `VCENTER_URL`         | URL of the vCenter Server                                                                         | `"https://vcenter-01.corp.local"` | **Yes**  |
| `VCENTER_INSECURE`    | When set to `false` require strict TLS (certificate) validation when connecting to vCenter Server | `"true"`                          | No       |
| `VCENTER_SECRET_PATH` | The path where the vSphere credentials secret will be mounted                                     | `"/var/bindings/vsphere"`         | No       |
| `NSX_URL`             | URL of the NSX Manager Server                                                                     | `"https://nsx-01.corp.local"`     | **Yes**  |
| `NSX_INSECURE`        | When set to  `false` require strict TLS (certificate) validation when connecting to NSX Manager   | `"true"`                          | No       |
| `NSX_SECRET_PATH`     | The path where the NSX credentials secret will be mounted                                         | `"/var/bindings/nsx"`             | No       |
| `DEBUG`               | Enable debug logging                                                                              | `"true"`                          | No       |

## Deploy the Function

⚠️ If you made changes to the Go code/container image in [Step
1](#step-1---build) edit the `function.yaml` file with the custom name of the
container image used to build and push.

Deploy the function to the VMware Event Broker Appliance (VEBA):

```bash
kubectl apply -f function.yaml -n vmware-functions
```

For testing purposes, the [Knative manifest](function.yaml) contains the
following annotations, which will ensure the Knative Service Pod will always run
**exactly** one instance for debugging purposes. Functions deployed through
through the VMware Event Broker Appliance UI defaults to scale to 0, which means
the pods will only run when it is triggered by an vCenter Event.

```yaml
annotations:
  autoscaling.knative.dev/maxScale: "1"
  autoscaling.knative.dev/minScale: "1"
```

## Advanced Settings

The following sections describe advanced settings for the `function.yaml` file.

⚠️ Only change the default values once you understand the implications outlined
in the [performance and scalability](#performance-and-scalability) section.

### `containerConcurrency` (Function)

The field `containerConcurrency` influences how many events the function will
receive concurrently (i.e. "in-flight") from the attached trigger.

In the default `containerConcurrency: 1` setting, the function will process
exactly one event at the same time (no concurrency).

When changing this field to a higher value,
`rabbitmq.eventing.knative.dev/prefetchCount` must also be changed accordingly.

### `rabbitmq.eventing.knative.dev/prefetchCount` (Trigger)

The field `rabbitmq.eventing.knative.dev/prefetchCount` influences how many
events the corresponding trigger will pull (prefetch) from the event queue
(broker) and process them in parallel, i.e. send to the function.

In order for this setting to be effective, the function must be configured with
`containerConcurrency >= prefetchCount` (recommended) or
`autoscaling.knative.dev/maxScale >= prefetchCount`.

### `autoscaling.knative.dev/[min|max]Scale` (Function)

The fields `autoscaling.knative.dev/minScale` and
`autoscaling.knative.dev/maxScale` influence how many instances of the function
are allowed to run. 

In the default setting (below) the function is a singleton and will not scale to
zero:

```yaml
autoscaling.knative.dev/maxScale: "1"
autoscaling.knative.dev/minScale: "1"
```

To enable [scale to
zero](https://knative.dev/docs/serving/autoscaling/scale-to-zero/), set
`autoscaling.knative.dev/minScale: "0"`.

In busy vSphere environments with lots of tagging operations and many objects,
it might be required to allow multiple instances of the function to share the
event queue ("worker pattern"). This should be the last resort if changing
`containerConcurrency` is not sufficient.

Depending on the number of events per second and settings in
`containerConcurrency` and `rabbitmq.eventing.knative.dev/prefetchCount`, the
autoscaler will then create multiple instances of the function.

### Example

To handle up to **100 concurrent** tagging events if the default values (FIFO
semantics) are not appropriate:

*Trigger settings:*

```yaml
# pull up to 200 events (batch) from broker
rabbitmq.eventing.knative.dev/prefetchCount: "200"
```

*Function settings:*

```yaml
# each function instance will process up to 20 events concurrently
containerConcurrency: 20
```

```yaml
# scale to zero and max 10 function instances
autoscaling.knative.dev/maxScale: "10"
autoscaling.knative.dev/minScale: "0"
```

# Step 4 - Undeploy

```bash
# undeploy function
kubectl delete -f function.yaml -n vmware-functions

# delete secret
kubectl delete secret vsphere-credentials -n vmware-functions
kubectl delete secret nsx-credentials -n vmware-functions
```
