# kn-pcli-tag
Example Knative PowerCLI function for applying a vSphere Tag when a Virtual Machine powers on.

# Step 1 - Build

Create the container image locally to test your function logic.

```
export TAG=<version>
docker build -t <docker-username>/kn-pcli-tag:${TAG} .
```

# Step 2 - Test

Verify the container image works by executing it locally.

Change into the `test` directory
```console
cd test
```

Update the following variable names within the `docker-test-env-variable` file

* VCENTER_SERVER - IP Address or FQDN of the vCenter Server to connect to for vSphere Tagging
* VCENTER_USERNAME - vCenter account with permission to apply vSphere Tagging
* VCENTER_PASSWORD - vCenter credentials to account with permission to apply vSphere Tagging
* VCENTER_TAG_NAME - Name of the vSphere Tag
* VCENTER_CERTIFCATE_ACTION - Set-PowerCLIConfiguration Action to configure when connection fails due to certificate error, default is Fail. (Possible values: Fail, Ignore or Warn)


Start the container image by running the following command:

```console
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 <docker-username>/kn-pcli-tag:${TAG}
```

In the `test` directory, edit `test-payload.json`. Locate the `Vm` section of the JSON file. Change the `Name:` property from `REPLACE-ME` to the name of a test VM currently in your vCenter inventory. If you do not make this change, the function will still be invoked, but the tag operation will fail because the VM will not be found.

```json
"Vm": {
  "Name": "REPLACE-ME",
  "Vm": {
	"Type": "VirtualMachine",
	"Value": "vm-11099"
  }
}
```
In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent payload being sent to the local container image

```console
Testing Function ...
See docker container console for output

# Output from docker container console
05/26/2021 13:46:44 - PowerShell HTTP server start listening on 'http://*:8080/'
05/26/2021 13:46:44 - Processing Init

05/26/2021 13:46:44 - Configuring PowerCLI Configuration Settings

05/26/2021 13:46:44 - Connecting to vCenter Server vcsa.primp-industries.local

05/26/2021 13:47:19 - Successfully connected to vcsa.primp-industries.local

05/26/2021 13:47:19 - Init Processing Completed

05/26/2021 13:47:32 - Processing Handler

05/26/2021 13:47:32 - Start CloudEvent Decode

05/26/2021 13:47:32 - CloudEvent Decode Complete

DEBUG: K8s Secrets:
{"VCENTER_SERVER":"vcsa.primp-industries.local","VCENTER_USERNAME":"administrator@vsphere.local","VCENTER_PASSWORD":"****","VCENTER_TAG_NAME":"Demo"}

DEBUG: CloudEventData

Name                           Value
----                           -----
Key                            2816789
Vm                             {Vm, Name}
Host                           {Host, Name}
Template                       False
CreatedTime                    05/17/2021 21:39:03
Net
Ds
Datacenter                     {Datacenter, Name}
ChainId                        2816787
UserName                       VSPHERE.LOCAL\Administrator
FullFormattedMessage           DRS powered on K8s-User-Group-Test on 192.168.30.5 in Primp-Datacenter
ChangeTag
ComputeResource                {ComputeResource, Name}
Dvs


05/26/2021 13:47:32 - Applying vSphere Tag "Demo" to K8s-User-Group-Test ...

05/26/2021 13:47:42 - vSphere Tag Operation complete ...

05/26/2021 13:47:42 - Handler Processing Completed
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit `broker`. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the `broker` is also available there.

Push your container image to an accessible registry such as Docker once you're done developing and testing your function logic.

```console
docker push <docker-username>/kn-pcli-tag:${TAG}
```

Update the `tag_secret.json` file with your vCenter Server credentials and configurations and then create the kubernetes secret which can then be accessed from within the function by using the environment variable named called `TAG_SECRET`.

```console
# create secret

kubectl -n vmware-functions create secret generic tag-secret --from-file=TAG_SECRET=tag_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret tag-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `DrsVmPoweredOnEvent` vCenter Server Event. If you wish to change this, update the `subject` field within `function.yaml` to the desired event type.


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

# delete secret
kubectl -n vmware-functions delete secret tag-secret
```
