# kn-pcli-datastore-usage-email
Example Knative PowerCLI function kn-pcli-datastore-usage-email. Sends email notifications to a specified email address for datastore usage on disk alarms. Optional configuration for a per-datastore email, enabling different recipients for different datastores.

# Step 1 - Build

> **Note:** This step is only required if you made code changes to `handler.ps1`
> or `Dockerfile`.

Create the container image locally to test your function logic.

Mac/Linux
```
# change the IMAGE name accordingly, example below for Docker
export IMAGE=<docker-username>/kn-pcli-datastore-usage-email:1.0
docker build -t ${IMAGE} .
```

Windows
```
# change the IMAGE name accordingly, example below for Docker
$IMAGE="<docker-username>/kn-pcli-datastore-usage-email:1.0"
docker build -t ${IMAGE} .
```
# Step 2 - Test

Verify the container image works by executing it locally.

Change into the `test` directory
```console
cd test
```

Update the following variable names within the `docker-test-env-variable` file.

* `VCENTER_SERVER` - IP Address or FQDN of the vCenter Server to connect to
* `VCENTER_USERNAME` - vCenter account with permission to reconfigure distributed virtual switches
* `VCENTER_PASSWORD` - vCenter password associated with the username
* `VCENTER_CERTIFCATE_ACTION` - Set-PowerCLIConfiguration Action to configure when connection fails due to certificate error, default is `Fail`. (Possible values: `Fail`, `Ignore` or `Warn`)
* `VC_ALARM_NAME` - The alarm to trigger alerts for. The default is the default datastore usage alarm for all vCenter installations. If you have a custom
* `DATASTORE_NAMES` - A list of datastore names that you want monitored by the function
* `SMTP_SERVER` - SMTP server IP or FQDN
* `SMTP_PORT` - SMTP port, typically 25 for unauthenticated and 587 for authenticated
* `SMTP_USERNAME` - Optional. Username for authenticated SMTP
* `SMTP_PASSWORD` - Optional. Password for authenticated SMTP
* `EMAIL_SUBJECT` - The subject line of the notification email
* `EMAIL_TO` - A list of recipients for the notification
* `EMAIL_FROM` - The email address the notification email comes from
* `DATASTORE_CUSTOM_PROP_EMAIL_TO` - Optional. The name of a custom attribute containing datastore-specific notification email addresses. See the `Custom Recipients` section for details.
## Custom recipients

The function always sends notification emails to `EMAIL_TO`. Some customers want a specific group notified based on the datastore. For example, you might want a group of database administrators notified if a datastore dedicated to MySQL databses begins to fill.

You can accomplish per-datastore notifications by configuring any datastore with a custom attribute. For example, you might add a custom attribute named `notify_email` to `datastore1`. Another datastore, `datastore2`, does not have the custom attribute. You then update `DATASTORE_CUSTOM_PROP_EMAIL_TO` with a value of `notify_email`. When the function runs, it will check the alarming datastore for the presence of a custom attribute named `notify_email`. If the custom attribute is found, notifications for `datastore1` will be sent to `EMAIL_TO` and the email address found in custom attribute `notify_email`. Notifications for `datastore2` will be sent only to `EMAIL_TO`.

## Run Container

If you built a custom image in Step 1, comment out the default `IMAGE` command below - the `docker run` command will then use use the value previously stored in the `IMAGE` variable. Otherwise, use the default image as shown below.  Start the container image by running the following commands:

Mac/Linux
```console
export IMAGE=us.gcr.io/daisy-284300/veba/kn-pcli-datastore-usage-email:1.0
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```
Windows
```console
$IMAGE="us.gcr.io/daisy-284300/veba/kn-pcli-datastore-usage-email:1.0"
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```
If you are not using the custom attribute functionality, you can test the function without making any changes. You can skip the next section.

## Payload changes for custom attribute functionality

In the `test` directory, edit `test-payload.json`. Locate the `Ds` section of the JSON file. Change the `Name:` property from `ProdDatastore1` to the name of the datastore in your vCenter inventory that contains the custom attribute. If you do not make this change, the function will still be invoked, but no notifications will be sent because the datastore will not be found.

```json
	"Ds": {
	  "Name": "ProdDatastore1",
	  "Datastore": {
		"Type": "Datastore",
		"Value": "datastore-60"
	  }
	},
```

You must also edit `docker-test-env-variable`, replace one of the existing values with the same datastore name you used in `test-payload.json`
```json
"DATASTORE_NAMES":["ProdDatastore1","ProdDatastore2"]
```
>Note : If you make a change to `docker-test-env-variable`, you must run the `docker build` command again.
## Using the test scripts

In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent payload being sent to the local container image. When run with no arguments, the scripts will send the contents of `test-payload.json` as the payload. If you pass the scripts a different filename as an argument, they will send the contents of the specified file instead. Example: `send-cloudevent-test.ps1 test-payload2.json`. This technique can be useful if you want to test notifications for multiple datastores.

```console
Testing Function ...
See docker container console for output

# Output from docker container console
04/28/2022 22:01:44 - DEBUG: Alarm Name: Datastore usage on disk
04/28/2022 22:01:44 - DEBUG: DS Name: VEBA-DS-01
04/28/2022 22:01:44 - DEBUG: Alarm Status: yellow
04/28/2022 22:01:44 - DEBUG: vCenter: source-123
04/28/2022 22:01:44 - DEBUG: Data Center:  DataCenter1
04/28/2022 22:01:44 - DEBUG: Alarm to Monitor: Datastore usage on disk
04/28/2022 22:01:44 - DEBUG: Datastores to Monitor: VEBA-DS-01 VEBA-DS-02
04/28/2022 22:01:44 - DEBUG: Message Subject: ⚠️ [VMC Datastore Notification Alarm] ⚠️
04/28/2022 22:01:44 - DEBUG: Message Body: Datastore usage on disk VEBA-DS-01 has reached warning threshold.
Please log in to source-123 and ensure that everything is operating as expected.
      vCenter Server: source-123
       Datacenter: DataCenter1
       Datastore: VEBA-DS-01
04/28/2022 22:01:44 - DEBUG: custom prop: notify_email
04/28/2022 22:01:44 - DEBUG: email Key: 101
04/28/2022 22:01:44 - INFO: Datastore VEBA-DS-01 has Custom Field: notify_email with value: notify2@vmweventbroker.io

04/28/2022 22:01:44 - DEBUG: Found key 101 with value notify2@vmweventbroker.io
04/28/2022 22:01:44 - Sending notification to notify1@vmweventbroker.io notify2@vmweventbroker.io  ...

04/28/2022 22:01:45 - datastore-usage-email operation complete ...

04/28/2022 22:01:45 - Handler Processing Completed ...
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

Update the `datastore_secret.json` file with your vCenter Server credentials and configurations and then create the kubernetes secret which can then be accessed from within the function by using the environment variable named called `DATASTORE_SECRET`.

```console
# create secret
kubectl -n vmware-functions create secret generic datastore-secret --from-file=DATASTORE_SECRET=datastore_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret function-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `AlarmStatusChangedEventt` vCenter Server Event. If you wish to change this, update the `subject` field within `function.yaml` to the desired event type.


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