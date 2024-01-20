# kn-ps-google-chat
Example Knative PowerShell function for sending to a [Google Chat webhook](https://developers.google.com/chat/how-tos/webhooks) when a failed vCenter Server backup has been detected.

# Step 1 - Build


Create the container image locally to test your function logic.

```
export TAG=<version>
docker build -t <docker-username>/kn-ps-google-chat:${TAG} .
```

# Step 2 - Test

Verify the container image works by executing it locally.

Change into the `test` directory
```console
cd test
```

Update the following variable names within the `docker-test-env-variable` file

* GOOGLE_CHAT_WEBHOOK_URL - Google Chat webhook URL

Start the container image by running the following command:

```console
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 <docker-username>/kn-ps-google-chat:${TAG}
```

In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent payload being sent to the local container image

```console
Testing Function ...
See docker container console for output

# Output from docker container console
01/17/2024 00:22:13 - Sending message to Google Chat Webhook ...
01/17/2024 00:22:13 - Successfully sent Google Chat message ...
01/17/2024 00:23:25 - PowerShell HTTP Server stop requested. Waiting for server to stop
01/17/2024 00:23:25 - Processing Shutdown

01/17/2024 00:23:25 - Shutdown Processing Completed

01/17/2024 00:23:25 - PowerShell HTTP server stop requested
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit `broker`. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the `broker` is also available there.

Push your container image to an accessible registry such as Docker once you're done developing and testing your function logic.

```console
docker push <docker-username>/kn-ps-google-chat:${TAG}
```

Update the `google_chat_secret.json` file with your Google Chat webhook configurations and then create the kubernetes secret which can then be accessed from within the function by using the environment variable named called `GOOGLE_CHAT_SECRET`.

```console
# create secret

kubectl -n vmware-functions create secret generic google-chat-secret --from-file=GOOGLE_CHAT_SECRET=google_chat_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret google-chat-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `com.vmware.vsphere.com.vmware.applmgmt.backup.job.failed.event.v0` vCenter Server Event. If you wish to change this, update the `type` field within `function.yaml` to the desired event type.


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
kubectl -n vmware-functions delete secret google-chat-secret
```