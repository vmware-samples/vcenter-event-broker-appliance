# Knative Function Template 
Delete this section before publishing your new function
- Find all instances of #REPLACE-ME# and replace it with descriptive names for your function
	- For example, if you are writing a function for a distributed virtual switch, you might name the function `kn-pcli-dvs`
- Find all instances of FUNCTION_SECRET and function_secret, replacing it with descriptive names for your function. For example, if you are writing a function for a distributed virtual switch, you might use DVS_SECRET
	- test/docker-test-env-variable
	- handler.ps1
	- README.md
	- function_secret.json (rename the file)
- Obtain a new payload file
	- The example `test/test-payload.json` is a sample event payload file for event type `DvsReconfiguredEvent`. This needs to be replaced with the payload for your specific event. This is easily obtained using the built-in Sockeye service. Browse to the `/events` endpoint of your VEBA deployment and cause your event to trigger. You can then can easily copy and paste the payload from the output in Sockeye.
# kn-pcli-#REPLACE-ME#
Example Knative PowerCLI function #REPLACE-ME#

# Step 1 - Build

> **Note:** This step is only required if you made code changes to `handler.ps1`
> or `Dockerfile`.

Create the container image locally to test your function logic.

Mac/Linux
```
# change the IMAGE name accordingly, example below for Docker
export IMAGE=<docker-username>/kn-pcli-#REPLACE-ME#:1.0
docker build -t ${IMAGE} .
```

Windows
```
# change the IMAGE name accordingly, example below for Docker
$IMAGE="<docker-username>/kn-pcli-#REPLACE-ME#:1.0"
docker build -t ${IMAGE} .
```
# Step 2 - Test

Verify the container image works by executing it locally.

Change into the `test` directory
```console
cd test
```

Update the following variable names within the `docker-test-env-variable` file.
> Note - The sample variables are built for a vCenter function. You will need to replace them if you are authoring a function for a different purpose.
> Any changes to variables must also be updated in `function_secret.yaml` and `test/docker-test-env-variable`

* `VCENTER_SERVER` - IP Address or FQDN of the vCenter Server to connect to
* `VCENTER_USERNAME` - vCenter account with permission to reconfigure distributed virtual switches
* `VCENTER_PASSWORD` - vCenter password associated with the username
* `VCENTER_CERTIFCATE_ACTION` - Set-PowerCLIConfiguration Action to configure when connection fails due to certificate error, default is Fail. (Possible values: Fail, Ignore or Warn)

If you built a custom image in Step 1, comment out the default `IMAGE` command below - the `docker run` command will then use use the value previously stored in the `IMAGE` variable. Otherwise, use the default image as shown below.  Start the container image by running the following commands:

Mac/Linux
```console
export IMAGE=us.gcr.io/daisy-284300/veba/kn-#REPLACE-ME#:1.0
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```
Windows
```console
$IMAGE="us.gcr.io/daisy-284300/veba/kn-pcli-#REPLACE-ME#:1.0"
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```

# This section will vary based on the payload for the function you write. Delete this heading before you publish your function.
---
In the `test` directory, edit `test-payload.json`. Locate the `Dvs` section of the JSON file. Change the `Name:` property from `REPLACE-ME` to the name of a distributed virtual switch currently in your vCenter inventory. If you do not make this change, the function will still be invoked, but the configuration operation will fail because the distributed virtual switch will not be found.

```json
	"Dvs": {
	  "Name": "REPLACE-ME",
	  "Dvs": {
		"Type": "DistributedVirtualSwitch",
		"Value": "dvs-7005"
	  }
	},
```

Next, locate the `ConfigSpec` section of the JSON file. Change the `MaxMTU` property from `REPLACE-ME` to a numeric value. In order for the function to make a change to the distributed virtual switch, the value in the JSON file must be set to a different value than the `ENFORCED_MTU` value in the `docker-test-env-variable` file. The `ConfigSpec` section of the file below has been truncated for brevity.

```json
"ConfigSpec": {
	  "ConfigVersion": "4",
	  "VspanConfigSpec": null,
	  "MaxMtu": REPLACE-ME,
	  "LinkDiscoveryProtocolConfig": null,
	},
```
**WARNING** - This function will reconfigure your distributed virtual switch - if your MTU is not set to 1500, it will be reset by this function.
---

In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent payload being sent to the local container image. When run with no arguments, the scripts will send the contents of `test-payload.json` as the payload. If you pass the scripts a different filename as an argument, they will send the contents of the specified file instead. Example: `send-cloudevent-test.ps1 test-payload2.json`. This testing technique is useful when writing complex functions with varying payloads.

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
Dvs                            {Dvs, Name}
UserName                       VSPHERE.LOCAL\Administrator
Host
ConfigSpec                     {MaxPorts, ExtensionKey, Host, Name…}
Key                            567176
ComputeResource
Ds
Vm
ConfigChanges                  {Modified, Added, Deleted}
ChangeTag
Net
CreatedTime                    12/19/2021 21:01:35
FullFormattedMessage           The vSphere Distributed Switch dvsTest in HomeLab was reconfigured. …
Datacenter                     {Datacenter, Name}
ChainId                        567175

Found VDS from event: dvsTest
Current MTU from event: 1500
MTU currently set to 1502 - resetting to 1500
Resetting MTU to 1500

12/19/2021 21:01:39 - VDS reconfig operation complete ...

12/19/2021 21:01:39 - Handler Processing Completed ...
```

---

> Pro Tip - If you are rapidly iterating on the code and want to easily rebuild and launch the container,
> you can chain all of the commands together with ampersands. This will allow you to re-run
> the commands by simply pressing the `up` arrow and `Enter`.

```console
cd .. && docker build -t ${IMAGE} . && cd test && docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```
# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit `broker`. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the `broker` is also available there.

If you built a custom image, push it to an accessible registry such as Docker once you're done developing and testing your function logic.

```console
docker push ${IMAGE}
```

Update the `function_secret.json` file with your vCenter Server credentials and configurations and then create the kubernetes secret which can then be accessed from within the function by using the environment variable named called `FUNCTION_SECRET`.

```console
# create secret

kubectl -n vmware-functions create secret generic function-secret --from-file=FUNCTION_SECRET=function_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret function-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `DvsReconfiguredEvent` vCenter Server Event. If you wish to change this, update the `subject` field within `function.yaml` to the desired event type.


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
kubectl -n vmware-functions delete secret function-secret
```
