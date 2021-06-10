# kn-go-echo

Simple Go function running in Knative to echo [CloudEvents](https://github.com/cloudevents/sdk-go).
In addition, [ko](https://github.com/google/ko) is used to create the artifacts.

Valid events will be printed out JSON-encoded, separated by their CloudEvent
`attributes` and `data` for better readability. This example service also
returns the received event, enabling it to be used as part of a pipeline of
events.

## Step 1: Build with `ko`

`ko` was created to help Knative developers build images and binaries without
the need for Dockerfiles nor Makefiles. See more information in [this Knative
blog](https://knative.dev/blog/2018/12/18/ko-fast-kubernetes-microservice-development-in-go/).

Follow setup instructions at https://github.com/google/ko.

After installing ko, set the destination for images with an environment variable.

``` bash
KO_DOCKER_REPO=my-dockerhub-user
```

Run the following command from the `kn-go-echo` directory. It will build and
push the image to your local Docker daemon. To have `ko` publish and use a
container image registry, remove the `--local` flag.

```bash
ko publish --local .
```

## Step 2: Test

To run the container using a locally stored image, use

```bash
docker run -p 8080:8080 $(ko publish --local .)
```

In a separate window, send a POST request. You can send fake cloud event for
testing purposes.

```bash
$ curl -X POST -i -d@testdata/cloudevent.json \
  --header 'Content-Type: application/cloudevents+json' \
  localhost:8080
HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Ce-Id: 08179137-b8e0-4973-b05f-8f212bf5003b
Ce-Source: https://10.0.0.1:443/sdk
Ce-Specversion: 1.0
Ce-Subject: VmPoweredOffEvent
Ce-Time: 2020-02-11T21:29:54.9052539Z
Ce-Type: com.vmware.event.router/event
Content-Length: 1042
Content-Type: application/json
Date: Fri, 07 May 2021 21:29:27 GMT

 {        "Key": 9902,        "ChainId": 9895,        "CreatedTime": "2020-02-11T21:28:23.677595Z",        "UserName": "VSPHERE.LOCAL\\Administrator",        "Datacenter": {            "Name": "testDC",            "Datacenter": {                "Type": "Datacenter",                "Value": "datacenter-2"            }        },        "ComputeResource": {            "Name": "cls",            "ComputeResource": {                "Type": "ClusterComputeResource",                "Value": "domain-c7"            }        },        "Host": {            "Name": "10.185.22.74",            "Host": {                "Type": "HostSystem",                "Value": "host-21"            }        },        "Vm": {            "Name": "test-01",            "Vm": {                "Type": "VirtualMachine",                "Value": "vm-56"            }        },        "Ds": null,        "Net": null,        "Dvs": null,        "FullFormattedMessage": "test-01 on  10.0.0.1 in testDC is powered off",        "ChangeTag": "",        "Template": false    }
```

The following lines should appear in the docker container:

```
2021/05/06 21:09:03 listening on :8080
***cloud event***
Context Attributes,
  specversion: 1.0
  type: com.vmware.event.router/event
  source: https://10.0.0.1:443/sdk
  subject: VmPoweredOffEvent
  id: 08179137-b8e0-4973-b05f-8f212bf5003b
  time: 2020-02-11T21:29:54.9052539Z
  datacontenttype: application/json
Data,
  {
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
  }
```

## Step 3: Deploy

**Note:** The following steps assume a working Knative environment using the
`default` Rabbit broker. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the broker is also available there.

Edit the function.yaml file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice and then deploy the function to VMware Event Broker Appliance (VEBA):

```bash
# Deploy function
kubectl -n vmware-functions apply -f function.yaml
```

For testing purposes, the function.yaml contains the following annotations, which will ensure the Knative Service Pod will always run exactly one instance for debugging purposes. Functions deployed through through the VMware Event Broker Appliance UI defaults to scale to 0, which means the pods will only run when it is triggered by an vCenter Event.

```yaml
annotations:
  autoscaling.knative.dev/maxScale: "1"
  autoscaling.knative.dev/minScale: "1"
```

## Step 4: Undeploy

```bash
# Undeploy function
kubectl -n vmware-functions delete -f function.yaml
```
