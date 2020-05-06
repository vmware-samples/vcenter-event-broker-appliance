import sys, os
from shutil import copyfile
from handler import handler

#
## Unit Test - helps testing the function locally
## Uncomment Config - update the path to the file accordingly
## Uncomment handle('...') to test the function with the event samples provided below test without deploying to OpenFaaS
#
CONFIG = "sfconfig.json"

#
## TEST CASE 
#
testlist = [
    '', 
    '"test":"ok"',
    '{"test":"ok"}',
    '{"data":"ok"}',
    '{"id":"17e1027a-c865-4354-9c21-e8da3df4bff9","source":"https://vcsa.pdotk.local/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"UserLogoutSessionEvent","time":"2020-04-14T00:28:36.455112549Z","data":{"Key":7775,"ChainId":7775,"CreatedTime":"2020-04-14T00:28:35.221698Z","UserName":"machine-b8eb9a7f","Datacenter":null,"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"User machine-b8ebe7eb9a7f@127.0.0.1 logged out (login time: Tuesday, 14 April, 2020 12:28:35 AM, number of API invocations: 34, user agent: pyvmomi Python/3.7.5 (Linux; 4.19.84-1.ph3; x86_64))","ChangeTag":"","IpAddress":"127.0.0.1","UserAgent":"pyvmomi Python/3.7.5 (Linux; 4.19.84-1.ph3; x86_64)","CallCount":34,"SessionId":"52edf160927","LoginTime":"2020-04-14T00:28:35.071817Z"},"datacontenttype":"application/json"}',
    '{"id":"0707d7e0-269f-42e7-ae1c-18458ecabf3d","source":"https://vcsa.pdotk.local/sdk","specversion":"1.0","type":"com.vmware.event.router/eventex","subject":"vim.event.ResourceExhaustionStatusChangedEvent","time":"2020-04-14T00:20:15.100325334Z","data":{"Key":7715,"ChainId":7715,"CreatedTime":"2020-04-14T00:20:13.76967Z","UserName":"machine-bb9a7f","Datacenter":null,"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"vCenter Log File System Resource status changed from Yellow to Green on vcsa.pdotk.local  ","ChangeTag":"","EventTypeId":"vim.event.ResourceExhaustionStatusChangedEvent","Severity":"info","Message":"","Arguments":[{"Key":"resourceName","Value":"storage_util_filesystem_log"},{"Key":"oldStatus","Value":"yellow"},{"Key":"newStatus","Value":"green"},{"Key":"reason","Value":" "},{"Key":"nodeType","Value":"vcenter"},{"Key":"_sourcehost_","Value":"vcsa.pdotk.local"}],"ObjectId":"","ObjectType":"","ObjectName":"","Fault":null},"datacontenttype":"application/json"}',
    '{"id":"453120cd-3d19-4c43-aadc-df0cdbce3887","source":"https://vcsa.pdotk.local/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmPoweredOnEvent","time":"2020-04-13T23:46:10.402531287Z","data":{"Key":7441,"ChainId":7438,"CreatedTime":"2020-04-13T23:46:09.387283Z","UserName":"Administrator","Datacenter":{"Name":"PKLAB","Datacenter":{"Type":"Datacenter","Value":"datacenter-3"}},"ComputeResource":{"Name":"esxi01.pdotk.local","ComputeResource":{"Type":"ComputeResource","Value":"domain-s29"}},"Host":{"Name":"esxi01.pdotk.local","Host":{"Type":"HostSystem","Value":"host-31"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-33"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Test VM on esxi01.pdotk.local in PKLAB has powered on","ChangeTag":"","Template":false},"datacontenttype":"application/json"}',
    '{"id":"d77a3767-1727-49a3-ac33-ddbdef294150","source":"https://vcsa.pdotk.local/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmPoweredOffEvent","time":"2020-04-14T00:33:30.838669841Z","data":{"Key":7825,"ChainId":7821,"CreatedTime":"2020-04-14T00:33:30.252792Z","UserName":"Administrator","Datacenter":{"Name":"PKLAB","Datacenter":{"Type":"Datacenter","Value":"datacenter-3"}},"ComputeResource":{"Name":"esxi01.pdotk.local","ComputeResource":{"Type":"ComputeResource","Value":"domain-s29"}},"Host":{"Name":"esxi01.pdotk.local","Host":{"Type":"HostSystem","Value":"host-31"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-33"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Test VM on  esxi01.pdotk.local in PKLAB is powered off","ChangeTag":"","Template":false},"datacontenttype":"application/json"}'
]

if __name__ == "__main__":
    #COPYING FILE to where the function expects the secrets/config
    #make sure to have a folder created with write permissions
    src = CONFIG
    dst= "/var/openfaas/secrets/sfconfig"
    copyfile(src, dst)
  
    #Testing all test cases
    #ensure success case is at the end
    for test in testlist:
        print(f"BEGIN TEST for input:\n'{test}'")
        print(f"TEST RESULTS:\n")
        ret = handler.handle(test)
        if ret != None:
            print(f"RESPONSE:\n")
            print(ret)
        print(f"END TEST")
        print(f"----------------")