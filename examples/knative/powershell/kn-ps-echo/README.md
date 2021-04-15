# kn-ps-echo
Simple Powershell function with an HTTP listener running in Knative to echo
[CloudEvents](https://github.com/cloudevents/).

> **Note:** CloudEvents using structured or binary mode are supported.

# Step 1 - Build

Create the container image and optionally push to an external registry such as Docker.

```
docker build -t <docker-username>/kn-ps-echo:1.0 .
docker push -t <docker-username>/kn-ps-echo:1.0
```

# Step 2 - Test

Verify the container image works by executing it locally.

```bash
docker run -e PORT=8080 -it --rm -p 8080:8080 <docker-username>/kn-ps-echo:1.0

# now in a separate window run the following

# Run either test.ps1 (PowerShell Script) or test.sh (Bash Script) to simulate a CloudEvent payload being sent to the container image
./test.ps1

Testing Function ...
See docker container console for output

# Output from docker container console
Server start listening on 'http://*:8080/'
Cloud Event
  Source: source-123
  Type: binary
  Subject: subject-123
  Id: id-123
CloudEvent Data:

Name                           Value
----                           -----
UserName                       VSPHERE.LOCAL\Administrator
Key                            607954
Datacenter                     {Datacenter, Name}
Dvs                            None
Net                            None
FullFormattedMessage           Test on  192.168.30.5 in Primp-Datacenter is powered off
ChangeTag
Ds                             None
ChainId                        607952
Template                       False
Vm                             {Vm, Name}
CreatedTime                    02/16/2021 20:27:31
ComputeResource                {ComputeResource, Name}
Host                           {Host, Name}
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit broker. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the broker is also available there.

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice and then deploy the function to VMware Event Broker Appliance (VEBA):

```bash
# Deploy function

kubectl apply -f function.yaml
```

For testing purposes, the `function.yaml` contains the following annotations, which will ensure the Knative Service Pod will always run **exactly** one instance for debugging purposes. Functions deployed through through the VMware Event Broker Appliance UI defaults to scale to 0, which means the pods will only run when it is triggered by an vCenter Event.

```yaml
annotations:
  autoscaling.knative.dev/maxScale: "1"
  autoscaling.knative.dev/minScale: "1"
```