# Installing and using PagerDuty Trigger Function in Go

## Function Purpose

A benefit of using VEBA is being able to integrate vSphere events with external
applications. This can be done by writing a handler function that makes an HTTP
request to an API. This function is an example written in Go that responds to a
`VmReconfiguredEvent`, sends information regarding the event to PagerDuty.

## Clone the VEBA repo

The VEBA repository contains many function examples, including go-pagerduty-trigger.
Clone it to get a local copy.

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/go/pagerduty-trigger
git checkout master
```

## PagerDuty

For go-pagerduty-trigger function to work, you'll need a PagerDuty account and
setup, which will not be covered here. PK goes over the setup in (a Medium blog)[https://medium.com/@pkblah/integrating-vcenter-with-pagerduty-241d3813871c].

## Customize the function

For security reasons, do not expose sensitive data. We will create a Kubernetes [secret](https://kubernetes.io/docs/concepts/configuration/secret/)
which will hold the PagerDuty routing key and trigger. This secret will be mounted
(by the appliance) into the function during runtime. The secret will need to be
created via `faas-cli`.

First, change the configuration file [pdconfig.toml](pdconfig.toml) holding your secret vCenter
information located in this folder:

```toml
# pdconfig.toml contents
# Replace with your own values.
[pagerduty]
routingKey = "<replace with your routing key>"
eventAction = "trigger"
```

Store the pdconfig.toml configuration file as secret in the appliance using the
following:

```bash
# Set up faas-cli for first use.
export OPENFAAS_URL=https://VEBA_FQDN_OR_IP

# Log into the OpenFaaS client. The username and password were set during the vCenter
# Event Broker Appliance deployment process. The username can be excluded in the
# command if the default, admin, was used.
faas-cli login -p VEBA_OPENFAAS_PASSWORD --tls-no-verify

# Create the secret
faas-cli secret create pdconfig --from-file=pdconfig.toml --tls-no-verify

# Update the secret if needed
faas-cli secret update pdconfig --from-file=pdconfig.toml --tls-no-verify
```

> **Note:** Delete the local `pdconfig.toml` after you're done with this exercise
to not expose this sensitive information.

Lastly, define the vCenter event which will trigger this function. Such function-
specific settings are performed in the `stack.yml` file. Open and edit the `stack.yml`
provided in the examples/go/pagerduty-trigger directory. Change `gateway` and `topic`
as per your environment/needs.

> **Note:** A key-value annotation under `topic` defines which VM event should
trigger the function. A list of VM events from vCenter can be found [here](https://code.vmware.com/doc/preview?id=4206#/doc/vim.event.VmEvent.html).
A single topic can be written as `topic: VmPoweredOnEvent`. Multiple topics can
be specified using a `","` delimiter syntax, e.g. "`topic: "VmPoweredOnEvent,VmPoweredOffEvent"`".

```yaml
version: 1.0
provider:
  name: openfaas
  gateway: https://veba.yourdomain.com # Replace with your VEBA environment.
functions:
  go-pagerduty-trigger-fn:
    lang: golang-http
    handler: ./handler
    image: vmware/veba-go-pagerduty-trigger:latest
    secrets:
      - pdconfig # Ensure this name matches the secret you created.
    annotations:
      topic: VmReconfiguredEvent
```

> **Note:** If you are running a vSphere DRS-enabled cluster the topic annotation
above should be `DrsVmPoweredOnEvent`. Otherwise, the function would never be triggered.

## Deploy the function

After you've performed the steps and modifications above, you can deploy the function:

```bash
faas template store pull golang-http # only required during the first deployment
faas-cli deploy -f stack.yml --tls-no-verify
Deployed. 202 Accepted.
```

## Trigger the function

1. Reconfigure the VM (e.g. change numCPU, memoryMB, cpuHotAddEnabled, etc.)
1. Log into PagerDuty to see the reaction to the trigger event.

> **Note:** If the function doesn't trigger a PagerDuty notification, then verify
that you correctly followed each step above. Ensure the IPs/FQDNs and credentials
are correct and see the [troubleshooting](#troubleshooting) section below.

## Troubleshooting

If you reconfigured your VM but don't see a PagerDuty notification, verify:

- PagerDuty routing key
- Whether the components can talk to each other (VMware Event Router to vCenter
and OpenFaaS, function to vCenter)
- Check the logs:

```bash
faas-cli logs go-pagerduty-trigger-fn --follow --tls-no-verify

# Successful log message will look something similar to
POST / - 200 OK - ContentLength: 95
```
