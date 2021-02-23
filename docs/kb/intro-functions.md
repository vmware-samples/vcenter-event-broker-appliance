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
    - text: Install the [Appliance with Knative](install-knative) to extend your SDDC with our [community-sourced functions](/examples-knative)
    - text: Install the [Appliance with OpenFaaS](install-openfaas) to extend your SDDC with our [community-sourced functions](/examples)
    - text: Learn more about the [Events in vCenter](vcenter-events) and the [Event Specification](eventspec) used to send the events to a Function
    - text: Find steps to deploy a function - [instructions](use-functions).
---

# Functions

The VMware Event Broker Appliance can be deployed using either Knative or OpenFaaS event processor to provides customers with a Function-as-a-Service (FaaS) platform.

## Table of Contents
- [Knative](#knative)
  - [Knative Naming and Version Control](#knative-naming-and-version-control)
    - [Knative Service](#knative-service)
    - [Knative Trigger](#knative-service)
    - [Knative Combined Service and Trigger](#knative-combined-service-and-trigger)
    - [Knative Secrets](#knative-secrets)
    - [Knative Environment Variables](#knative-environment-variables)
- [OpenFaaS](#openfaas)
  - [OpenFaaS Naming and Version Control](#openfaas-naming-and-version-control)

## Knative

Users who directly want to jump into VMware vSphere-related function code might want to check out the examples we provide [here](/examples/knative).

### Knative Naming and Version Control

When it comes to authoring functions, it's important to understand the different components that make up a Knative function deployment. Let's take the following excerpt as an example:

### Knative Service

A Knative Service `kn-service.yaml` defines the container image that will be executed upon invocation.

```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
 name: kn-echo
spec:
 template:
  metadata:
    annotations:
      autoscaling.knative.dev/maxScale: "1"
      autoscaling.knative.dev/minScale: "1"
  spec:
   containers:
    - image: embano1/kn-echo:latest
```

`kn-echo`: The name of the Knative Service.

The value of this field:

- must not conflict with an existing function
- should not contain special characters, e.g. "$" or "/"
- should represent the intent of the function, e.g. "tag" or "tagging"

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
  name: veba-echo-trigger
spec:
  broker: rabbit
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-echo
```

`veba-echo-trigger`: The name of the Knative trigger

The value of this field:

- must not conflict with an existing trigger
- should not contain special characters, e.g. "$" or "/"
- should represent the intent of the trigger, e.g. "tag" or "tagging"

`rabbit`: The name of Knative broker. For VEBA with Embedded Knative Broker, the value will be `rabbit`

`kn-echo`: The name of the Knative Service

To subscribe to a specific vCenter Server event, we can apply a filtering to our Knative Trigger like the example below:

```
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-echo-trigger
spec:
  broker: rabbit
  filter:
    attributes:
       type: com.vmware.event.router/event
       subject: VmPoweredOffEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-echo
```

`veba-echo-trigger`: The name of the Knative trigger

The value of this field:

- must not conflict with an existing trigger
- should not contain special characters, e.g. "$" or "/"
- should represent the intent of the trigger, e.g. "tag" or "tagging"

`rabbit`: The name of Knative broker. For VEBA with Embedded Knative Broker, the value will be `rabbit`

`VmPoweredOffEvent`: The name of the vCenter Server event Id. Please refer to [vCenter Events](vcenter-events) for list of supported events.

`kn-echo`: The name of the Knative Service

> **Note:** Today, a single Knative Trigger can only filter on one vCenter Server event. To associate multiple vCenter Server events to a given Knative Service, you simply create a Knative Trigger for each event.

## Knative Combined Service and Trigger

To simplify Knative function deployment, we can also combine the multiple manifest files into a single file. In the example below, I have `function.yaml` which contains both my Knative Service and Trigger definition.

```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
 name: kn-echo
spec:
 template:
  metadata:
    annotations:
      autoscaling.knative.dev/maxScale: "1"
      autoscaling.knative.dev/minScale: "1"
  spec:
   containers:
    - image: embano1/kn-echo:latest
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-echo-trigger
spec:
  broker: rabbit
  filter:
    attributes:
       type: com.vmware.event.router/event
       subject: VmPoweredOffEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-echo
```

To deploy the Knative Service/Trigger, just run:

```
kubectl apply -f function.yaml
```

To undeploy the Knative Service/Trigger, just run:

```
kubectl delete -f function.yaml
```

## Knative Secrets

Knative functions also support secrets or sensitive information which are passed in as part of the function deployment. Within the function, you can then access the secrets using an environmental variable and the name that the secret reference was created with.

A secrets file (without file extension) should be created that contains the sensitive values including the structure you wish to process within your function. For example, you can encode your secrets into a JSON structure which means you can use the language specific JSON parser to easily extract out specific values.

In the example below, I have a file called `secret`, this can be named anything but make sure it does not contain a file extension.
```
cat > secret <<EOF
{
  "SLACK_WEBHOOK_URL": "YOUR-WEBHOOK-URL"
}
EOF
```

Next, we need to create the Kubernetes secret by using the `--from-file` option and passing in the name of the file as well as the name of the secret which in this example is called `slack-secret`
```
# create secret
kubectl create secret generic slack-secret --from-file secret
```

To use the newly created Kubernetes secret, we will need to add a new section to our Knative Service called `envFrom` as shown in the example below.

```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
 name: kn-echo
spec:
 template:
  metadata:
    annotations:
      autoscaling.knative.dev/maxScale: "1"
      autoscaling.knative.dev/minScale: "1"
  spec:
   containers:
    - image: embano1/kn-echo:latest
      envFrom:
        - secretRef:
            name: slack-secret
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-echo-trigger
spec:
  broker: rabbit
  filter:
    attributes:
       type: com.vmware.event.router/event
       subject: VmPoweredOffEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-echo
```

Finally, to access the secret from within your function handler, use the language specific option to access the environmental variable called `slack-secret` and decode the contents.

## Knative Environment Variables

Knative functions also support defining additional environmental variables that can be passed in as part of the function deployment. To do so, define a new `env:` section with a list of name and values. In the example below, I have defined `function_debug` and the vale of `"true"` (notice this must be an encapsulated string).

From within your function handler, you can now access the environmental variable called `function_debug`. Using additional variables can be useful for debugging/troubleshooting purposes and can easily be changed by updating your Knative function deployment.

```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
 name: kn-echo
spec:
 template:
  metadata:
    annotations:
      autoscaling.knative.dev/maxScale: "1"
      autoscaling.knative.dev/minScale: "1"
  spec:
   containers:
    - image: embano1/kn-echo:latest
      envFrom:
        - secretRef:
            name: slack-secret
      env:
        - name: function_debug
          value: "true"
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-echo-trigger
spec:
  broker: rabbit
  filter:
    attributes:
       type: com.vmware.event.router/event
       subject: VmPoweredOffEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-echo
```

---

## OpenFaaS

Alex Ellis, the creator of OpenFaaS, and the community have put together comprehensive documentation and workshop materials to get you started with writing your first functions:

- [OpenFaaS Workshop](https://docs.openfaas.com/tutorials/workshop/){:target="_blank"}
- [Your first OpenFaaS Function with Python](https://docs.openfaas.com/tutorials/first-python-function/){:target="_blank"}
- [Writing your first Serverless function](https://medium.com/@pkblah/writing-your-first-serverless-function-23508cb4ea11?source=friends_link&sk=90cbed9b0dadb67578cebe54a88df494){:target="_blank"}
- [Serverless Function - Quickstart templates](https://medium.com/@pkblah/serverless-function-templates-available-2642bb92f58b?source=friends_link&sk=888a695eb9b4c1105f2bedc8478700b1){:target="_blank"}

Users who directly want to jump into VMware vSphere-related function code might want to check out the examples we provide [here](/examples/openfaas).
### OpenFaaS Naming and Version Control

When it comes to authoring functions, it's important to understand how the different fields in the OpenFaaS function's stack definition, e.g. `stack.yml`, are used throughout the appliance. Let's take the following excerpt as an example:

```yaml
# stack.yaml snippet
[...]
functions:
  pytag-fn:
    lang: python3
    handler: ./handler
    image: embano1/pytag-fn:0.2
```

`pytag-fn:` The name of the function used by OpenFaaS as the canonical name and identifier throughout the lifecycle of the function. Internally this will be the name used by Kubernetes to run the function as a Kubernetes deployment.

<!-- TODO: clarify deployment/pod via OpenFaaS -->

The value of this field:

- must not conflict with an existing function
- should not contain special characters, e.g. "$" or "/"
- should represent the intent of the function, e.g. "tag" or "tagging"
- may use a major version suffix, e.g. "pytag-fn-v3" in case of breaking changes/when multiple versions of the function need to run in parallel for backwards compatibility

`image:` The name of the resulting container image following Docker naming conventions `"<repo>/<image>:<tag>"`. OpenFaaS uses this field during the build and deployment phases, i.e. `faas-cli [build|deploy]`. Internally this will be the image pulled by Kubernetes when creating the function.

The value of this field:

- must resolve to a valid Docker container name (see convention above)
- should reflect the name of the function for clarity
- should use a tag other than `"latest"`, e.g. `":0.2"` or `":$GIT_COMMIT"`
- should be updated whenever changes to the function logic are made (before `faas-cli [build|deploy]`)
  - avoids overwriting the existing container image which ensures audibility and eases troubleshooting
  - supports common CI/CD version control flows
  - changing the tag is sufficient


> **Note:** `functions` can contain multiple functions described as a list in YAML (not shown here).
