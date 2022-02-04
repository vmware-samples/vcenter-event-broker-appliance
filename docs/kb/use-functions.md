---
layout: docs
toc_id: use-functions
title: VMware Event Broker Appliance - Using Functions
description: VMware Event Broker Appliance - Using Functions
permalink: /kb/use-functions
cta:
 title: What's next?
 description: Extend your vCenter quickly with our pre-built functions
 actions:
    - text: See our complete list of prebuilt functions - [here](/examples)
    - text: Learn how to write your own function - [here](contribute-functions).
---

# Getting started with using functions

The steps below describe a generalized deployment step of a function on the VMware Event Broker Appliance. For customers looking to get started quickly, please look at deploying from our growing list of [Prebuilt Functions](/examples). The functions are organized by the language that they are written in and have well-documented README.md files with detailed deployment steps.

Deployment or development of functions can require the use of the following development tools: `git`, Docker, and `kubectl`.  If you would like some help in how to setup your workstation to use those tools, an in-depth tutorial is available that shows how to do that and to modify and deploy the [kn-ps-slack](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/examples/knative/powershell/kn-ps-slack) function (knative powershell slack notification function) as an example.  The tutorial can be found here: [in-depth function tutorial](function-tutorial-intro).


## Table of Contents
- [Knative](#knative)
  - [Knative Prerequisites](#knative-prerequisites)
  - [Knative Function Deployment using kubectl](#knative-function-deployment-using-kubectl)
  - [Knative Function Deployment using vSphere UI](#knative-function-deployment-using-vsphere-ui)

## Knative

For this walk-through, the [`kn-ps-echo`](/examples-knative) function from the example folder is used.

### Knative Prerequisites

Before proceeding to deploy a function, you must have VMware Event Broker Appliance deployed and be able to SSH to the appliance or have access to the kubernetes configuration file (/root/.kube/config) locally on your desktop to authenticate


### Knative Function Deployment using kubectl
Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/knative/powershell/kn-ps-echo
git checkout master
```

Step 2 - Edit the configuration files

* Edit `function.yml` to update `subject:` with the specific vCenter Event. All available events can be reviewed in the [vCenter Event Mapping](https://github.com/lamw/vcenter-event-mapping){:target="_blank"} document.

```yaml
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
      subject: DrsVmPoweredOnEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-ps-echo
```

Step 3 - Deploy function to VMware Event Broker Appliance

```
kubectl -n vmware-functions apply -f function.yaml
```

Step 4 - Test your functions

* Your function is now deployed and available for the VMware Event Router to invoke when it sees a matching event

### Knative Function Deployment using vSphere UI

Step 1 - Login to the vSphere UI and navigate to the VMware Event Broker Appliance UI by going to `Menu->VMware Event Broker`.

Step 2 (Optional) - If a secret is required for a function. Select the `Secrets` tab and then click Add to create a new Kubernetes secret.

Step 3 - Under the `Events` tab, search and filter for the specific vSphere Event to deploy a function. One an event has been identified, click on the `>>` icon to expand and click on Deploy button to begin the guided wizard.

Step 4 - In the `Function` section of the wizard, users will need to provide the following required parameters:

* `Function name` - A name for the function to be deployed

* `Docker Image` - The Docker image that contains the function logic

Step 5 - (optional) In the `Configuration` section of the wizard, users can control the concurrency and scale bounds for the function:

* `Container concurrency:` - Specifies the maximum allowed in-flight (concurrent) requests per container of the Revision. Defaults to 0 which means concurrency to the application is not limited, and the system decides the target concurrency for the autoscaler.

* `Minimum scale:` - Controls the minimum number of replicas that each Revision should have. Knative will attempt to never have less than this number of replicas at any one point in time. Defaults to 0 which means scale to zero.

* `Maximum scale:` - Controls the maximum number of replicas that each revision should have. Knative will attempt to never have more than this number of replicas running, or in the process of being created, at any one point in time. Defaults to 0 which means unlimited scale.

Step 6 - (optional) In the `Configure Variables` section of the wizard, users can define new environment variables that can be used within their function

* `Name` - The name should begin with an uppercase letter and contain only uppercase letters(A-Z), numbers(0-9) and underscores(_)

* `Value` - The value for the environment variable

Step 7 - (optional) In the `4 Configure Secrets` section of the wizard, users can select specific Kubernetes Secrets that were created earlier in Step 2 for use within their function

Step 8 - Finally, review the summary for the function deployment and click `Finish` to begin the function deployment.

To verify the function deployment was successful, click on the `Functions` tab and ensure the function status `True` for the following:

* `ConfigurationsReady` : True

* `Ready` : True

* `RoutesReady` : True

* Most functions also have a secrets configuration file that you must edit to match your environment. For the `hostmaint-alarms` function, the file is named `vc-hostmaint-config.json`
```json
{
    "VC" : "https://veba.primp-industries.com",
    "VC_USERNAME" : "veba@vsphere.local",
    "VC_PASSWORD" : "FillMeIn"
}
```
Then create the secret in OpenFaaS with this command:
```bash
faas-cli secret create vc-hostmaint-config --from-file=vc-hostmaint-config.json
```


Step 3 - Deploy function to VMware Event Broker Appliance

```
faas-cli deploy -f stack.yml
```

Step 4 - Test and Invoke your functions

* Your function is now deployed to OpenFaaS and available for VMware Event Router to invoke when it sees a matching event
* You can also test or invoke your functions using the http endpoint for the function that OpenFaaS makes available. Pass the expected CloudEvents to the function as the http request parameter
