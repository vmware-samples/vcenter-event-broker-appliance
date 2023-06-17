# kn-ps-harbor-slack

Example Knative PowerShell function for sending Harbor CloudEvents to a Slack webhook. This function relies on the Harbor webhook [function example](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/development/examples/knative/go/kn-go-harbor-webhook) which is a requirement for this example.

# Step 1 - Build

Create the container image locally to test your function logic. Change the IMAGE name accordingly, example below for Docker.

```console
export TAG=<version>
export IMAGE=<docker-username>/kn-ps-#REPLACE-FN-NAME#:${TAG}
docker build -t ${IMAGE}
```

# Step 2 - Test

Verify the container image works by executing it locally.

Change into the `test` directory

```console
cd test
```

Update the following variable names within the `docker-test-env-variable` file

* SLACK_WEBHOOK_URL - Slack webhook URL
* SLACK_MESSAGE_PRETEXT - Text displayed for Slack notification

Start the container image by running the following command:

```console
docker run -e FUNCTION_DEBUG=true -e PORT=8080 --env-file docker-test-env-variable -it --rm -p 8080:8080 ${IMAGE}
```

In a separate terminal, run either `send-cloudevent-test.ps1` (PowerShell Script) or `send-cloudevent-test.sh` (Bash Script) to simulate a CloudEvent payload being sent to the local container image

```console
Testing Function ...
See docker container console for output

# Output from docker container console
06/27/2022 09:47:31 - DEBUG: K8s Secrets:
{"SLACK_WEBHOOK_URL":"**********","SLACK_MESSAGE_PRETEXT":":harbor: Harbor Slack Function :veba_official:"}

06/27/2022 09:47:31 - DEBUG: CloudEventData

Name                           Value
----                           -----
event_data                     {resources, repository}
occur_at                       1656076946
type                           PUSH_ARTIFACT
operator                       admin



06/27/2022 09:47:31 - DEBUG: "{
  "attachments": [
    {
      "footer_icon": "https://raw.githubusercontent.com/vmware-samples/vcenter-event-broker-appliance/development/logo/veba_icon_only.png",
      "footer": "Powered by VEBA",
      "pretext": ":harbor: Harbor Slack Function :veba_official:",
      "fields": [
        {
          "short": "false",
          "value": "PUSH_ARTIFACT",
          "title": "Event Type"
        },
        {
          "short": "false",
          "value": "2022-06-25T11:42:42+00:00",
          "title": "DateTime in UTC"
        },
        {
          "short": "false",
          "value": "admin",
          "title": "Username"
        },
        {
          "short": "false",
          "value": "veba-webhook/bitnami-nginx",
          "title": "Repository Name"
        },
        {
          "short": "false",
          "value": "public",
          "title": "Repository Type"
        },
        {
          "short": "false",
          "value": "1.21.6-debian-10-r117",
          "title": "Image Tag"
        },
        {
          "short": "false",
          "value": "harbor.jarvis.tanzu/veba-webhook/bitnami-nginx:1.21.6-debian-10-r117",
          "title": "Image Resource Data"
        },
        {
          "short": "false",
          "value": "sha256:d3890814cc5a7cfc02403435281cdf51adfb6b67e223934d9d6137a4ad364286",
          "title": "Image Digest"
        }
      ]
    }
  ]
}"
06/27/2022 09:47:31 - Sending Webhook payload to Slack ...
06/27/2022 09:47:31 - Successfully sent Webhook ...
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit `broker`. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the `broker` is also available there.
>
> **Note:** Also, in order to receive incoming Harbor events and to ultimately invoke the Harbor-Slack-Function, it's necessary to have the [kn-go-harbor-webhook function example](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/development/examples/knative/go/kn-go-harbor-webhook) running properly first.

Update the `slack_secret.json` file with your Slack webhook configurations and then create the kubernetes secret which can then be accessed from within the function by using the environment variable named called `SLACK_SECRET`.

```console
# create secret

kubectl -n vmware-functions create secret generic harbor-slack-secret --from-file=SLACK_SECRET=slack_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret harbor-slack-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `com.vmware.harbor.push_artifact.v0` Harbor Event. If you wish to change this, update the `type` field within `function.yaml` to the desired event type. A list of supported notification events is available on the official Harbor documentation under [Configure Webhook Notifications](https://goharbor.io/docs/2.5.0/working-with-projects/project-configuration/configure-webhooks/). Furthermore, use the VEBA Event viewer endpoint (`https://<VEBA-FQDN>/events`) to display all incoming events.

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
kubectl -n vmware-functions delete secret harbor-slack-secret
```