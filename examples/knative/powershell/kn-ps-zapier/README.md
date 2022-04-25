# kn-ps-zapier
Example Knative PowerShell function for sending to a [Zapier webhook](https://zapier.com/page/webhooks/) when a failed vCenter Server login has been detected.

# Step 1 - Build

Create the container image locally to test your function logic.

```
export TAG=<version>
docker build -t <docker-username>/kn-ps-zapier:${TAG} .
```

# Step 2 - Test

Verify the container image works by executing it locally.

Change into the `test` directory
```console
cd test
```

Update the following variable names within the `docker-test-env-variable` file

* ZAPIER_WEBHOOK_URL - Zapier webhook URL

Start the container image by running the following command:

```console
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 <docker-username>/kn-ps-zapier:${TAG}
```

In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent payload being sent to the local container image

```console
Testing Function ...
See docker container console for output

# Output from docker container console
Detected change to subject-123 ...
04/22/2022 22:52:00 - PowerShell HTTP server start listening on 'http://*:8080/'
04/22/2022 22:52:00 - Processing Init

04/22/2022 22:52:00 - Init Processing Completed

04/22/2022 22:52:00 - Starting HTTP CloudEvent listener
04/22/2022 22:52:08 - Sending Webhook payload to Zapier ...
04/22/2022 22:52:09 - Successfully sent Webhook ...
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit `broker`. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the `broker` is also available there.

Push your container image to an accessible registry such as Docker once you're done developing and testing your function logic.

```console
docker push <docker-username>/kn-ps-zapier:${TAG}
```

Update the `zapier_secret.json` file with your Zapier webhook configurations and then create the kubernetes secret which can then be accessed from within the function by using the environment variable named called `ZAPIER_SECRET`.

```console
# create secret

kubectl -n vmware-functions create secret generic zapier-secret --from-file=ZAPIER_SECRET=zapier_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret zapier-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `com.vmware.sso.LoginFailure` vCenter Server Event. If you wish to change this, update the `subject` field within `function.yaml` to the desired event type.


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
kubectl -n vmware-functions delete secret zapier-secret
```