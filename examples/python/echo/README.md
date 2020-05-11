# VMware Event Broker Appliance Echo Event Function

## Description

This function helps users understand the structure and data of a given vCenter Event which will be useful when creating brand new Functions.

Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/python/echo
git checkout master
```

Step 2 - Edit `stack.yml` and update the topic with the specific vCenter Server Event(s) from [vCenter Event Mapping](https://github.com/lamw/vcenter-event-mapping) document

Step 3 - Login to the OpenFaaS gateway on VMware Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY}

faas-cli login --username admin --password-stdin --tls-no-verify
```

Step 4 - Deploy function to VMware Event Broker Appliance

```
faas-cli deploy -f stack.yml --tls-no-verify
```

Step 5 - Tail the logs of the veba-echo function

```
faas-cli logs veba-echo --tls-no-verify
```

Step 6 - Trigger the vCenter Event such as powering off the VM for the VmPoweredOffEvent and you should see output like the following in the console:

```
2020-02-23T22:29:28Z 2020/02/23 22:29:28 Forking fprocess.
2020-02-23T22:29:28Z 2020/02/23 22:29:28 Query
2020-02-23T22:29:28Z 2020/02/23 22:29:28 Path  /
2020-02-23T22:29:28Z {"id":"6be1aa78-4e34-4697-87bd-fd189934804d","source":"https://vcenter.sddc-a-b-c-d.vmwarevmc.com/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmPoweredOffEvent","time":"2020-02-23T22:29:28.911840208Z","data":{"Key":303794,"ChainId":303792,"CreatedTime":"2020-02-23T22:29:28.226884Z","UserName":"VMC.LOCAL\\cloudadmin","Datacenter":{"Name":"SDDC-Datacenter","Datacenter":{"Type":"Datacenter","Value":"datacenter-3"}},"ComputeResource":{"Name":"Cluster-1","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c8"}},"Host":{"Name":"10.20.32.4","Host":{"Type":"HostSystem","Value":"host-11"}},"Vm":{"Name":"Test","Vm":{"Type":"VirtualMachine","Value":"vm-1081"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Test on  10.20.32.4 in SDDC-Datacenter is powered off","ChangeTag":"","Template":false},"datacontenttype":"application/json"}
2020-02-23T22:29:28Z 2020/02/23 22:29:28 Duration: 0.061631 seconds
```

For readability, you can copy the JSON:

```
{"id":"6be1aa78-4e34-4697-87bd-fd189934804d","source":"https://vcenter.sddc-a-b-c-d.vmwarevmc.com/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmPoweredOffEvent","time":"2020-02-23T22:29:28.911840208Z","data":{"Key":303794,"ChainId":303792,"CreatedTime":"2020-02-23T22:29:28.226884Z","UserName":"VMC.LOCAL\\cloudadmin","Datacenter":{"Name":"SDDC-Datacenter","Datacenter":{"Type":"Datacenter","Value":"datacenter-3"}},"ComputeResource":{"Name":"Cluster-1","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c8"}},"Host":{"Name":"10.20.32.4","Host":{"Type":"HostSystem","Value":"host-11"}},"Vm":{"Name":"Test","Vm":{"Type":"VirtualMachine","Value":"vm-1081"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Test on  10.20.32.4 in SDDC-Datacenter is powered off","ChangeTag":"","Template":false},"datacontenttype":"application/json"}
```

and format that using a JSON Linter website such as [https://jsonlint.com/](https://jsonlint.com/)