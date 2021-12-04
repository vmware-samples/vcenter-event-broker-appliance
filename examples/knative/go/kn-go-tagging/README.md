# kn-go-tagging

Example Knative Go function for attaching or detaching a tag on a vCenter VM that sends a power on/off [CloudEvent](https://github.com/cloudevents/sdk-go) via the VMware Event Broker.

The [ko](https://github.com/google/ko) tool will be used to create the artifacts.

> **Note:** This guide assumes that you have stood up a working Knative environment using the vCenter Event Broker Appliance (VEBA).

# Step 1 - Build

> **Note:** This step is only required if you made code changes to any of the \*.go files.

`ko` was created to help Knative developers build images and binaries without the need for Dockerfiles nor Makefiles.

See more information in [this Knative blog](https://knative.dev/blog/articles/ko-fast-kubernetes-microservice-development-in-go/).

Follow setup instructions at https://github.com/google/ko.

You will also need to install [Go](https://golang.org/doc/install) for the ko tool to work.

After installing ko, set the destination for images with an environment variable.

```bash
KO_DOCKER_REPO=my-dockerhub-user
```

Run the following command from the `kn-go-tagging` directory. It will build and push the image to your local Docker daemon. To have `ko` publish and use a container image registry, remove the `--local` flag.

```bash
ko publish --local .
```

Save the image name so you can use it in the [function.yaml]'s Service container image section (mentioned in Step 3).

# Step 2 - Tests

Run unit tests using the following command:

`go test ./...`

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
> `default` Rabbit `broker`. The Knative `service` and `trigger` will be installed
> in the `vmware-functions` Kubernetes namespace, assuming that the `broker` is
> also available there.

Push your container image to an accessible registry (if it wasn't already done in Step 1 - Build above).

## Create vSphere Credentials Secret

Create a secret holding the username and password needed to access vCenter.

```bash
# create secret
kubectl create secret generic vsphere-credentials \
--from-literal=username=administrator@vsphere.local \
--from-literal=password='ReplaceMe' \
--namespace vmware-functions

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret tag-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from `Step 1`
if you made any changes. If not, the default VMware container image will
suffice.

Get the name of the tag that will be attached to the VM by creating a tag within vSphere.
Put the tag name in the `VALUE` field of `TAG_Name` for the Service manifest inside the `function.yaml` file.

You can specify if the tag should be attached or detached to the VM that triggers the event by setting `TAG_ACTION` to `"attach"` or `"detach"` in the `function.yaml` Service manifest.

These are all the configuration values that can be set in the Knative Service spec.template.spec.containers portion of the [Knative manifest](function.yaml):

| Configuration         | Description                                                                                                                | Example Values             | Required/Optional |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------- | -------------------------- | ----------------- |
| `image`               | Image containing the function. **Note:** You can keep the image as is in the `function.yaml` if you don't change the code. | ko.local/kn-go-tagging:0.1 | Required          |
| `VCENTER_INSECURE`    | Use TLS to connect to vCenter or not                                                                                       | `"true"` or `"false"`      | Optional          |
| `VCENTER_SECRET_PATH` | The path where the vSphere credentials secret will be mounted                                                              | `"/var/bindings/vsphere"`  | Optional          |
| `DEBUG`               | Set the logging verbosity for help with service development                                                                | `"true"` or `"false"`      | Optional          |
| `TAG_NAME`            | Name given to a tag created in vSphere                                                                                     | `"some-name"`              | Required          |
| `TAG_ACTION`          | Action for tag when specified VM event reaches service. Default is `"attach"`.                                             | `"attach"` or `"detach"`   | Optional          |

The Knative Trigger, also in the [Knative manifest](function.yaml), can also be updated.
The `spec.filter` section determines which events reach the Service.

The `spec.filter.attributes.subject` can be set to a VM event such as `VmPoweredOffEvent` or `VmPoweredOnEvent`.
Only exact matching on string values are supported by Knative trigger filtering.

The `spec.broker` in the Knative environment is assumed to be a Rabbit broker named `default`.

Deploy the function to the VMware Event Broker Appliance (VEBA):

```bash
# deploy function
kubectl apply -f function.yaml -n vmware-functions
```

For testing purposes, the [Knative manifest](function.yaml) contains the following annotations, which will ensure the Knative Service Pod will always run **exactly** one instance for debugging purposes.
Functions deployed through through the VMware Event Broker Appliance UI defaults to scale to 0, which means the pods will only run when it is triggered by an vCenter Event.

```yaml
annotations:
  autoscaling.knative.dev/maxScale: '1'
  autoscaling.knative.dev/minScale: '1'
```

# Step 4 - Undeploy

```bash
# undeploy function
kubectl delete -f function.yaml -n vmware-functions

# delete secret
kubectl delete secret vsphere-credentials -n vmware-functions
```
