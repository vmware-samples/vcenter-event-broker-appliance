# vCenter Managed Object Pre-Filter

## Description
This function allows you to use regex filters to match against the inventory paths of any managed object references received from vCenter. If all the filters you define match, the function will trigger a definable [chained function](https://github.com/openfaas/workshop/blob/master/lab4.md#call-one-function-from-another). As such it can be used in front of any existing function to limit the scope of when that function is run - for example, if you only want to tag VMs within a specific folder or resource pool when they're powered on (rather than all VMs). 


## Get the example function
Clone this repository which contains the example functions. 

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/python/pre-filter
git checkout master
```


## Deploy a main function
The pre-filter function requires a secondary function to call, which must already be deployed to your VEBA appliance. This can be any valid function but __probably shouldn't__ have any topics defined in its `stack.yml` file. 
If you are just getting started, see the [veba-echo](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/development/examples/python/echo) function for an example. 


## Customize the function
As there is a good chance you may want to deploy several pre-filter functions each with a slightly different config, most of the config is done in the `stack.yml` file. The vCenter secrets can be shared between all your functions. 


### Configure vCenter Connection 
For security reasons to not expose sensitive data we will create a Kubernetes [secret](https://kubernetes.io/docs/concepts/configuration/secret/) which will hold the vCenter credentials. This secret will be mounted into the function during runtime. This is all taken care of for you by the appliance. We only have to create the secret with a simple command through `faas-cli`.

First, change the configuration file [vcconfig.toml](vcconfig.toml) holding your secret vCenter information located in this folder:

```toml
# vcconfig.toml contents
# replace with your own values and use a dedicated user/service account. Read Only permissions are all that is required
[vcenter]
server = "VCENTER_FQDN/IP"
user = "ro-user@vsphere.local"
pass = "DontUseThisPassword"
```

Now go ahead and store this configuration file as a secret in the appliance.

```bash
# set up faas-cli for first use
export OPENFAAS_URL=https://VEBA_FQDN_OR_IP
faas-cli login -p VEBA_OPENFAAS_PASSWORD --tls-no-verify # vCenter Event Broker Appliance is configured with authentication, pass in the password used during the vCenter Event Broker Appliance deployment process

# now create the secret
faas-cli secret create vcconfig --from-file=vcconfig.toml --tls-no-verify
```

> **Note:** Delete the local `vcconfig.toml` after you're done with this exercise to not expose this sensitive information.

### Configure the function 
Next, configure the function by defining your filter rules, the vCenter event which will trigger this function, and the function to call. Such function-specific settings are performed in the `stack.yml` file. Open and edit the `stack.yml` provided within the Python pre-filter example code. 
- Change `gateway` and `topic` as per your environment/needs.
- Set the `call_function` environment variable to the name of the function you want this pre-filter function to call.
- Define one or more `filter_... ` environment variables using a regex expression (see [below](#defining-filters) for details and examples).

> **Note:** A key-value annotation under `topic` defines which VM event should trigger the function. A list of VM events from vCenter can be found [here](https://code.vmware.com/doc/preview?id=4206#/doc/vim.event.VmEvent.html). Multiple topics can be specified using a `","` delimiter syntax, e.g. "`topic: "VmPoweredOnEvent,VmPoweredOffEvent"`".

```yaml
provider:
  name: openfaas
  gateway: https://VEBA_FQDN_OR_IP # replace with your vCenter Event Broker Appliance environment
functions:
  pre-filter:
    lang: python3-flask
    handler: ./pre-filter
    image: vmware/veba-python-pre-filter:latest
    environment:
      write_debug: true
      read_debug: true
      insecure_ssl: true
      call_function: veba-echo
      filter_vm: '.*'
    secrets:
      - vcconfig
    annotations:
      topic: VmPoweredOnEvent # or DrsVmPoweredOnEvent in a DRS-enabled cluster
```

> **Note:** If you are running a vSphere DRS-enabled cluster the topic annotation above should be `DrsVmPoweredOnEvent`. Otherwise the function would never be triggered.

If you wish to define multiple pre-filter functions, you can declare them all in the same `stack.yml` as multiple objects under the `functions:` section:

```yaml
provider:
  name: openfaas
  gateway: https://VEBA_FQDN_OR_IP # replace with your vCenter Event Broker Appliance environment
functions:
  pre-filter-poweron:
    lang: python3-flask
    handler: ./pre-filter
    image: vmware/veba-python-pre-filter:latest
    environment:
      write_debug: true
      read_debug: true
      insecure_ssl: true
      call_function: veba-echo
      filter_vm: '\/Test VM 1$'
    secrets:
      - vcconfig
    annotations:
      topic: VmPoweredOnEvent # or DrsVmPoweredOnEvent in a DRS-enabled cluster
  pre-filter-poweroff:
    lang: python3-flask
    handler: ./pre-filter
    image: vmware/veba-python-pre-filter:latest
    environment:
      write_debug: true
      read_debug: true
      insecure_ssl: true
      call_function: veba-echo
      filter_vm: '\/Test VM 2$'
    secrets:
      - vcconfig
    annotations:
      topic: VmPoweredOffEvent
```
> **Note** If you want to test filters without actually calling a chained function you can simply leave the `call_function` environment variable undefined. The pre-filter function will still run and the logs will detail all the filters performed. 

## Deploy the function
After you've performed the steps and modifications above, you can go ahead and deploy the function:

```bash
faas-cli template store pull python3-flask # only required during the first deployment
faas-cli deploy -f stack.yml --tls-no-verify
Deployed. 202 Accepted.
```


## Trigger the function
Turn on a virtual machine to trigger the function via a `(DRS)VmPoweredOnEvent`. Review the logs of your pre-filter function to determine if your filters are working as expected, and the logs of your chained function to ensure it is being correctly called. 

```bash
faas-cli logs pre-filter --tls-no-verify

... 
2020-08-02T00:08:11Z 2020/08/02 00:08:11 stderr: All filters matched. Calling chained function veba-echo
2020-08-02T00:08:13Z 2020/08/02 00:08:13 POST / - 200 OK - ContentLength: 884
```

> **Note:** If you don't see the above two log entries, verify that you correctly followed each step above, IPs/FQDNs and credentials are correct and see the [troubleshooting](#troubleshooting) section below.


## Defining Filters
vCenter Object filtering is performed using a regex search against the full inventory path to the object - for example, for a VM this could be `/Datacenter/vm/Folder/Test VM`. 

Each event received by VEBA will usually contain several references to vCenter objects - for a VmPoweredOnEvent for example you would get (at least) a VirtualMachine object, a HostSystem object and a Datacenter object.

Object Filters are defined as environment variables in the function's `stack.yml` file and must be named `filter_OBJECT` where OBJECT is the parameter name of the object from the event data passed to the function. For example `filter_vm` for VirtualMachine objects, or `filter_host` for HostSystem objects.

```yaml
# Object filter examples
filter_vm: '\/Test VM 2$'
filter_host: 'esxi[12]$'
```

The filtering is performed using the Python `re.search()` function - see the [official documentation](https://docs.python.org/3/library/re.html#regular-expression-syntax) for details on writing regex patterns. 

To find out which vCenter objects are available to match against, monitor the function logs and look for lines starting `Apply Filter >`

```bash
Apply Filter > "datacenter" object (Datacenter): "filter_datacenter" = .*
...
Apply Filter > "computeresource" object (ClusterComputeResource): "filter_computeresource" = .*
...
Apply Filter > "host" object (HostSystem): "filter_host" = .*
...
Apply Filter > "vm" object (VirtualMachine): "filter_vm" = .*

```
Each `Apply Filter >` line details a vCenter object found in the event data and its matching `filter_...` environment variable name and current value. 

> **Note:** If any specific `filter_... ` environment variable is not defined it defaults to `.*` - i.e. match anything


### Example filters
- `filter_vm: '\/DC01\/vm/Servers\/' ` - Matches all VMs in the Servers VM folder of the "DC01" datacenter
- `filter_computeresource: 'Lab$'` - Matches events raised by objects within a cluster called "Lab"


## Troubleshooting
If your chained function did not run, verify:

- vCenter IP/username/password
- Permissions of the vCenter user
- Whether the components can talk to each other (VMware Event Router to vCenter and OpenFaaS, function to vCenter)
- Check the logs (`kubectl` is installed and configured locally on the appliance)):

```bash
faas-cli logs pre-filter --follow --tls-no-verify 

# Successful log message in the OpenFaaS pre-filter function
2020/08/02 20:44:13 stderr: Validation passed! Applying object filters:
2020/08/02 20:44:13 stderr: Managed object > datacenter-2 has name Pilue and type Datacenter
2020/08/02 20:44:13 stderr: Apply Filter > "datacenter" object (Datacenter): "filter_datacenter" = .*
2020/08/02 20:44:13 stderr: Datacenter Path > /Pilue
2020/08/02 20:44:13 stderr: Match > Filter matched Datacenter path
2020/08/02 20:44:13 stderr: Managed object > domain-c47 has name Lab and type ClusterComputeResource
2020/08/02 20:44:13 stderr: Apply Filter > "computeresource" object (ClusterComputeResource): "filter_computeresource" = .*
2020/08/02 20:44:13 stderr: ClusterComputeResource Path > /Pilue/host/Lab
2020/08/02 20:44:13 stderr: Match > Filter matched ClusterComputeResource path
2020/08/02 20:44:13 stderr: Managed object > host-3605 has name esxi01.lab.core.pilue.co.uk and type HostSystem
2020/08/02 20:44:13 stderr: Apply Filter > "host" object (HostSystem): "filter_host" = .*
2020/08/02 20:44:13 stderr: HostSystem Path > /Pilue/host/Lab/esxi01.lab.core.pilue.co.uk
2020/08/02 20:44:13 stderr: Match > Filter matched HostSystem path
2020/08/02 20:44:13 stderr: Managed object > vm-82 has name sexigraf and type VirtualMachine
2020/08/02 20:44:13 stderr: Apply Filter > "vm" object (VirtualMachine): "filter_vm" = .*

2020/08/02 20:44:13 stderr: VirtualMachine Path > /Pilue/vm/Infrastructure/Other/sexigraf
2020/08/02 20:44:13 stderr: Match > Filter matched VirtualMachine path
2020/08/02 20:44:13 stderr: All filters matched. Calling chained function veba-echo
2020/08/02 20:44:13 POST / - 200 OK - ContentLength: 884
```

Filters that fail to match will cause the function to exit early. [Regexr](https://regexr.com/) is a good tool for testing and debugging regex searches. 

Any errors or issues will be logged and a 400 or 500 error code returned. A 404 error on calling the chained function is probably that the function defined in the `call_function` environment variable does not exist. Check your spelling and use `faas-cli list --tls-no-verify` to confirm the deployed functions in your appliance. 

If your chained function is running twice or is running when filters don't match, ensure that the chained function's `stack.yml` does __not__ define any event types in the `topic` parameter.
