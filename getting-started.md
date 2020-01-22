# Getting Started with vCenter Event Broker Appliance

## Appliance Deployment

### Requirements

* 2 vCPU and 8GB of memory for vCenter Event Broker Appliance
* vCenter Server 6.x or greater
* Account to login to vCenter Server (readOnly is sufficient)

**Step 1** - Download the vCenter Event Broker Appliance (OVA) from the [VMware Fling site](https://flings.vmware.com/vcenter-event-broker-appliance).

**Step 2** - Deploy the vCenter Event Broker Appliance OVA to your vCenter Server using the vSphere HTML5 Client. As part of the deployment you will be prompted to provide the following input:

*Networking*

  * Hostname - The FQDN of the vCenter Event Broker Appliance. If you do not have DNS in your environment, make sure the hostname provide is resolvable from your desktop which may require you to manually add a hosts entry. Proper DNS resolution is recommended
  * IP Address - The IP Address of the vCenter Event Broker Appliance
  * Netmask Prefix - CIDR Notation (e.g. 24 = 255.255.255.0)
  * Gateway - The Network Gateway address
  * DNS - DNS Server(s) that will be able to resolve to external sites such as Github for initial configuration. If you have multiple DNS Servers, input needs to be space separated.
  * DNS Domain - The DNS domain of your network

*Credentials*

  * Root Password - This is the OS root password for the vCenter Event Broker Appliance
  * OpenFaaS Password - This is the Admin password for OpenFaaS UI

*vSphere*

  * vCenter Server - This FQDN or IP Address of your vCenter Server that you wish to associate this vCenter Event Broker Appliance to for Event subscription
  * vCenter Username - The username to login to vCenter Server, as mentioned earlier, readOnly account is sufficient
  * vCenter Password - The password to the vCenter Username
  * Disable vCenter Server TLS Verification - If you have a self-signed SSL Certificate, you will need to check this box

*zDebug*

  * Debugging - When enabled, this will output a more verbose log file that can be used to troubleshoot failed deployments

**Step 3** - Power On the vCenter Event Broker Appliance after successful deployment. Depending on your external network connectivity, it can take a few minutes while the system is being setup. You can open the VM Console to view the progress. Once everything is completed, you should see an updated login banner for the various endpoints:

```
Appliance Status: https://[hostname]/status
Install Logs: https://[hostname]/bootstrap
OpenFaaS UI: https://[hostname]
```

**Note**: If you enable Debugging, the install logs endpoint will automatically contain the more verbose log entries.

**Step 4** - You can verify that everything was deployed correctly by opening a web browser to the OpenFaaS UI available on https://[hostname]/ and logging in with the Admin credentials (user:admin) you had specified as part of the OVA deployment.

At this point, you have successfully deployed the vCenter Event Broker Appliance and you are ready to start deploying your functions!


## Function Deployment

The following example walks you through the steps of deploying your first function. The function will apply a vSphere tag to virtual machines after a `VmPoweredOnEvent`. Please make sure to run this example in an environment that is suitable for this exercise, i.e. not production.

### Requirements

* vCenter Event Broker Appliance fully configured and running
* `git` to clone the function example ([Download](https://git-scm.com/downloads))
* `faas-cli` to deploy the function ([Download](https://github.com/openfaas/faas-cli#get-started-install-the-cli))
* `govc` to create/retrieve tag information ([Download](https://github.com/vmware/govmomi/releases))

### How it works

Functions can subscribe to events in vCenter through the `topic` [annotations](https://docs.openfaas.com/reference/yaml/#function-annotations) configured in the function deployment file (`stack.yml`). Based on these events a function can perform any action (i.e. business logic), such as tagging a VM, run post-processing scripts, audit to an external system, etc.

**Note:** The current version only allows one event per function. A simple workaround is to deploy the same function with different associated events.

vCenter events can be easily mapped to functions. For example, a `VmPoweredOnEvent` from vCenter would have a function `topic` annotation `vm.powered.on`.

**Note:** In a DRS-enabled cluster the annotation would be `drs.vm.powered.on`.

### Categories and tags

For this exercise we need to create a category and tag unless you want to use an existing tag to follow along.

Create a category/tag to be attached to a VM when it is powered on. Since we need the unique tag ID (i.e. vSphere URN) we will use [govc](https://github.com/vmware/govmomi/tree/master/govc) for this job. You can also use vSphere APIs (REST/SOAP) to retrieve the URN.

```bash
# Test connection to vCenter, ignore TLS warnings
export GOVC_INSECURE=true # only needed if vCenter certificates cannot be verified
export GOVC_URL='https://vcuser:vcpassword@vcenter.ip' # replace with your environment details
./govc tags.ls # should not error out, otherwise check parameters above

# If the connection is successful create a demo category/tag to be used by the function
./govc tags.category.create democat1
urn:... # we don't need the category URN for this example
./govc tags.create -c democat1 demotag1
urn:vmomi:InventoryServiceTag:019c0a9e-0672-48f5-ac2a-e394669e2916:GLOBAL
```

**Note:** You can also create the demo vSphere Category/Tag by using the vSphere UI. Once you have created the vSphere Tag, you can browse to the Tag inventory object and in the browser, you can copy the URN which will be in the format `urn:vmomi:InventoryServiceTag:<UUID>:GLOBAL`

Take a note of the `urn:...` for `demotag1` as we will need it for the next steps.

### Get the example function

Clone this repository which contains the example functions.

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/python/tagging
```

### Customize the function

For security reasons to not expose sensitive data we will create a Kubernetes [secret](https://kubernetes.io/docs/concepts/configuration/secret/) which will hold the vCenter credentials and tag information. This secret will be mounted into the function during runtime. This is all taken care of for your by the appliance. We only have to create the secret with a simple command through `faas-cli`.

First, change the configuration file `vcconfig.toml` holding your secret vCenter information located in the tagging example folder:

```toml
# vcconfig.toml contents
# replace with your own values and use a dedicated user/service account with permissions to tag VMs if possible
[vcenter]
server = "VCENTER_FQDN/IP"
user = "tagging-admin@vsphere.local"
password = "DontUseThisPassword"

[tag]
urn = "urn:vmomi:InventoryServiceTag:019c0a9e-0672-48f5-ac2a-e394669e2916:GLOBAL" # replace with the one noted above
action = "attach" # tagging action to perform, i.e. attach or detach tag
```

Now go ahead and store this configuration file as secret in the appliance.

```bash
# set up faas-cli for first use
export OPENFAAS_URL=https://VEBA_FQDN_OR_IP
faas-cli login -p VEBA_OPENFAAS_PASSWORD --tls-no-verify # vCenter Event Broker Appliance is configured with authentication, pass in the password used during the vCenter Event Broker Appliance deployment process

# now create the secret
faas-cli secret create vcconfig --from-file=vcconfig.toml --tls-no-verify
```

**Note:** Delete the local `vcconfig.toml` after you're done with this exercise to not expose this sensitive information.

Lastly, define the vCenter event which will trigger this function. Such function-specific settings are performed in the `stack.yml` file. Open and edit the `stack.yml` provided with in the Python tagging example code. Change `gateway` and `topic` as per your environment/needs.

```yaml
provider:
  name: openfaas
  gateway: https://VEBA_FQDN_OR_IP # replace with your vCenter Event Broker Appliance environment
functions:
  pytag-fn:
    lang: python3
    handler: ./handler
    image: embano1/pytag-fn:0.2
    environment:
      write_debug: true
      read_debug: true
    secrets:
      - vcconfig # leave as is unless you changed the name during the creation of the vCenter credentials secrets above
    annotations:
      topic: vm.powered.on # or drs.vm.powered.on in a DRS-enabled cluster
```

**Note:** If you are running a vSphere DRS-enabled cluster the topic annotation above should be `drs.vm.powered.on`. Otherwise the function would never be triggered.

### Deploy the function

After you've performed the steps and modifications above, you can go ahead and deploy the function:

```bash
faas-cli template pull # only required during the first deployment
faas-cli deploy -f stack.yml --tls-no-verify
Deployed. 202 Accepted.
```

### Trigger the function

Turn on a virtual machine, e.g. in vCenter or via `govc` CLI, to trigger the function via a `(DRS)VmPoweredOnEvent`. Verify the virtual machine was correctly tagged.

**Note:** If you don't see a tag being assigned verify that you correctly followed each step above, IPs/FQDNs and credentials are correct and see the [troubleshooting](#troubleshooting) section below.

## Mapping vCenter Events

A vCenter Server instance ships with a number of "default" Events but it can also include custom and extended events which maybe published by both 2nd and 3rd party solutions. In addition, with each version of vSphere, additional Events may also be included. For these reasons, it is very difficult to publish a single list containing all possible configurations.

To help, you can refer to this blog post [here](https://www.virtuallyghetto.com/2019/12/listing-all-events-for-vcenter-server.html) which includes a script to help extract all Events for a specific vCenter Server deployment including the vCenter Event ID, Type and Description.

## Troubleshooting

If your VM did not get the tag attached, verify:

- vCenter IP/username/password
- Permissions of the vCenter user
- Whether the components can talk to each other (connector to vCenter and OpenFaaS, function to vCenter)
- Check the logs (`kubectl` is installed and configured locally on the appliance)):

```bash
faas-cli logs pytag-fn --follow --tls-no-verify

# Successful log message in the OpenFaaS tagging function
2019/01/25 23:48:55 Forking fprocess.
2019/01/25 23:48:55 Query
2019/01/25 23:48:55 Path  /

{"status": "200", "message": "successfully attached tag on VM: vm-267"}
2019/01/25 23:48:56 Duration: 1.551482 seconds
```

Or via `kubectl` locally on the appliance:

```bash
kubectl -n openfaas logs deploy/vcenter-connector -f

# Successful log message in the OpenFaaS vCenter connector
2019/01/25 23:39:09 Message on topic: vm.powered.on
2019/01/25 23:39:09 Invoke function: pytag-fn
2019/01/25 23:39:10 Response [200] from pytag-fn
```

You can access appliance specific logs on the endpoint `https://VEBA_FQDN/boostrap`. For debug level information, turn on debugging during the appliance deployment process.
