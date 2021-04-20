# vCenter Managed Object Pre-Filter

## Description
This function allows you to use regex filters to match against any cloud event data field, including the inventory paths of any managed object references received from vCenter. If the filters you define match, the function will trigger a definable [chained function](https://github.com/openfaas/workshop/blob/master/lab4.md#call-one-function-from-another). As such it can be used in front of any existing function to limit the scope of when that function is run - for example, if you only want to tag VMs within a specific folder or resource pool when they're powered on (rather than all VMs). 


## Get the example function
Clone this repository which contains the example functions. 

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/python/pre-filter
git checkout master
```


## Deploy a secondary function
The pre-filter function requires a secondary function to call, which must already be deployed to your VEBA appliance. This can be any valid function but __probably shouldn't__ have any topics defined in its `stack.yml` file. 
If you are just getting started, see the [veba-echo](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/examples/python/echo) function for an example. 


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
- Set `match_all` to `true` to ensure that __all__ your defined `filter_...` environment variables are matched. If this is set to `false` then any defined `filter_...` environment variables that reference a parameter that is not present in the cloud event are ignored.

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
      insecure_ssl: true # set to true to disable validation of vcenter ssl certificate
      match_all: false # require that all filters be positively matched to event data
      call_function: veba-echo # chained function to call
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
      insecure_ssl: true # set to true to disable validation of vcenter ssl certificate
      match_all: false # require that all filters be positively matched to event data
      call_function: veba-echo # chained function to call
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
      insecure_ssl: true # set to true to disable validation of vcenter ssl certificate
      match_all: false # require that all filters be positively matched to event data
      call_function: veba-echo # chained function to call
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
Filtering is performed using a regex search against the event data, or in the case of vCenter managed objects, the full inventory path to the object - for example, for a VM this could be `/Datacenter/vm/Folder/Test VM`. 

Each event received by VEBA will usually contain several references to vCenter managed objects - for a VmPoweredOnEvent for example you would get (at least) a VirtualMachine object, a HostSystem object and a Datacenter object.

Filters are defined as environment variables in the function's `stack.yml` file and must be named `filter_PARAM` where `PARAM` is the name of the parameter from the event data passed to the function. Parameter names can be specified using dot notation to traverse the event data structure. For example, take the following event data snippet:

```json
{
  "CreatedTime": "2020-07-02T15:16:11.207727Z",
  "UserName": "ro-user@vsphere.local",
  "Vm": {
    "Name": "Test VM",
    "Vm": {
      "Type": "VirtualMachine",
      "Value": "vm-82"
    }
  },
```

For example you could specifiy:
- `filter_username` to match against `ro-user@vsphere.local`
- `filter_vm.name` to match against `Test VM`
- `filter_vm` to match against the full inventory path to the vm object (e.g. `/My Datacenter/vm/Test VMs/Test VM`)

Filter values use regex notation and should be enclosed in single quotes. For example:

```yaml
# Object filter examples
filter_vm: '\/Test\/'  # Match any VM in a folder called "Test"
filter_host: 'esxi0[12]$'   # Match a host whose name ends with esxi01 or esxi02
```

The filtering is performed using the Python `re.search()` function - see the [official documentation](https://docs.python.org/3/library/re.html#regular-expression-syntax) for details on writing regex patterns.

Finally, if `write_debug` is set to true in `stack.yml` then all available event data parameters are logged in the function logs with lines beginning `Event Data > `

> **Note:** It is best practice to set `write_debug` to false once you have finished configuring your filters to avoid excessive logging and speed up function execution

### Example filters
- `filter_vm.name: '^prod-'` - Matches all VMs named with the 'prod-' prefix
- `filter_vm: '\/DC01\/vm/Servers\/' ` - Matches all VMs in the Servers VM folder of the "DC01" datacenter
- `filter_computeresource: 'Lab$'` - Matches events raised by objects within a cluster whose name ends with "Lab"


### Filtering on Event Data Arrays
Some cloud events contain data parameters that are an array. Take the `Arguments` parameter in the following `vim.event.ResourceExhaustionStatusChangedEvent` snippet for example:
```json
{
  "data": {
    "Arguments": [
      {
        "Key": "resourceName",
        "Value": "storage_util_filesystem_log"
      },
      {
        "Key": "oldStatus",
        "Value": "yellow"
      },
      {
        "Key": "newStatus",
        "Value": "green"
      },
      {
        "Key": "reason",
        "Value": " "
      },
      {
        "Key": "nodeType",
        "Value": "vcenter"
      },
      {
        "Key": "_sourcehost_",
        "Value": "vcsa.lab"
      }
    ],
```
There are two options to define filters for array parameters:
1. Define a `filter_...` environment variable using the numeric index. e.g. `filter_arguments.2.value = 'green'` would match element 2 in the above.
2. Replace one or more numeric indicies with a single `n`. This will cause the filter to essentially search all elements of the array. e.g. `filter_arguments.n.value = 'vcenter'` would match element 4 in the above.

## Faas Stack
Just before your chained function is called, an array called `faasstack` is inserted/updated as an extension attribute in the cloud event containing the name of your pre-filter function. This can then be used and/or appended to in the chained function as a way of tracking what called your function.

## Troubleshooting
If your chained function did not run, verify:

- vCenter IP/username/password
- Permissions of the vCenter user
- Whether the components can talk to each other (VMware Event Router to vCenter and OpenFaaS, function to vCenter)
- Check the logs (`kubectl` is installed and configured locally on the appliance)):

```bash
faas-cli logs pre-filter --follow --tls-no-verify 

# Successful log message in the OpenFaaS pre-filter function
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Validation passed! Applying object filters:
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > Key = 661858
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > ChainId = 661856
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > CreatedTime = 2020-10-19T20:37:39.445544Z
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > UserName = CORE\david
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > datacenter.name = Pilue
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > datacenter.datacenter.type = Datacenter
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > datacenter.datacenter.value = datacenter-2
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > computeresource.name = Lab
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > computeresource.computeresource.type = ClusterComputeResource
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > computeresource.computeresource.value = domain-c47
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > host.name = esxi01.lab.core.pilue.co.uk
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > host.host.type = HostSystem
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > host.host.value = host-36331
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > vm.name = gps unit
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > vm.vm.type = VirtualMachine
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > vm.vm.value = vm-23311
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > Ds = None
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > Net = None
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > Dvs = None
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > FullFormattedMessage = gps unit on esxi01.lab.core.pilue.co.uk in Pilue has powered on
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > ChangeTag =
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Event Data > Template = False
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Key > use
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Value not found for key 'use'
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Key > vm
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Found Managed object > vm-23311 has name gps unit and type VirtualMachine
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: VirtualMachine Path > /Pilue/vm/Testing/gps unit
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: Match > Filter ".*" matched "/Pilue/vm/Testing/gps unit"
2020-10-19T20:37:40Z 2020/10/19 20:37:40 stderr: All filters matched. Calling chained function veba-echo
2020-10-19T20:37:40Z 2020/10/19 20:37:40 POST / - 200 OK - ContentLength: 0
```

Filters that fail to match will cause the function to exit early. [Regexr](https://regexr.com/) is a good tool for testing and debugging regex searches. 

Any errors or issues will be logged and a 400 or 500 error code returned. A 404 error on calling the chained function is probably that the function defined in the `call_function` environment variable does not exist. Check your spelling and use `faas-cli list --tls-no-verify` to confirm the deployed functions in your appliance. 

If your chained function is running twice or is running when filters don't match, ensure that the chained function's `stack.yml` does __not__ define any event types in the `topic` parameter.
