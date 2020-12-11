# VMware Event Broker Appliance Echo Event Function for Knative

## Description

This function helps users understand the structure and data of a given vCenter
Event which will be useful when creating brand new Functions. This function is
intended to be used with the `Knative` event `processor` in the VMware Event
Broker Appliance (VEBA).

> **Note:** The example assumes you have a working VEBA and Knative environment with a
> broker named `"rabbit"` installed in the `"default"` Kubernetes namespace of
> your Knative cluster. If you haven't done so, please follow the steps as described [here](https://github.com/knative-sandbox/eventing-rabbitmq/tree/master/broker).

### Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/python/kn-echo
git checkout master
```

### Step 2 - Install the Knative client (`kn`)

Use one of the following installation methods described
[here](https://github.com/knative/client/blob/master/docs/README.md#installing-kn)
to download the `kn` client. 

> **Note:** Make the binary executable (e.g. `chmod +x <file>`) and put it in a folder which is in your `$PATH` .

Verify `kn` is successfully installed:

```bash
kn version

# example output
Version:
Build Date:
Git Revision:
Supported APIs:
* Serving
  - serving.knative.dev/v1 (knative-serving v0.19.0)
* Eventing
  - sources.knative.dev/v1alpha2 (knative-eventing v0.19.0)
  - eventing.knative.dev/v1beta1 (knative-eventing v0.19.0)
```

### Step 3 - Deploy the `kn-echo` Knative Service

Next deploy the `kn-echo` function as a Knative
[Service](https://knative.dev/docs/serving/). This function will be used as the
target ("sink") for the Knative broker to send vCenter events to. We will
disable the scale-to-zero functionality so it is easier to tail the logs during
this exercise.

> **Note:** The source code of the function can be
found [here](https://github.com/embano1/kn-echo).

```bash
# deploy the container/service disabling scale-to-zero
kn service create kn-echo --port 8080 --image embano1/kn-echo:latest --scale 1
Creating service 'kn-echo' in namespace 'default':

  0.029s The Configuration is still working to reflect the latest desired specification.
  0.119s The Route is still working to reflect the latest desired specification.
  0.231s Configuration "kn-echo" is waiting for a Revision to become ready.
 28.223s ...
 28.312s Ingress has not yet been reconciled.
 28.375s Waiting for load balancer to be ready
 28.724s Ready to serve.
```

Verify the service is running (`READY == True`):

```bash
kn service list
NAME      URL                                       LATEST            AGE   CONDITIONS   READY   REASON
kn-echo   http://kn-echo.default.127.0.0.1.nip.io   kn-echo-sjcyf-1   41m   3 OK / 3     True
```

### Step 4 - Create the Knative Trigger

Even though the broker might already receive vCenter events and the `kn-echo` function
(Knative Service) is deployed, you won't yet see any events in the function
logs.

In order for the broker to know which events to send to the desired target(s)
Knative has the concept of a
[Trigger](https://knative.dev/docs/eventing/triggers/). We will create a basic
trigger to send all incoming events from the broker to our `kn-echo` function.

```bash
kn trigger create veba-kn-echo --broker rabbit --sink ksvc:kn-echo
Trigger 'veba-kn-echo' successfully created in namespace 'default'.
```

Verify the trigger is correctly initialized (`READY == True`):

```bash
kn trigger list
NAME            BROKER    SINK                 AGE   CONDITIONS   READY   REASON
veba-kn-echo    rabbit    ksvc:kn-echo         2s    5 OK / 5     True
```

### Step 5 - Tail the log of the `kn-echo` Function

Depending on your vCenter environment you might already see events coming in to
the `kn-echo` function. Start tailing the log:

```bash
DEPLOYMENT=$(echo $(kn service describe kn-echo -o=jsonpath="{.spec.template.metadata.name}")-deployment)
kubectl logs deploy/${DEPLOYMENT} -c user-container -f
 * Serving Flask app "handler.py" (lazy loading)
 * Environment: development
 * Debug mode: on
 * Running on http://0.0.0.0:8080/ (Press CTRL+C to quit)
 * Restarting with stat
 * Debugger is active!
 * Debugger PIN: 964-990-737
```

Now in vCenter trigger an event, e.g. `VmPoweredOffEvent` and watch the log stream:

```bash
[...]
2020-12-11 09:57:24,953 INFO werkzeug Thread-326 : 127.0.0.1 - - [11/Dec/2020 09:57:24] "POST / HTTP/1.1" 204 -
2020-12-11 09:57:24,966 INFO handler Thread-327 : {'attributes': 
{'specversion': '1.0', 'id': '10854380-5d90-4dbc-9a22-d3ef92f74b39', 
'source': 'https://10.184.99.138/sdk', 
'type': 'com.vmware.event.router/event', 
'datacontenttype': 'application/json', 
'subject': 'VmPoweredOffEvent', 
'time': '2020-12-11T09:57:24.9159985Z'}, 
'data': {'Key': 19416, 'ChainId': 19409, 
'CreatedTime': '2020-12-11T09:57:22.208999Z', 
'UserName': 'VSPHERE.LOCAL\\Administrator', 
'Datacenter': {'Name': 'vcqaDC', 'Datacenter': {'Type': 'Datacenter', 'Value': 'datacenter-2'}}, 
'ComputeResource': {'Name': 'cls', 'ComputeResource': {'Type': 'ClusterComputeResource', 'Value': 'domain-c7'}}, 
'Host': {'Name': '10.184.111.20', 'Host': {'Type': 'HostSystem', 'Value': 'host-33'}}, 
'Vm': {'Name': 'test-01', 'Vm': {'Type': 'VirtualMachine', 'Value': 'vm-48'}}, 
'Ds': None, 
'Net': None, 
'Dvs': None, 
'FullFormattedMessage': 'test-01 on  10.184.111.20 in vcqaDC is powered off', 'ChangeTag': '', 'Template': False}}
[...]
```

> **Note:** Output above reformatted for readability.