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
    - text: Install the [Appliance with OpenFaaS](install-openfaas) to extend your SDDC with our [community-sourced functions](/examples)
    - text: Learn more about the [Events in vCenter](vcenter-events) and the [Event Specification](eventspec) used to send the events to a Function
    - text: Find steps to deploy a function - [instructions](use-functions).
---

# Functions

## Getting Started

The VMware Event Broker Appliance deployed with OpenFaaS provides a Function-as-a-Service (FaaS) platform. Alex Ellis, the creator of OpenFaaS, and the community have put together comprehensive documentation and workshop materials to get you started with writing your first functions:

- [OpenFaaS Workshop](https://docs.openfaas.com/tutorials/workshop/){:target="_blank"}
- [Your first OpenFaaS Function with Python](https://docs.openfaas.com/tutorials/first-python-function/){:target="_blank"}
- [Writing your first Serverless function](https://medium.com/@pkblah/writing-your-first-serverless-function-23508cb4ea11?source=friends_link&sk=90cbed9b0dadb67578cebe54a88df494){:target="_blank"}
- [Serverless Function - Quickstart templates](https://medium.com/@pkblah/serverless-function-templates-available-2642bb92f58b?source=friends_link&sk=888a695eb9b4c1105f2bedc8478700b1){:target="_blank"}

Users who directly want to jump into VMware vSphere-related function code might want to check out the examples we provide [here](/examples).

## Naming and Version Control

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

