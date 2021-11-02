# kn-pcli-tag
Example Knative PowerCLI function for synchronizing vSphere VM tags to NSX-T
(via `api/v1/fabric/virtual-machines?action=update_tags` API).

# Step 1 - Build

> **Note:** This step is only required if you made code changes to `handler.ps1`
> or `Dockerfile`.

Create the container image locally to test your function logic.

```
# change the IMAGE name accordingly
export IMAGE=us.gcr.io/daisy-284300/veba/kn-pcli-nsx-tag-sync:1.0
docker build -t ${IMAGE} .
```

# Step 2 - Test

Verify the container image works by executing it locally.

Change into the `test` directory
```console
cd test
```

Update the following variable names within the `docker-test-env-variable` file

* `VCENTER_SERVER` - IP Address or FQDN of the vCenter Server to connect to for
  vSphere Tagging
* `VCENTER_USERNAME` - vCenter account with permission to apply vSphere Tagging
* `VCENTER_PASSWORD` - vCenter account password
* `VCENTER_CERTIFCATE_ACTION` - Set-PowerCLIConfiguration Action to configure when
  connection fails due to certificate error, default is Fail. (Possible values:
  `"Fail"`, `"Ignore"` or `"Warn"`)
* `NSX_SERVER` - IP Address or FQDN of the NSX-T manager Server to connect to for
  Tagging
* `NSX_USERNAME` - NSX-T manager account with permission to apply vSphere Tagging
* `NSX_PASSWORD` - NSX-T manager account password
* `NSX_SKIP_CERT_CHECK` - (Possible values: `"true"`, `"false"`)

Start the container image by running the following command:

```console
export IMAGE=us.gcr.io/daisy-284300/veba/kn-pcli-nsx-tag-sync:1.0
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```

In the `test` directory, edit `test-payload.json`. Locate the section of the
JSON file identifying a VM in your inventory. Change the `"Value": "example-vm"`
property to the name of a test VM currently in your vCenter inventory. If you do
not make this change, the function will still be invoked, but the tag
synchronization will fail because the VM will not be found.

```json
  {
    "Key": "Object",
    "Value": "example-vm"
  }
```
In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell
Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent
payload being sent to the local container image

```console
10/17/2021 18:36:40 - PowerShell HTTP server start listening on 'http://*:8080/'
10/17/2021 18:36:40 - Processing Init
10/17/2021 18:36:40 - Configuring PowerCLI Configuration Settings

[snip...]

10/17/2021 18:36:48 - Successfully connected to vcenter.local
10/17/2021 18:36:48 - Init Processing Completed
10/17/2021 18:36:48 - Starting HTTP CloudEvent listener

10/17/2021 18:47:35 - DEBUG: K8s Secrets:
{"VCENTER_SERVER":"vcenter.local","VCENTER_USERNAME":"administrator@vsphere.local","VCENTER_PASSWORD":"FILL_ME_IN","VCENTER_CERTIFICATE_ACTION":"Ignore","NSX_SERVER":"nsxt.local:443","NSX_USERNAME":"FILL-ME-IN","NSX_PASSWORD":"FILL-ME-IN","NSX_SKIP_CERT_CHECK":"FILL-ME-IN"}
10/17/2021 18:47:35 - DEBUG: CloudEventData:

Name                           Value
----                           -----
Net
ComputeResource
UserName                       VSPHERE.LOCAL\Administrator
Fault
Message
Host
ObjectName
ObjectId
ChangeTag
Dvs
Severity                       info
Ds
CreatedTime                    10/14/2021 09:47:57
Vm
ObjectType
EventTypeId                    com.vmware.cis.tagging.attach
ChainId                        326221
FullFormattedMessage           User VSPHERE.LOCAL\Administrator attached tag example-tag to object example-vm
Key                            326221
Arguments                      {Tag, Object, User}
Datacenter


10/17/2021 18:47:36 - DEBUG: CloudEventDataArguments:

Name                           Value
----                           -----
Value                          example-tag
Key                            Tag
Value                          example-vm
Key                            Object
Value                          VSPHERE.LOCAL\Administrator
Key                            User


10/17/2021 18:47:36 - DEBUG: VM name: example-vm
10/17/2021 18:47:36 - DEBUG: VM Persistence ID: 503a37c2-2b61-5747-94bb-43f622b8380f
10/17/2021 18:47:42 - DEBUG: Tag: resources/preemptible
10/17/2021 18:47:42 - DEBUG: nsxURL="https://nsxt.local:443/api/v1/fabric/virtual-machines?action=update_tags
"
10/17/2021 18:47:42 - DEBUG: headers="
Name  : Accept=
Value : application/json

Name  : Content-Type
Value : application/json

Name  : Authorization
Value : Basic RklMTC1NRS1JTjpGSUxMLU1FLUlO


"
10/17/2021 18:47:42 - DEBUG: nsxbody="{
  "external_id": "503a37c2-2b61-5747-94bb-43f622b8380f",
  "tags": [
    {
      "scope": "resources",
      "tag": "example-tag"
    }
  ]
}
"
10/17/2021 18:47:42 - DEBUG: Applying vSphere Tags for  example-vm to NSX-T
10/17/2021 18:47:42 - vSphere Tag to NSX Operation complete
10/17/2021 18:47:42 - Handler Processing complete
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
> `default` Rabbit `broker`. The Knative `service` and `trigger` will be installed
> in the `vmware-functions` Kubernetes namespace, assuming that the `broker` is
> also available there.

Push your container image to an accessible registry such as Docker once you're
done developing and testing your function logic.

> **Note:** This step is only required if you made code changes to `handler.ps1`
> or `Dockerfile`.

```console
export IMAGE=us.gcr.io/daisy-284300/veba/kn-pcli-nsx-tag-sync:1.0
docker push ${IMAGE}
```

Update the `tag_secret.json` file with your vCenter Server credentials and
configurations and then create the kubernetes secret which can then be accessed
from within the function by using the environment variable named called
`TAG_SECRET`.

```console
# create secret

kubectl -n vmware-functions create secret generic tag-secret --from-file=TAG_SECRET=tag_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret tag-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from `Step 1`
if you made any changes. If not, the default VMware container image will
suffice. By default, the function deployment will filter (`trigger`) on the
`com.vmware.cis.tagging.attach` and `com.vmware.cis.tagging.detach` vCenter
Server events. If you wish to change this, update the `subject` field within
`function.yaml` to the desired event type.

Deploy the function to the VMware Event Broker Appliance (VEBA).

```console
# deploy function

kubectl -n vmware-functions apply -f function.yaml
```

For testing purposes, the `function.yaml` contains the following annotations,
which will ensure the Knative Service Pod will always run **exactly** one
instance for debugging purposes. Functions deployed through through the VMware
Event Broker Appliance UI defaults to scale to 0, which means the pods will only
run when it is triggered by an vCenter Event.

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
