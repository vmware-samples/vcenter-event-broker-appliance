# kn-pcli-hostmaint-alarms
Example Knative PowerCLI function kn-pcli-hostmaint-alarms. This function will disable host alarm actions when a host is placed in maintenance mode, and enable host alarm actions when a host is removed from maintenance mode.

# Step 1 - Build

> **Note:** This step is only required if you made code changes to `handler.ps1`
> or `Dockerfile`.

Create the container image locally to test your function logic.

Mac/Linux
```
# change the IMAGE name accordingly, example below for Docker
export TAG=<version>
export IMAGE=<docker-username>/kn-pcli-hostmaint-alarms:${TAG}
docker build -t ${IMAGE} .
```

Windows
```
# change the IMAGE name accordingly, example below for Docker
$TAG=<version>
$IMAGE="<docker-username>/kn-pcli-hostmaint-alarms:${TAG}"
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
> Any changes to variables must also be updated in `hostmaint_secret.yaml` and `test/docker-test-env-variable`

* `VCENTER_SERVER` - IP Address or FQDN of the vCenter Server to connect to
* `VCENTER_USERNAME` - vCenter account with permission to reconfigure distributed virtual switches
* `VCENTER_PASSWORD` - vCenter password associated with the username
* `VCENTER_CERTIFCATE_ACTION` - Set-PowerCLIConfiguration Action to configure when connection fails due to certificate error, default is Fail. (Possible values: Fail, Ignore or Warn)

If you built a custom image in Step 1, comment out the default `IMAGE` command below - the `docker run` command will then use use the value previously stored in the `IMAGE` variable. Otherwise, use the default image as shown below.  Start the container image by running the following commands:

Mac/Linux
```console
export TAG=<version>
export IMAGE=ghcr.io/vmware-samples/vcenter-event-broker-appliance/kn-pcli-hostmaint-alarms:${TAG}
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```
Windows
```console
$TAG=<version>
$IMAGE="ghcr.io/vmware-samples/vcenter-event-broker-appliance/kn-pcli-hostmaint-alarms:${TAG}"
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```

---
This function has two sample payload files: `enter-maint-payload.json` and `exit-maint-payload.json`. Locate the `Host` section of each file:
```json
    "Host": {
      "Name": "esx02.cl.vmweventbroker.io",
      "Host": {
        "Type": "HostSystem",
        "Value": "REPLACE-ME"
      }
    },
```
Change `Value:` from `REPLACE-ME` to the ID of a host currently in your vCenter inventory. If you do not make this change, the function will still be invoked, but the configuration operation will fail because the host will not be found. The function output will also make more sense if you change `Name:` to match the name of your host. The function will still work without changing the name.

One way to figure out a host's ID is with this PowerCLI command: 
```powershell
(Get-VMHost "esx02.cl.vmweventbroker.io").id
HostSystem-host-58
```

Based on the previous example host ID, the JSON should look like this:

```json
    "Host": {
      "Name": "esx02.cl.vmweventbroker.io",
      "Host": {
        "Type": "HostSystem",
        "Value": "host-58"
      }
    },
```

**WARNING** - This function will reconfigure alarm actions on the host you specify.
---

In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent payload being sent to the local container image. When run with no arguments, the scripts will send the contents of `enter-maint-payload.json` as the payload and an event subject of `EnteredMaintenanceModeEvent`. 

This is an example of running a test for entering maintenance mode.
```console
> .\send-cloudevent-test.ps1
Testing Function ...
See docker container console for output
```
```console
# Output from docker container console
04/21/2022 13:52:51 - DEBUG: Event - EnteredMaintenanceModeEvent
04/21/2022 13:52:52 - Disabling alarm actions on host: esx02.cl.vmweventbroker.io
04/21/2022 13:52:52 - kn-pcli-hostmaint-alarms operation complete ...
```

To simulate an exit maintenance mode event, pass the JSON file name along with the exit event name . Example: `./send-cloudevent-test.ps1 ./exit-maint-payload.json ExitMaintenanceModeEvent`.

```console
> ./send-cloudevent-test.ps1 ./exit-maint-payload.json ExitMaintenanceModeEvent
Testing Function ...
See docker container console for output
```

```console
# Output from docker container console
04/21/2022 14:01:31 - DEBUG: Event - ExitMaintenanceModeEvent
04/21/2022 14:01:31 - Enabling alarm actions on host: esx02.cl.vmweventbroker.io
04/21/2022 14:01:31 - kn-pcli-hostmaint-alarms operation complete ...
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

Update the `hostmaint_secret.json` file with your vCenter Server credentials and configurations and then create the kubernetes secret which can then be accessed from within the function by using the environment variable named called `HOSTMAINT_SECRET`.

```console
# create secret
kubectl -n vmware-functions create secret generic hostmaint-secret --from-file=HOSTMAINT_SECRET=hostmaint_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret hostmaint-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `EnteredMaintenanceModeEvent` and `ExitMaintenanceModeEvent` vCenter Server events. If you wish to change this, update the `subject` field within `function.yaml` to the desired event type.


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
kubectl -n vmware-functions delete secret hostmaint-secret
```
