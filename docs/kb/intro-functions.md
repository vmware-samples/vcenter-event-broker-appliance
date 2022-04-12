---
layout: docs
toc_id: intro-functions
title: VMware Event Broker Appliance - Intro to Functions
description: VMware Event Broker Appliance - Intro to Functions
permalink: /kb/functions
cta:
 title: Get Started
 description: Extend your vCenter seamlessly with our pre-built functions
 actions:
    - text: Install the [Appliance with Knative](install-knative) to extend your SDDC with our [community-sourced functions](/examples)
    - text: Learn more about the [Events in vCenter](vcenter-events) and the [Event Specification](eventspec) used to send the events to a Function
    - text: Find steps to deploy a function - [instructions](use-functions).
---

# Functions

The VMware Event Broker Appliance can be deployed using the Knative event processor to provide customers with a Function-as-a-Service (FaaS) platform.

## Table of Contents
- [Knative](#knative)
  - [Knative Naming and Version Control](#knative-naming-and-version-control)
    - [Knative Service](#knative-service)
    - [Knative Trigger](#knative-trigger)
    - [Knative Combined Service and Trigger](#knative-combined-service-and-trigger)
    - [Knative Secrets](#knative-secrets)
    - [Knative Environment Variables](#knative-environment-variables)

## Knative

Users who directly want to jump into VMware vSphere-related function code might want to check out the examples we provide [here](/examples).

### Knative Naming and Version Control

When it comes to authoring functions, it's important to understand the different components that make up a Knative function deployment. Let's take the following excerpt as an example:

### Knative Service

A Knative Service `kn-service.yaml` defines the container image that will be executed upon invocation.

```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
 name: kn-ps-echo
 namespace: vmware-functions
 labels:
   app: veba-ui
spec:
 template:
  spec:
   containers:
    - image: projects.registry.vmware.com/veba/kn-ps-echo:1.0
```

`kn-ps-echo`: The name of the Knative Service.

The value of this field:

- must not conflict with an existing function
- should not contain special characters, e.g. "$" or "/"
- should represent the intent of the function, e.g. "tag" or "tagging"

`app: veba-ui`: The Kubernetes label that is required for the VMware Event Broker Appliance UI to display a manually deployed Knative Service

`vmware-functions`: The Kubernetes namespace to deploy all functions to. By default, this is `vmware-functions` which is automatically created as part of the VMware Event Broker Appliance setup.

> **Note**: It is recommended to deploy all functions to the default `vmware-functions` namespace since the VMware Event Broker Appliance UI is automatically been configured to manage all functions and secrets within this namespace. The Rabbit Broker also lives in the `vmware-functions` namespace and would otherwise break if functions are not deployed within this namespace

`image:` The name of the resulting container image following Docker naming conventions `"<repo>/<image>:<tag>"`.

The value of this field:

- must resolve to a valid Docker container name (see convention above)
- should reflect the name of the function for clarity
- should use a tag other than `"latest"`, e.g. `":0.2"` or `":$GIT_COMMIT"`
- should be updated whenever changes to the function logic are made
  - avoids overwriting the existing container image which ensures audibility and eases troubleshooting
  - supports common CI/CD version control flows
  - changing the tag is sufficient

## Knative Trigger

A Knative Trigger `kn-trigger.yaml` defines the vCenter Server events to subscribe from a given broker. By default, all events are subscribed to as shown in the example below.

```
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-ps-echo-trigger
  labels:
    app: veba-ui
spec:
  broker: default
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-ps-echo
```

`veba-ps-echo-trigger`: The name of the Knative trigger

The value of this field:

- must not conflict with an existing trigger
- should not contain special characters, e.g. "$" or "/"
- should represent the intent of the trigger, e.g. "tag" or "tagging"

`app: veba-ui`: The Kubernetes label that is required for the VMware Event Broker Appliance UI to display a manually deployed Knative Trigger

`default`: The name of Knative broker. For VEBA with Embedded Knative Broker, the value will be `default`

`kn-ps-echo`: The name of the Knative Service

To subscribe to a specific vCenter Server event, we can apply a filtering to our Knative Trigger like the example below:

```
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-ps-echo-trigger
  labels:
    app: veba-ui
spec:
  broker: default
  filter:
    attributes:
       type: com.vmware.event.router/event
       subject: VmPoweredOffEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-ps-echo
```

`veba-ps-echo-trigger`: The name of the Knative trigger

The value of this field:

- must not conflict with an existing trigger
- should not contain special characters, e.g. "$" or "/"
- should represent the intent of the trigger, e.g. "tag" or "tagging"

`veba-ui`: The Kubernetes label that is required for the VMware Event Broker Appliance UI to display manually deployed Knative Trigger

`default`: The name of Knative broker. For VEBA with Embedded Knative Broker, the value will be `default`

`VmPoweredOffEvent`: The name of the vCenter Server event Id. Please refer to [vCenter Events](vcenter-events) for list of supported events.

`kn-ps-echo`: The name of the Knative Service

Today, a single Knative Trigger can only filter on one vCenter Server event. To associate multiple vCenter Server events to a given Knative Service, you simply create a Knative Trigger for each event as shown in the two examples below, one for `VmPoweredOffEvent` and `DrsVmPoweredOnEvent` vCenter Events respectively:

`kn-trigger-1.yaml`
```
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-ps-echo-trigger-1
  labels:
    app: veba-ui
spec:
  broker: default
  filter:
    attributes:
       type: com.vmware.event.router/event
       subject: VmPoweredOffEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-ps-echo
```

`kn-trigger-2.yaml`
```
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-ps-echo-trigger-2
  labels:
    app: veba-ui
spec:
  broker: default
  filter:
    attributes:
       type: com.vmware.event.router/event
       subject: DrsVmPoweredOnEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-ps-echo
```

## Knative Combined Service and Trigger

To simplify Knative function deployment, we can also combine the multiple manifest files into a single file. In the example below, the `function.yaml` contains both the Knative Service and Trigger definition.

```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
 name: kn-ps-echo
 labels:
   app: veba-ui
spec:
 template:
  spec:
   containers:
    - image: projects.registry.vmware.com/veba/kn-ps-echo:1.0
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-ps-echo-trigger
  labels:
    app: veba-ui
spec:
  broker: default
  filter:
    attributes:
       type: com.vmware.event.router/event
       subject: VmPoweredOffEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-ps-echo
```

To deploy the Knative Service/Trigger, just run:

```
kubectl -n vmware-functions apply -f function.yaml
```

To undeploy the Knative Service/Trigger, just run:

```
kubectl -n vmware-functions delete -f function.yaml
```

## Knative Secrets

Knative functions also support secrets or sensitive information which are passed in as part of the function deployment. Within the function, you can then access the secrets using an environment variable and the name that the secret reference was created with.

A secrets file should be created that contains the sensitive values including the structure you wish to process within your function. For example, you can encode your secrets into a JSON structure which means you can use the language specific JSON parser to easily extract out specific values.

In the example below, the file containing your secret is called `secret`.
```
cat > secret <<EOF
{
  "SLACK_WEBHOOK_URL": "YOUR-WEBHOOK-URL"
}
EOF
```

Next, we need to create the Kubernetes secret by using the `--from-file` option, the name of the Kubernetes secret and the name of the environment variable `SLACK_SECRET` which will be accessed by your functional handler.

> **Note**: The environment variable name must only contain uppercase letters(A-Z), numbers(0-9) and underscores(_)

```
# create secret
kubectl -n vmware-functions create secret generic slack-secret --from-file=SLACK_SECRET=secret

# update label
kubectl -n vmware-functions label secret slack-secret app=veba-ui
```

> **Note:** The VMware Event Broker Appliance UI in the vSphere UI can also be used to view and manage Kubernetes secrets. However, to ensure Kubernetes secrets that were manually created is visible in the UI, you need to also upate the `app` label with the value of `veba-ui` which the VMware Event Broker Appliance UI will only filter on.

To use the newly created Kubernetes secret, we will need to add a new section to our Knative Service called `envFrom` as shown in the example below.

```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: kn-ps-slack
  labels:
    app: veba-ui
spec:
  template:
    metadata:
    spec:
      containers:
        - image: projects.registry.vmware.com/veba/kn-ps-slack:1.0
          envFrom:
            - secretRef:
                name: slack-secret
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-ps-slack-trigger
  labels:
    app: veba-ui
spec:
  broker: default
  filter:
    attributes:
      type: com.vmware.event.router/event
      subject: VmPoweredOffEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-ps-slack
```

Finally, to access the secret from within your function handler, use the language specific option to access the environment variable that you had named earlier called `SLACK_SECRET` and decode the contents.

Here is an example using the PowerShell language to access the environment variable:

```
$jsonSecrets = ${env:SLACK_SECRET} | ConvertFrom-Json
```

## Knative Environment Variables

Knative functions also support defining additional environment variables that can be passed in as part of the function deployment. To do so, define a new `env:` section with a list of name and values. In the example below, the additional variable is named `FUNCTION_DEBUG` and contains the vaue of `"true"` (notice this must be an encapsulated string).

> **Note**: The environment variable name must only contain uppercase letters(A-Z), numbers(0-9) and underscores(_)

From within your function handler, you can now access the environment variable called `FUNCTION_DEBUG`. Using additional variables can be useful for debugging/troubleshooting purposes and can easily be changed by updating your Knative function deployment.

```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
 name: kn-ps-slack
 labels:
   app: veba-ui
spec:
 template:
  metadata:
  spec:
   containers:
    - image: projects.registry.vmware.com/veba/kn-ps-echo:1.0
      envFrom:
        - secretRef:
            name: slack-secret
      env:
        - name: FUNCTION_DEBUG
          value: "true"
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-ps-echo-trigger
  labels:
    app: veba-ui
spec:
  broker: default
  filter:
    attributes:
       type: com.vmware.event.router/event
       subject: VmPoweredOffEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-ps-slack
```