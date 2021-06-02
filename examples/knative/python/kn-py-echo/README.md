# kn-py-echo
Example Python function with `Flask` REST API running in Knative to echo
[CloudEvents](https://github.com/cloudevents/sdk-python).

# Step 1 - Build with `pack`

[Buildpacks](https://buildpacks.io) are used to create the container image.

```bash
IMAGE=<docker-username>/kn-py-echo:1.0
pack build -B gcr.io/buildpacks/builder:v1 ${IMAGE}
```

# Step 2 - Test 

Verify the container image works by executing it locally.

```bash
docker run -e PORT=8080 -it --rm -p 8080:8080 <docker-username>/kn-py-echo:1.0
```
You should see output similar to the following:
```
* Serving Flask app "handler.py" (lazy loading)
 * Environment: development
 * Debug mode: on
 * Running on all addresses.
   WARNING: This is a development server. Do not use it in a production deployment.
 * Running on http://172.17.0.2:8080/ (Press CTRL+C to quit)
 * Restarting with stat
 * Debugger is active!
 * Debugger PIN: 994-125-687
 ```


In a separate terminal window, go to the test directory and use the `testevent.json` file to validate the function is working. 

```console
cd test
curl -i -d@testevent.json localhost:8080
```
You should see output similar to this below.
```
HTTP/1.1 100 Continue

HTTP/1.0 204 NO CONTENT
Content-Type: application/json
Server: Werkzeug/2.0.1 Python/3.8.6
Date: Wed, 26 May 2021 18:56:27 GMT
```
Return to the previous terminal window where you started the docker image, and you should see output similar to the following:
```
* Serving Flask app "handler.py" (lazy loading)
 * Environment: development
 * Debug mode: on
 * Running on all addresses.
   WARNING: This is a development server. Do not use it in a production deployment.
 * Running on http://172.17.0.2:8080/ (Press CTRL+C to quit)
 * Restarting with stat
 * Debugger is active!
 * Debugger PIN: 994-125-687
2021-05-26 18:56:27,719 INFO handler Thread-3 : "***cloud event*** {"attributes": {"specversion": "1.0", "id": "08179137-b8e0-4973-b05f-8f212bf5003b", "source": "https://10.0.0.1:443/sdk", "type": "com.vmware.event.router/event", "datacontenttype": "application/json", "subject": "VmPoweredOffEvent", "time": "2020-02-11T21:29:54.9052539Z"}, "data": {"Key": 9902, "ChainId": 9895, "CreatedTime": "2020-02-11T21:28:23.677595Z", "UserName": "VSPHERE.LOCAL\\Administrator", "Datacenter": {"Name": "testDC", "Datacenter": {"Type": "Datacenter", "Value": "datacenter-2"}}, "ComputeResource": {"Name": "cls", "ComputeResource": {"Type": "ClusterComputeResource", "Value": "domain-c7"}}, "Host": {"Name": "10.185.22.74", "Host": {"Type": "HostSystem", "Value": "host-21"}}, "Vm": {"Name": "test-01", "Vm": {"Type": "VirtualMachine", "Value": "vm-56"}}, "Ds": null, "Net": null, "Dvs": null, "FullFormattedMessage": "test-01 on  10.0.0.1 in testDC is powered off", "ChangeTag": "", "Template": false}}
172.17.0.1 - - [26/May/2021 18:56:27] "POST / HTTP/1.1" 204 -
2021-05-26 18:56:27,720 INFO werkzeug Thread-3 : 172.17.0.1 - - [26/May/2021 18:56:27] "POST / HTTP/1.1" 204 -
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit `broker`. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the `broker` is also available there.

Push your container image to an accessible registry such as Docker once you're done developing and testing your function logic.

```console
docker push <docker-username>/kn-py-echo:1.0
```
Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `VmPoweredOffEvent` vCenter Server Event. If you wish to change this, update the `subject` field within `function.yaml` to the desired event type.

Deploy the function to the VMware Event Broker Appliance (VEBA).

```console
# deploy function
kubectl -n vmware-functions apply -f function.yaml
```

For testing purposes, the `function.yaml` contains the following annotations, which will ensure the Knative Service Pod will always run **exactly** one instance for debugging purposes. Functions deployed through through the VMware Event Broker Appliance UI defaults to scale to 0, which means the pods will only run when it is triggered by an vCenter Event.

```yaml
annotations:
  autoscaling.knative.dev/maxScale: "1"
  autoscaling.knative.dev/minScale: "1"
```
# Step 4 - Undeploy

```console
# undeploy function
kubectl -n vmware-functions delete -f function.yaml
```