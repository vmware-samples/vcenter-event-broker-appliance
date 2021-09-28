# kn-pcli-snapshot-cron
Example Knative PowerCLI function using a [Knative PingSource](https://knative.dev/docs/developer/eventing/sources/ping-source/) to run a scheduled job (cron) for managing VM snapshot retention policies.

**Note:** There is currently a hard coded limit (`$maxVM`) of 20 VMs within the [handler.ps1](handler.ps1) to limit the number of snapshot operations that are performed on your vCenter Server.

# Step 1 - Build

Create the container image locally to test your function logic.

```
export TAG=<version>
docker build -t <docker-username>/kn-pcli-snapshot-cron:${TAG} .
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
* VCENTER_CERTIFCATE_ACTION - Set-PowerCLIConfiguration Action to configure when connection fails due to certificate error, default is Fail. (Possible values: Fail, Ignore or Warn)

Start the container image by running the following command:

```console
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 <docker-username>/kn-pcli-snapshot-cron:${TAG}
```

In the `test` directory, edit `test-payload.json` to match your desired snapshot retention configuration along with VMs that exists in your development environment for testing purposes.

* `dryRun` - Determines whether to remove a snapshot when one of the retention policies has been violated. During development and/or testing purposes, you can leave the default value of `true` and only change it to `false` when you are confident in your testing
* `retentionConfig` - VM snapshot retention policy configuration which can be evaluated on a per-VM basis.
  * `sizeGB` - Maximum snapshot size (GB) for a given snapshot
  * `days` - Maximum age (Days) for a given snapshot
* `virtualMachines` - Name of VM(s) to apply the snapshot retention policy

**Note:** At least one snapshot retention policy must be defined. If values are defined for all snapshot policies, then the order of evaluation will be based on `sizeGB` and then `days` and which ever condition is true.

```json
{
    "dryRun": true,
    "retentionConfig": {
        "sizeGB": "",
        "days": "5"
    },
	"virtualMachines": [
		"VM-1",
		"VM-2"
	]
}
```

In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent payload being sent to the local container image

```console
Testing Function ...
See docker container console for output

# Output from docker container console
09/29/2021 14:11:48 - PowerShell HTTP server start listening on 'http://*:8080/'
09/29/2021 14:11:48 - Processing Init

09/29/2021 14:11:48 - Configuring PowerCLI Configuration Settings


DefaultVIServerMode         : Multiple
ProxyPolicy                 : UseSystemProxy
ParticipateInCEIP           : True
CEIPDataTransferProxyPolicy : UseSystemProxy
DisplayDeprecationWarnings  : True
InvalidCertificateAction    : Ignore
WebOperationTimeoutSeconds  : 300
VMConsoleWindowBrowser      :
Scope                       : Session

DefaultVIServerMode         :
ProxyPolicy                 :
ParticipateInCEIP           : True
CEIPDataTransferProxyPolicy :
DisplayDeprecationWarnings  :
InvalidCertificateAction    : Ignore
WebOperationTimeoutSeconds  :
VMConsoleWindowBrowser      :
Scope                       : User

DefaultVIServerMode         :
ProxyPolicy                 :
ParticipateInCEIP           :
CEIPDataTransferProxyPolicy :
DisplayDeprecationWarnings  :
InvalidCertificateAction    :
WebOperationTimeoutSeconds  :
VMConsoleWindowBrowser      :
Scope                       : AllUsers

09/29/2021 14:11:48 - Connecting to vCenter Server vcsa.primp-industries.local

IsConnected   : True
Id            : /VIServer=vsphere.local\administrator@vcsa.primp-industries.local:443/
ServiceUri    : https://vcsa.primp-industries.local/sdk
SessionSecret : "e614889b9a2d292216febfa7020d0f0321d91708"
Name          : vcsa.primp-industries.local
Port          : 443
SessionId     : "e614889b9a2d292216febfa7020d0f0321d91708"
User          : VSPHERE.LOCAL\Administrator
Uid           : /VIServer=vsphere.local\administrator@vcsa.primp-industries.local:443/
Version       : 7.0.2
Build         : 18455184
ProductLine   : vpx
InstanceUuid  : 056a402b-2b3d-4f74-93f7-fa818adff697
RefCount      : 1
ExtensionData : VMware.Vim.ServiceInstance

09/29/2021 14:12:11 - Successfully connected to vcsa.primp-industries.local

09/29/2021 14:12:11 - Init Processing Completed

09/29/2021 14:12:11 - Starting HTTP CloudEvent listener
09/29/2021 14:12:14 - DEBUG: K8s Secrets:
{"VCENTER_SERVER": "vcsa.primp-industries.local","VCENTER_USERNAME" : "XXX","VCENTER_PASSWORD" : "XXX","VCENTER_CERTIFICATE_ACTION" : "Ignore"}

09/29/2021 14:12:14 - DEBUG: CloudEvent

DataContentType : application/json
Data            : {123, 32, 32, 32…}
Id              : id-123
DataSchema      :
Source          : source-123
SpecVersion     : V1_0
Subject         :
Time            :
Type            : dev.knative.sources.ping

09/29/2021 14:12:14 - DEBUG: CloudEventData
 {
  "virtualMachines": [
    "VM-1",
    "VM-2"
  ],
  "retentionConfig": {
    "sizeGB": "",
    "days": "5"
  },
  "dryRun": false
}

09/29/2021 14:12:14 - Checking VM: VM-2
09/29/2021 14:12:14 - Checking VM: VM-1
09/29/2021 14:12:14 - 	Snapshot Test-Snap-1 is 10 days old and exceeds maximum number of days (5)
09/29/2021 14:12:14 - 	Snapshot removal started for Test-Snap-1
09/29/2021 14:12:14 - Handler Processing Completed ...
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit `broker`. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the `broker` is also available there.

Push your container image to an accessible registry such as Docker once you're done developing and testing your function logic.

```console
docker push <docker-username>/kn-pcli-snapshot-cron${TAG}
```

Update the `snapshot_secret.json` file with your vCenter Server credentials and configurations and then create the kubernetes secret which can then be accessed from within the function by using the environment variable named called `SNAPSHOT_SECRET`.

```console
# create secret

kubectl -n vmware-functions create secret generic snapshot-secret --from-file=SNAPSHOT_SECRET=snapshot_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret snapshot-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. Next update the `spec.data` (JSON object) within the `PingSource` section for the following configurations.

1. Update the `schedule` parameter with the desired cron schedule. If you are unsure about the correct syntax for cron, you can use the free online tool [crontab.guru](https://crontab.guru/) for additional assistance. When specifying a schedule, ensure there is sufficient buffer time for when a VM is in violation of your snapshot policy and needs to be removed, which can take some time. A daily schedule during off-peak hours is recommended to ensure additional I/O will not impact your running workloads
1. Update the `retentionConfig` settings to reflect the desired snapshot policies and for policies not in use, simply leave value as an empty string `""`
1. Update `virtualMachines` which should be an array of the VMs you wish to apply your snapshot retention policies against

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
kubectl -n vmware-functions delete secret snapshot-secret
```
