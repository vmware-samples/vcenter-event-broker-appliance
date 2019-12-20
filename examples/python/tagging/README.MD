### Get the example function

Clone this repository which contains the example functions.

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
git checkout development
cd vcenter-event-broker-appliance/examples/python/tagging
```

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

Take a note of the `urn:...` for `demotag1` as we will need it for the next steps.

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