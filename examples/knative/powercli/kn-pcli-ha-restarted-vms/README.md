# kn-pcli-ha-restarted-vms
Example Knative PowerCLI function kn-pcli-ha-restarted-vms. This function emails an administrator a list of virtual machines that were restarted by vSphere High Availability (HA) after an HA event. The VM list contains all VMs that were restarted on the date the cluster failover event completes. Example: A failover event happens on Feb. 10th at 8:50AM. The function will email a timestamped list of all VMs that were failed over during the 24 hour period between midnight on Feb. 10th and midnight on Feb. 11th

# Step 1 - Build

> **Note:** This step is only required if you made code changes to `handler.ps1`
> or `Dockerfile`.

Create the container image locally to test your function logic.

Mac/Linux
```
# change the IMAGE name accordingly, example below for Docker
export TAG=<version>
export IMAGE=<docker-username>/kn-pcli-ha-restarted-vms:${TAG}
docker build -t ${IMAGE} .
```

Windows
```
# change the IMAGE name accordingly, example below for Docker
$TAG=<version>
$IMAGE="<docker-username>/kn-pcli-ha-restarted-vms:${TAG}"
docker build -t ${IMAGE} .
```
# Step 2 - Test

Verify the container image works by executing it locally.

Change into the `test` directory
```console
cd test
```

Update the following variable names within the `docker-test-env-variable` file.
> Any changes to variables must also be updated in `ha_secret.yaml` and `test/docker-test-env-variable`

* `VCENTER_SERVER` - IP Address or FQDN of the vCenter Server to connect to
* `VCENTER_USERNAME` - vCenter account with permission to reconfigure distributed virtual switches
* `VCENTER_PASSWORD` - vCenter password associated with the username
* `VCENTER_CERTIFCATE_ACTION` - Set-PowerCLIConfiguration Action to configure when connection fails due to certificate error, default is Fail. (Possible values: `Fail`, `Ignore` or `Warn`)
* `SMTP_SERVER` - The SMTP server responsible for relaying the e-mail notification
* `SMTP_PORT` - The port `SMTP_SERVER` is listening on
* `SMTP_USERNAME` - Optional. Username for SMTP authentication
* `SMTP_PASSWORD` - Optional. Username for SMTP authentication
* `EMAIL_TO` - Email address to receive notifications. At least one is required, multiple are allowed
* `EMAIL_FROM` - Email address to send notifications
* `DISPLAY_HOST_FQDN` - `true` or `false` - Set to true if you want the host's fully qualified domain name included in the report

If you built a custom image in Step 1, comment out the default `IMAGE` command below - the `docker run` command will then use use the value previously stored in the `IMAGE` variable. Otherwise, use the default image as shown below.  Start the container image by running the following commands:

Mac/Linux
```console
export TAG=<version>
export IMAGE=ghcr.io/vmware-samples/vcenter-event-broker-appliance/kn-pcli-ha-restarted-vms:${TAG}
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```
Windows
```console
$TAG=<version>
$IMAGE="ghcr.io/vmware-samples/vcenter-event-broker-appliance/kn-pcli-ha-restarted-vms:${TAG}"
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```

Unlike many other sample functions, you do not need to edit `test-payload.json` to test your function. However, your vCenter must have logged a recent HA event with VM restart for this function to work.

One way to simulate an HA event is by forcing a PSOD. You can do this at the command line of a vSphere host. Make sure to have at least one VM running on the host for HA to restart.
> Warning: This command will immediately cause a kernel panic, crashing your host! Do not run this in Production.
```console
vsish -e set /reliability/crashMe/Panic 1
```
Wait for HA to restart your VM(s) before continuing.

In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent payload being sent to the local container image. The scripts will send the contents of `test-payload.json` as the payload with a subject of `com.vmware.vc.HA.ClusterFailoverActionCompletedEvent`

```console
Testing Function ...
See docker container console for output

# Output from docker container console
04/25/2022 21:35:20 - PowerShell HTTP server start listening on 'http://*:8080/'
04/25/2022 21:35:20 - Processing Init

04/25/2022 21:35:20 - Configuring PowerCLI Configuration Settings

04/25/2022 21:35:21 - Connecting to vCenter Server vcsa.primp-industries.local

04/25/2022 21:35:25 - Successfully connected to vcsa.primp-industries.local

04/25/2022 21:35:25 - Init Processing Completed

04/25/2022 21:35:25 - Starting HTTP CloudEvent listener

04/25/2022 21:35:29 - DEBUG: From - noreply@vmweventbroker.io
04/25/2022 21:35:29 - DEBUG: To - notifications@vmweventbroker.io
04/25/2022 21:35:29 - Handler Processing Completed ...
```

> Pro Tip - If you are rapidly iterating on the code and want to easily rebuild and launch the container,
> you can chain all of the commands together with ampersands. This will allow you to re-run
> the commands by simply pressing the `up` arrow and `Enter`.

```console
cd .. && docker build -t ${IMAGE} . && cd test && docker run -e HA_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```
# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit `broker`. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the `broker` is also available there.

If you built a custom image, push it to an accessible registry such as Docker once you're done developing and testing your function logic.

```console
docker push ${IMAGE}
```

Update the `ha_secret.json` file with your vCenter Server credentials and configurations and then create the Kubernetes secret which can then be accessed from within the function by using the environment variable named called `HA_SECRET`.

```console
# create secret
kubectl -n vmware-functions create secret generic ha-secret --from-file=HA_SECRET=ha_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret ha-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `com.vmware.vc.HA.ClusterFailoverActionCompletedEvent` vCenter Server Event.

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
kubectl -n vmware-functions delete secret ha-secret
```
