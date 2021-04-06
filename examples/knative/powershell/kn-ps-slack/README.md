# kn-ps-slack
Simple Knative powershell function for sending Slack webhook
[CloudEvents](https://github.com/cloudevents/).

> **Note:** CloudEvents using structured or binary mode are supported.

# Step 1 - Build

Create the container image and optionally push to an external registry such as Docker.

```
docker build -t <docker-username>/kn-ps-slack:1.0 .
docker push -t <docker-username>/kn-ps-slack:1.0
```

# Step 2 - Test

Verify the container image works by executing it locally.

```bash
docker run -e FUNCTION_DEBUG=true -e PORT=8080 -e SLACK_SECRET='{"SLACK_WEBHOOK_URL": "YOUR-WEBHOOK-URL"}' -it --rm -p 8080:8080 <docker-username>/kn-ps-slack:1.0

# now in a separate window run the following

# Run either test.ps1 (PowerShell Script) or test.sh (Bash Script) to simulate a CloudEvent payload being sent to the container image
./test.ps1

Testing Function ...
See docker container console for output

# Output from docker container console
Detected change to subject-123 ...
Sending Webhook payload to Slack ...
StatusCode        : 200                                                                                                                                                                                                            StatusDescription : OK                                                                                                                                                                                                             Content           : ok
RawContent        : HTTP/1.1 200 OK
                    Date: Tue, 16 Feb 2021 21:56:16 GMT
                    Server: Apache
                    Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
                    x-slack-backend: r
                    referrer-policy: no-referrer
                    Vary: Accept-…
Headers           : {[Date, System.String[]], [Server, System.String[]], [Strict-Transport-Security, System.String[]], [x-slack-backend, System.String[]]…}
Images            : {}
InputFields       : {}
Links             : {}
RawContentLength  : 2
RelationLink      : {}
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit broker. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the broker is also available there.

Update the `secret` file with your Slack webhook URL and then create the kubernetes secret which can then accessed from within the function by using the environmental variable named called `SLACK_SECRET`.

```bash
# create secret

kubectl create secret generic slack-secret --from-file=SLACK_SECRET=secret
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `VmPoweredOffEvent` vCenter Server Event. If you wish to change this, update the `subject` field within `function.yaml` to the desired event type.


Deploy the function to the VMware Event Broker Appliance (VEBA).

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