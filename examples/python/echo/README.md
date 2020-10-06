# VMware Event Broker Appliance Echo Event Function

## Description

This function helps users understand the structure and data of a given vCenter
Event which will be useful when creating brand new Functions. 

Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/python/echo
git checkout master
```

Step 2 - Edit `stack.yml` and update the topic with the specific vCenter Server
Event(s) from [vCenter Event
Mapping](https://github.com/lamw/vcenter-event-mapping) document (by default the
function will be triggered on VM power on/off)

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

Step 6 - Trigger the vCenter Event such as powering on the VM for the
(Drs)VmPoweredOnEvent and you should see output like the following in the console:

```
Forking - python [index.py]
2020/10/05 18:55:16 Started logging stderr from function.
2020/10/05 18:55:16 Started logging stdout from function.
2020/10/05 18:55:16 OperationalMode: http
2020/10/05 18:55:16 Timeouts: read: 10s, write: 10s hard: 10s.
2020/10/05 18:55:16 Listening on port: 8080
2020/10/05 18:55:16 Writing lock-file to: /tmp/.lock
2020/10/05 18:55:16 Metrics listening on port: 8081
2020/10/05 19:40:49 stdout: Serving on http://0.0.0.0:5000
2020/10/05 19:40:49 stdout: b'{"data":{"Key":18968,"ChainId":18963,"CreatedTime":"2020-10-05T19:40:49.134Z","UserName":"VSPHERE.LOCAL\\\\Administrator","Datacenter":{"Name":"vcqaDC","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":{"Name":"cls","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c7"}},"Host":{"Name":"10.0.0.119","Host":{"Type":"HostSystem","Value":"host-21"}},"Vm":{"Name":"test-01","Vm":{"Type":"VirtualMachine","Value":"vm-57"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"DRS powered On test-01 on 10.0.0.119 in vcqaDC","ChangeTag":"","Template":false},"datacontenttype":"application/json","id":"1f2a18c6-c788-4a9a-a2fb-8be0ec34330f","source":"https://10.0.0.28/sdk","specversion":"1.0","subject":"DrsVmPoweredOnEvent","time":"2020-10-05T19:40:49.508328Z","type":"com.vmware.event.router/event"}'
2020/10/05 19:40:49 POST / - 200 OK - ContentLength: 0
```

For readability, you can copy the JSON data (everything in `{}` after `b'`) and format that using a JSON Linter website such as
[https://jsonlint.com/](https://jsonlint.com/) or a program like `jq`:

```
$ pbpaste| jq .
{
  "data": {
    "Key": 18968,
    "ChainId": 18963,
    "CreatedTime": "2020-10-05T19:40:49.134Z",
    "UserName": "VSPHERE.LOCAL\\\\Administrator",
    "Datacenter": {
      "Name": "vcqaDC",
      "Datacenter": {
        "Type": "Datacenter",
        "Value": "datacenter-2"
      }
    },
    "ComputeResource": {
      "Name": "cls",
      "ComputeResource": {
        "Type": "ClusterComputeResource",
        "Value": "domain-c7"
      }
    },
    "Host": {
      "Name": "10.0.0.119",
      "Host": {
        "Type": "HostSystem",
        "Value": "host-21"
      }
    },
    "Vm": {
      "Name": "test-01",
      "Vm": {
        "Type": "VirtualMachine",
        "Value": "vm-57"
      }
    },
    "Ds": null,
    "Net": null,
    "Dvs": null,
    "FullFormattedMessage": "DRS powered On test-01 on 10.0.0.119 in vcqaDC",
    "ChangeTag": "",
    "Template": false
  },
  "datacontenttype": "application/json",
  "id": "1f2a18c6-c788-4a9a-a2fb-8be0ec34330f",
  "source": "https://10.0.0.28/sdk",
  "specversion": "1.0",
  "subject": "DrsVmPoweredOnEvent",
  "time": "2020-10-05T19:40:49.508328Z",
  "type": "com.vmware.event.router/event"
}
```

## Build from scratch

```bash
# pull template
faas template store pull python3-http

# modify stack as per your needs, e.g. image name, then
faas build -f stack.yml
faas push -f stack.yml
```