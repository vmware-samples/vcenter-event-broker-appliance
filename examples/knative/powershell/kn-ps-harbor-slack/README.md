# kn-ps-harbor-slack

Example Knative PowerShell function for sending Harbor CloudEvents to a Slack webhook. This function only works with Harbor version 2.8.3 and onwards. For older Harbor versions, use the previous version of this function.

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
08/23/2023 13:58:21 - PowerShell HTTP server start listening on 'http://*:8080/'
08/23/2023 13:58:21 - Processing Init

08/23/2023 13:58:21 - Init Processing Completed

08/23/2023 13:58:21 - Starting HTTP CloudEvent listener
08/23/2023 13:58:24 - DEBUG: K8s Secrets:
{"SLACK_WEBHOOK_URL":"**************","SLACK_MESSAGE_PRETEXT":":harbor: Harbor Slack Function :veba_official:"}

08/23/2023 13:58:24 - DEBUG: CloudEventData

Name                           Value
----                           -----
repository                     {repo_type, repo_full_name, name, namespace…}
resources                      {System.Collections.Hashtable}

08/23/2023 13:58:24 - DEBUG: "{
  "attachments": [
    {
      "pretext": ":harbor: Harbor Slack Function :veba_official:",
      "fields": [
        {
          "title": "Event Type",
          "value": "harbor.artifact.pushed",
          "short": "false"
        },
        {
          "title": "DateTime in UTC",
          "value": "2023-08-22T15:57:41+00:00",
          "short": "false"
        },
        {
          "title": "Unique Identifier",
          "value": "291ee129-1d27-415c-bbe1-3ca45d5f230a",
          "short": "false"
        },
        {
          "title": "Username",
          "value": null,
          "short": "false"
        },
        {
          "title": "Repository Name",
          "value": "myapp/app",
          "short": "false"
        },
        {
          "title": "Repository Type",
          "value": "public",
          "short": "false"
        },
        {
          "title": "Image Tag",
          "value": "1.0",
          "short": "false"
        },
        {
          "title": "Image Resource Data",
          "value": "harbor-cloudevents.vmware.net/myapp/app:1.0",
          "short": "false"
        },
        {
          "title": "Image Digest",
          "value": "sha256:4d59a5bc7be95672d00edfd43622fc82f826bebbd5f497a7930a652031771ea8",
          "short": "false"
        }
      ],
      "footer": "Powered by VEBA",
      "footer_icon": "https://raw.githubusercontent.com/vmware-samples/vcenter-event-broker-appliance/development/logo/veba_icon_only.png"
    }
  ]
}"
08/23/2023 13:58:24 - Sending Webhook payload to Slack ...
08/23/2023 13:58:25 - Successfully sent Webhook ...
^C08/23/2023 13:59:53 - PowerShell HTTP Server stop requested. Waiting for server to stop
08/23/2023 13:59:53 - Processing Shutdown

08/23/2023 13:59:53 - Shutdown Processing Completed

08/23/2023 13:59:53 - PowerShell HTTP server stop requested
```

# Step 3 - Deploy

> **Note:** The following steps assume a working Knative environment using the
`default` Rabbit `broker`. The Knative `service` and `trigger` will be installed in the
`vmware-functions` Kubernetes namespace, assuming that the `broker` is also available there.

Update the `slack_secret.json` file with your Slack webhook configurations and then create the kubernetes secret which can then be accessed from within the function by using the environment variable named called `SLACK_SECRET`.

```console
# create secret

kubectl -n vmware-functions create secret generic harbor-slack-secret --from-file=SLACK_SECRET=slack_secret.json

# update label for secret to show up in VEBA UI
kubectl -n vmware-functions label secret harbor-slack-secret app=veba-ui
```

Edit the `function.yaml` file with the name of the container image from Step 1 if you made any changes. If not, the default VMware container image will suffice. By default, the function deployment will filter on the `harbor.artifact.pushed` Harbor Event. If you wish to change this, update the `type` field within `function.yaml` to the desired event type. A list of supported notification events is available on the official Harbor documentation under [Configure Webhook Notifications](https://goharbor.io/docs/2.8.0/working-with-projects/project-configuration/configure-webhooks/). Furthermore, use the VEBA Event viewer endpoint (`https://<VEBA-FQDN>/events`) to display all incoming events.

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