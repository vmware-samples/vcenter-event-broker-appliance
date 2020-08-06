# Installing and Using the VM-Reconfig-Via-Tag Function

## Function Purpose

Many reconfigurations of a virtual machine require a VM to be powered off. For
example, if CPU hot add has not been enabled, a VM will need to be powered off
before CPUs can be added. Enabling the CPU hot add also requires a power off. This
function makes it easier to reconfigure VMs by using tags to delay a reconfiguration
event until it is convenient to power down. When a `VmPoweredOff` event occurs, this
function will be triggered. The VM will be reconfigured automatically using attached
tags containing configuration information.

In preparation for this function, categories with the configuration paths, e.g.
`config.hardware.numCPU` will need to be created. Within the categories, tags with
the configuration values will also need to be created. Attach tags with the desired
configuration(s) to the VM. Although tags from multiple category types can be attached,
ensure only one of each category type tag is attached to a VM.

## Clone the VEBA repo

The VEBA repository contains many function examples, including vm-reconfig-via-tag.
Clone it to get a local copy.

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/go/vm-reconfig-via-tag
git checkout master
```

## Create configuration categories and tags

For vm-reconfig-via-tag function to work, we need to create categories and tags
that contain desired configuration information.

Create categories with names taken from the following list (only the ones you want
need to be created):

- config.hardware.numCPU
- config.hardware.memoryMB
- config.hardware.numCoresPerSocket
- config.memoryHotAddEnabled
- config.cpuHotRemoveEnabled
- config.cpuHotAddEnabled

Ensure that only one category type can be attached to the VM. This can be done
by choosing "One tag" for "Tags Per Object" in the vSphere client UI.

Then, create tags where the tag names are the values of the configurations. For
example, if you want to be able to set the number of CPUs to 2, create a tag with
the name "2" (don't include the quotes in the name).

You can use the vSphere client user interface to create tag categories and tags,
but you can also use the govmomi client (`govc`) from the command line. Download
`govc` for your operating system from the (govc releases page)[https://github.com/vmware/govmomi/releases].

```bash
# Allow for a non-TLS secure connection to govc. Insecure is needed if vCenter
# certificates cannot be verified.
export GOVC_INSECURE=true

# Set the vSphere credentials.
export GOVC_URL='https://vcuser:vcpassword@vcenter.ip'

# Test out the connection by listing the tags.
./govc tags.ls

# Create categories and tags that will be used by vm-reconfig-via-tag function.
# the -m=false flag indicates that only one category type can be attached to an
# object.
./govc tags.category.create -m=false config.hardware.numCPU
./govc tags.create -c config.hardware.numCPU 2
```

## Customize the function

For security reasons, do not expose sensitive data. We will create a Kubernetes
[secret](https://kubernetes.io/docs/concepts/configuration/secret/) which will
hold the vCenter credentials and tag information. This secret will be mounted
(by the appliance) into the function during runtime. The secret will need to be
created via `faas-cli`.

First, change the configuration file [vcconfig.toml](vcconfig.toml) holding your
secret vCenter information located in this folder:

```toml
# vcconfig.toml contents
# Replace with your own values and use a dedicated user/service account with
# permissions to tag VMs, if possible. Insecure indicates if TLS self-signed
# certificates is being enforced. Insecure = true means TLS is not enforced.
[vcenter]
server = "VCENTER_FQDN/IP"
user = "admin@vsphere.local"
password = "DontUseThisPassword"
insecure = true # If not set, false is the default.
```

Store the vcconfig.toml configuration file as secret in the appliance using the
following:

```bash
# set up faas-cli for first use
export OPENFAAS_URL=https://VEBA_FQDN_OR_IP

# Log into the OpenFaaS client. The username and password were set during the vCenter
# Event Broker Appliance deployment process. The username can be excluded in the
# command if the default, admin, was used.
faas-cli login -p VEBA_OPENFAAS_PASSWORD --tls-no-verify

# Create the secret
faas-cli secret create vcconfig --from-file=vcconfig.toml --tls-no-verify

# Update the secret if needed
faas-cli secret update vcconfig --from-file=vcconfig.toml --tls-no-verify
```

> **Note:** Delete the local `vcconfig.toml` after you're done with this exercise
to not expose this sensitive information.

Lastly, define the vCenter event which will trigger this function. Such function-
specific settings are performed in the `stack.yml` file. Open and edit the `stack.yml`
provided in the examples/go/vm-reconfig-via-tag directory. Change `gateway` and
`topic` as per your environment/needs.

> **Note:** A key-value annotation under `topic` defines which VM event should
trigger the function. A list of VM events from vCenter can be found [here](https://code.vmware.com/doc/preview?id=4206#/doc/vim.event.VmEvent.html).
A single topic can be written as `topic: VmPoweredOnEvent`. Multiple topics can
be specified using a `","` delimiter syntax, e.g. "`topic: "VmPoweredOnEvent,VmPoweredOffEvent"`".

```yaml
version: 1.0
provider:
  name: openfaas
  gateway: https://VEBA_FQDN_OR_IP # Replace with your VEBA environment.
functions:
  vm-reconfig-via-tag-fn:
    lang: golang-http
    handler: ./handler
    image: vmware/veba-go-vm-reconfig-via-tag:latest
    environment:
      write_debug: true # Enables verbose logging. Default is false.
    secrets:
      - vcconfig # Ensure this name matches the secret you created.
    annotations:
      topic: VmPoweredOnEvent
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

1. Turn on a virtual machine.

1. Attach one or more tags from different categories to the VM with the desired
configuration.

1. Turn off the VM. The desired configuration should have been set (or if already
set, nothing will happen).

> **Note:** If the desired configuration(s) have been set, then verify that you
correctly followed each step above. Ensure the IPs/FQDNs and credentials are correct
and see the [troubleshooting](#troubleshooting) section below.

## Troubleshooting

If your VM did not get reconfigured after a powered off event, verify:

- vCenter IP/username/password
- Permissions of the vCenter user
- Whether the components can talk to each other (VMware Event Router to vCenter
and OpenFaaS, function to vCenter)
- Check the logs (first set `write_debug: true` in stack.yml to enable verbose
logging):

```bash
faas-cli logs vm-reconfig-via-tag-fn --follow --tls-no-verify

# Successful log message will look something similar to
Task:task-11158 set cpuHotAddEnabled to true.
POST / - 200 OK - ContentLength: 45
```
