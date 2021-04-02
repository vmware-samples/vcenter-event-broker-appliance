# kn-echo
Simple Go function running in Knative to echo [CloudEvents](https://github.com/cloudevents/sdk-go).
In addition, [Buildpacks](https://buildpacks.io) are used to create the artifacts.

Valid events will be printed out JSON-encoded, separated by their CloudEvent
`attributes` and `data` for better readability.
## Deployment (Knative)

**Note:** The following steps assume a working Knative environment using the
`default` broker. The Knative `service` and `trigger` will be installed in the
`default` Kubernetes namespace, assuming that the broker is also available there.

```bash
# create the service
kn service create kn-go-echo --port 8080 --image vmware/veba-kn-go-echo:latest

# create the trigger
kn trigger create kn-go-echo --broker default --sink ksvc:kn-go-echo
```

## Build with `pack`

- Requirements:
  - `pack` (see: https://buildpacks.io/docs/app-developer-guide/)
  - Docker

```bash
IMAGE=vmware/veba-kn-go-echo:latest
pack build -B gcr.io/buildpacks/builder:v1 ${IMAGE}
```

## Verify the image works by executing it locally

```bash
docker run -e PORT=8080 -it --rm -p 8080:8080 vmware/veba-kn-go-echo:latest

# now in a separate window or use -d in the docker cmd above to detach
$ curl -X POST -i localhost:8080
HTTP/1.1 204 No Content
Date: Fri, 02 Apr 2021 21:15:20 GMT

# or using a fake event
$ curl -X POST -i -d@testdata/cloudevent.json localhost:8080
HTTP/1.1 100 Continue

HTTP/1.1 204 No Content
Date: Fri, 02 Apr 2021 21:23:40 GMT

# you should see the following lines printed in the docker container
$ docker run --rm -p 8080:8080 veba-kn-go-echo
2021/04/02 21:24:54 Starting server at port 8080.
2021/04/02 21:24:56 ***cloud event*** {"attributes":{"datacontenttype":"application/json","id":"08179137-b8e0-4973-b05f-8f212bf5003b","source":"https://10.0.0.1:443/sdk","specversion":"1.0","subject":"VmPoweredOffEvent","time":"2020-02-11T21:29:54.9052539Z","type":"com.vmware.event.router/event"},"data":{"ChainId":9895,"ChangeTag":"","ComputeResource":{"ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c7"},"Name":"cls"},"CreatedTime":"2020-02-11T21:28:23.677595Z","Datacenter":{"Datacenter":{"Type":"Datacenter","Value":"datacenter-2"},"Name":"testDC"},"Ds":null,"Dvs":null,"FullFormattedMessage":"test-01 on  10.0.0.1 in testDC is powered off","Host":{"Host":{"Type":"HostSystem","Value":"host-21"},"Name":"10.185.22.74"},"Key":9902,"Net":null,"Template":false,"UserName":"VSPHERE.LOCAL\\Administrator","Vm":{"Name":"test-01","Vm":{"Type":"VirtualMachine","Value":"vm-56"}}}}

```
