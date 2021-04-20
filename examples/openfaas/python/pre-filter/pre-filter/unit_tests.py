import handler
import os
import sys

r = []

handler.VC_CONFIG = 'vcconfig.toml'
handler.DEBUG = False
handler.INSECURE_SSL = True

#
# Previous TEST CASES :
#
# Invalid vc objects
# r = handle(r'{"id":"c7a6c420-f25d-4e6d-95b5-e273202e1164","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DrsVmPoweredOnEvent","time":"2020-07-02T15:16:13.533866543Z","data":{"Key":130278,"ChainId":130273,"CreatedTime":"2020-07-02T15:16:11.213467Z","UserName":"Administrator","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":{"Name":"Lab","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c47"}},"Host":{"Name":"esxi03.lab","Host":{"Type":"HostSystem","Value":"host-9999"}},"Vm":{"Name":"Bad VM","Vm":{"Type":"VirtualMachine","Value":"vm-9999"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"DRS powered on Bad VM on esxi01.lab in Lab","ChangeTag":"","Template":false},"datacontenttype":"application/json"}')
# Standard : UserLogoutSessionEvent
# r'{"id":"17e1027a-c865-4354-9c21-e8da3df4bff9","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"UserLogoutSessionEvent","time":"2020-04-14T00:28:36.455112549Z","data":{"Key":7775,"ChainId":7775,"CreatedTime":"2020-04-14T00:28:35.221698Z","UserName":"machine-b8eb9a7f","Datacenter":null,"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"User machine-b8ebe7eb9a7f@127.0.0.1 logged out (login time: Tuesday, 14 April, 2020 12:28:35 AM, number of API invocations: 34, user agent: pyvmomi Python/3.7.5 (Linux; 4.19.84-1.ph3; x86_64))","ChangeTag":"","IpAddress":"127.0.0.1","UserAgent":"pyvmomi Python/3.7.5 (Linux; 4.19.84-1.ph3; x86_64)","CallCount":34,"SessionId":"52edf160927","LoginTime":"2020-04-14T00:28:35.071817Z"},"datacontenttype":"application/json"}'
# Eventex : vim.event.ResourceExhaustionStatusChangedEvent
# r'{"id":"0707d7e0-269f-42e7-ae1c-18458ecabf3d","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/eventex","subject":"vim.event.ResourceExhaustionStatusChangedEvent","time":"2020-04-14T00:20:15.100325334Z","data":{"Key":7715,"ChainId":7715,"CreatedTime":"2020-04-14T00:20:13.76967Z","UserName":"machine-bb9a7f","Datacenter":null,"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"vCenter Log File System Resource status changed from Yellow to Green on vcsa.lab ","ChangeTag":"","EventTypeId":"vim.event.ResourceExhaustionStatusChangedEvent","Severity":"info","Message":"","Arguments":[{"Key":"resourceName","Value":"storage_util_filesystem_log"},{"Key":"oldStatus","Value":"yellow"},{"Key":"newStatus","Value":"green"},{"Key":"reason","Value":" "},{"Key":"nodeType","Value":"vcenter"},{"Key":"_sourcehost_","Value":"vcsa.lab"}],"ObjectId":"","ObjectType":"","ObjectName":"","Fault":null},"datacontenttype":"application/json"}'
# Standard : DrsVmPoweredOnEvent
# r'{"id":"c7a6c420-f25d-4e6d-95b5-e273202e1164","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DrsVmPoweredOnEvent","time":"2020-07-02T15:16:13.533866543Z","data":{"Key":130278,"ChainId":130273,"CreatedTime":"2020-07-02T15:16:11.213467Z","UserName":"Administrator","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":{"Name":"Lab","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c47"}},"Host":{"Name":"esxi03.lab","Host":{"Type":"HostSystem","Value":"host-3523"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-82"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"DRS powered on Test VM on esxi03.lab in Lab","ChangeTag":"","Template":false},"datacontenttype":"application/json"}'
# Standard : VmPoweredOffEvent
# r'{"id":"d77a3767-1727-49a3-ac33-ddbdef294150","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmPoweredOffEvent","time":"2020-04-14T00:33:30.838669841Z","data":{"Key":7825,"ChainId":7821,"CreatedTime":"2020-04-14T00:33:30.252792Z","UserName":"Administrator","Datacenter":{"Name":"PKLAB","Datacenter":{"Type":"Datacenter","Value":"datacenter-3"}},"ComputeResource":{"Name":"esxi01.lab","ComputeResource":{"Type":"ComputeResource","Value":"domain-s29"}},"Host":{"Name":"esxi01.lab","Host":{"Type":"HostSystem","Value":"host-31"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-33"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Test VM on  esxi01.lab in PKLAB is powered off","ChangeTag":"","Template":false},"datacontenttype":"application/json"}'
# Standard : DvsPortLinkUpEvent
# r'{"id":"a10f8571-fc2a-40db-8df6-8284cecf5720","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DvsPortLinkUpEvent","time":"2020-07-02T15:16:13.43892986Z","data":{"Key":130277,"ChainId":130277,"CreatedTime":"2020-07-02T15:16:11.207727Z","UserName":"","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":{"Name":"Lab Switch","Dvs":{"Type":"VmwareDistributedVirtualSwitch","Value":"dvs-22"}},"FullFormattedMessage":"The dvPort 2 link was up in the vSphere Distributed Switch Lab Switch in Lab","ChangeTag":"","PortKey":"2","RuntimeInfo":null},"datacontenttype":"application/json"}'
# Standard : DatastoreRenamedEvent
# r'{"id":"369b403a-6729-4b0b-893e-01383c8307ba","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DatastoreRenamedEvent","time":"2020-07-02T21:44:11.09338265Z","data":{"Key":130669,"ChainId":130669,"CreatedTime":"2020-07-02T21:44:08.578289Z","UserName":"","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Renamed datastore from esxi04-local to esxi04-localZ in Lab","ChangeTag":"","Datastore":{"Name":"esxi04-localZ","Datastore":{"Type":"Datastore","Value":"datastore-3313"}},"OldName":"esxi04-local","NewName":"esxi04-localZ"},"datacontenttype":"application/json"}'
# Standard : DVPortgroupRenamedEvent
# r'{"id":"aab77fd1-41ed-4b51-89d3-ef3924b09de1","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DVPortgroupRenamedEvent","time":"2020-07-03T19:36:38.474640186Z","data":{"Key":132376,"ChainId":132375,"CreatedTime":"2020-07-03T19:36:32.525906Z","UserName":"Administrator","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":{"Name":"vMotion AZ","Network":{"Type":"DistributedVirtualPortgroup","Value":"dvportgroup-3357"}},"Dvs":{"Name":"10G Switch A","Dvs":{"Type":"VmwareDistributedVirtualSwitch","Value":"dvs-3355"}},"FullFormattedMessage":"dvPort group vMotion A in Lab was renamed to vMotion AZ","ChangeTag":"","OldName":"vMotion A","NewName":"vMotion AZ"},"datacontenttype":"application/json"}'
# Standard : VmReconfiguredEvent
# r'{"id":"1fa118f5-c10e-4cd6-bff8-9f569adf19ad","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmReconfiguredEvent","time":"2020-08-25T16:43:38.891590532Z","data":{"Key":23765556,"ChainId":23765534,"CreatedTime":"2020-08-25T16:43:38.085242Z","UserName":"Administrator","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-21"}},"ComputeResource":{"Name":"Lab","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c26"}},"Host":{"Name":"esxi01.lab","Host":{"Type":"HostSystem","Value":"host-409"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-760"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Reconfigured Test VM on esxi01.lab in Lab.\n \nModified: \n \nconfig.hardware.device(2000).backing.fileName:\"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM.vmdk\"-\u003e \"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM_1.vmdk\"; \n\nconfig.hardware.device(2000).backing.contentId:\"29699e346b1037fb6d66817311f0def3\"-\u003e \"5a459ada07ea71fed8f3ba28178c333b\"; \n\nconfig.hardware.device(2000).backing.keyId:\u003cunset\u003e -\u003e (keyId = \"a4e7494f-da55-4917-9d90-560c5078866b\",providerId = (id = \"KMS\")); \n\nconfig.hardware.device(2000).iofilter:() -\u003e (\"VMW_vmwarevmcrypt_1.0.0\"); \n\nconfig.extraConfig(\"nvram\").value:\"Test VM.nvram\"-\u003e \"Test VM_1.nvram\"; \n\nconfig.keyId:\u003cunset\u003e -\u003e (keyId = \"a4e7494f-da55-4917-9d90-560c5078866b\",providerId = (id = \"KMS\")); \n\n Added: \n \nconfig.extraConfig(\"encryption.bundle\"):(key = \"encryption.bundle\",value = \"vmware:key/list/(pair/(fqid/\u003cVMWARE-NULL\u003e/KMS/a4e7494f-da55-4917-9d90-560c5078866b,HMAC-SHA-256,KPYD5DzgqxZVihEqnejnZEQfltXEmEu3w1sGVvoqtHJzbgDNqi/IMD2BehT/uykcKXaCaYeCDMIo0KswzfP3Gz1wXN3f5gyKgMhPBqGg/LDM4Ee5cqQ/bR8MbLromllqsD028Akd/9q/rfn3RAkMlBsQW0H8FKYwbYriKTIKr58dL5GhlCoiZdapRXsyJRvDFdMFdUpkujYUhbMyc2lQAuR35YsXN8WT+ynYWuEeQ9rraEndXRJg2/wLztvziK8bWNDLTe8h+3HN+cg/2aNHDfZCz9/nrpJrukSy6lG8Xa05MkVk4/tZatnwQXHGafhBg7oT6LVGQfl6muVSpxFhAaFpa0ZcHwadmhBeZDcrb4VeT/aC0TPn2Pfr3MyXpyb8))\"); \n\n Deleted: \n \n","ChangeTag":"","Template":false,"ConfigSpec":{"ChangeVersion":"","Name":"","Version":"","CreateDate":"2019-09-12T14:20:06.666314Z","Uuid":"","InstanceUuid":"","NpivNodeWorldWideName":null,"NpivPortWorldWideName":null,"NpivWorldWideNameType":"","NpivDesiredNodeWwns":0,"NpivDesiredPortWwns":0,"NpivTemporaryDisabled":null,"NpivOnNonRdmDisks":null,"NpivWorldWideNameOp":"","LocationId":"","GuestId":"","AlternateGuestName":"","Annotation":"","Files":{"VmPathName":"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM.vmx","SnapshotDirectory":"","SuspendDirectory":"","LogDirectory":"","FtMetadataDirectory":""},"Tools":null,"Flags":null,"ConsolePreferences":null,"PowerOpInfo":null,"NumCPUs":0,"NumCoresPerSocket":0,"MemoryMB":0,"MemoryHotAddEnabled":null,"CpuHotAddEnabled":null,"CpuHotRemoveEnabled":null,"VirtualICH7MPresent":null,"VirtualSMCPresent":null,"DeviceChange":[{"Operation":"edit","FileOperation":"","Device":{"Key":2000,"DeviceInfo":{"Label":"Hard disk 1","Summary":"83,886,080 KB"},"Backing":{"FileName":"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM.vmdk","Datastore":{"Type":"Datastore","Value":"datastore-822"},"BackingObjectId":"","DiskMode":"persistent","Split":false,"WriteThrough":false,"ThinProvisioned":false,"EagerlyScrub":true,"Uuid":"6000C291-35bd-fd2d-e77e-709fb2c53635","ContentId":"29699e346b1037fb6d66817311f0def3","ChangeId":"","Parent":null,"DeltaDiskFormat":"","DigestEnabled":false,"DeltaGrainSize":0,"DeltaDiskFormatVariant":"","Sharing":"sharingNone","KeyId":null},"Connectable":null,"SlotInfo":null,"ControllerKey":1000,"UnitNumber":0,"CapacityInKB":83886080,"CapacityInBytes":85899345920,"Shares":{"Shares":1000,"Level":"normal"},"StorageIOAllocation":{"Limit":-1,"Shares":{"Shares":1000,"Level":"normal"},"Reservation":0},"DiskObjectId":"151-2000","VFlashCacheConfigInfo":null,"Iofilter":null,"VDiskId":null,"NativeUnmanagedLinkedClone":false},"Profile":[{"ProfileId":"b08ff38b-b1d6-413a-9fc1-20aceca2b31c","ReplicationSpec":null,"ProfileData":{"ExtensionKey":"com.vmware.vim.sps","ObjectData":"\u003cns1:storageProfile xmlns:ns1=\"http://profile.policy.data.vasa.vim.vmware.com/xsd\"xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"xsi:type=\"ns1:StorageProfile\"\u003e\u003cns1:constraints\u003e\u003cns1:subProfiles\u003e\u003cns1:capability\u003e\u003ccapabilityId xmlns=\"http://capability.policy.data.vasa.vim.vmware.com/xsd\"\u003e\u003cid\u003evmwarevmcrypt@ENCRYPTION\u003c/id\u003e\u003cnamespace\u003eIOFILTERS\u003c/namespace\u003e\u003c/capabilityId\u003e\u003cconstraint xmlns=\"http://capability.policy.data.vasa.vim.vmware.com/xsd\"\u003e\u003cpropertyInstance\u003e\u003cid\u003eAllowCleartextFilters\u003c/id\u003e\u003cvalue xmlns:s2=\"http://www.w3.org/2001/XMLSchema\"xsi:type=\"s2:string\"\u003eFalse\u003c/value\u003e\u003c/propertyInstance\u003e\u003c/constraint\u003e\u003c/ns1:capability\u003e\u003cns1:name\u003eTag based placement\u003c/ns1:name\u003e\u003c/ns1:subProfiles\u003e\u003c/ns1:constraints\u003e\u003cns1:createdBy\u003eTemporary user handle\u003c/ns1:createdBy\u003e\u003cns1:creationTime\u003e2019-10-25T10:49:35.537+01:00\u003c/ns1:creationTime\u003e\u003cns1:description\u003e\u003c/ns1:description\u003e\u003cns1:generationId\u003e3\u003c/ns1:generationId\u003e\u003cns1:lastUpdatedBy\u003eTemporary user handle\u003c/ns1:lastUpdatedBy\u003e\u003cns1:lastUpdatedTime\u003e2019-11-18T15:20:24.521+00:00\u003c/ns1:lastUpdatedTime\u003e\u003cns1:name\u003eStaging 3PAR Storage with Encryption\u003c/ns1:name\u003e\u003cns1:profileId\u003eb08ff38b-b1d6-413a-9fc1-20aceca2b31c\u003c/ns1:profileId\u003e\u003c/ns1:storageProfile\u003e"},"ProfileParams":null}],"Backing":{"Parent":null,"Crypto":{"CryptoKeyId":{"KeyId":"a4e7494f-da55-4917-9d90-560c5078866b","ProviderId":{"Id":"KMS"}}}}}],"CpuAllocation":null,"MemoryAllocation":null,"LatencySensitivity":null,"CpuAffinity":null,"MemoryAffinity":null,"NetworkShaper":null,"CpuFeatureMask":null,"ExtraConfig":null,"SwapPlacement":"","BootOptions":null,"VAppConfig":null,"FtInfo":null,"RepConfig":null,"VAppConfigRemoved":null,"VAssertsEnabled":null,"ChangeTrackingEnabled":null,"Firmware":"","MaxMksConnections":0,"GuestAutoLockEnabled":null,"ManagedBy":null,"MemoryReservationLockedToMax":null,"NestedHVEnabled":null,"VPMCEnabled":null,"ScheduledHardwareUpgradeInfo":null,"VmProfile":[{"ProfileId":"b08ff38b-b1d6-413a-9fc1-20aceca2b31c","ReplicationSpec":null,"ProfileData":{"ExtensionKey":"com.vmware.vim.sps","ObjectData":"\u003cns1:storageProfile xmlns:ns1=\"http://profile.policy.data.vasa.vim.vmware.com/xsd\"xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"xsi:type=\"ns1:StorageProfile\"\u003e\u003cns1:constraints\u003e\u003cns1:subProfiles\u003e\u003cns1:capability\u003e\u003ccapabilityId xmlns=\"http://capability.policy.data.vasa.vim.vmware.com/xsd\"\u003e\u003cid\u003evmwarevmcrypt@ENCRYPTION\u003c/id\u003e\u003cnamespace\u003eIOFILTERS\u003c/namespace\u003e\u003c/capabilityId\u003e\u003cconstraint xmlns=\"http://capability.policy.data.vasa.vim.vmware.com/xsd\"\u003e\u003cpropertyInstance\u003e\u003cid\u003eAllowCleartextFilters\u003c/id\u003e\u003cvalue xmlns:s1=\"http://www.w3.org/2001/XMLSchema\"xsi:type=\"s1:string\"\u003eFalse\u003c/value\u003e\u003c/propertyInstance\u003e\u003c/constraint\u003e\u003c/ns1:capability\u003e\u003cns1:name\u003eTag based placement\u003c/ns1:name\u003e\u003c/ns1:subProfiles\u003e\u003c/ns1:constraints\u003e\u003cns1:createdBy\u003eTemporary user handle\u003c/ns1:createdBy\u003e\u003cns1:creationTime\u003e2019-10-25T10:49:35.537+01:00\u003c/ns1:creationTime\u003e\u003cns1:description\u003e\u003c/ns1:description\u003e\u003cns1:generationId\u003e3\u003c/ns1:generationId\u003e\u003cns1:lastUpdatedBy\u003eTemporary user handle\u003c/ns1:lastUpdatedBy\u003e\u003cns1:lastUpdatedTime\u003e2019-11-18T15:20:24.521+00:00\u003c/ns1:lastUpdatedTime\u003e\u003cns1:name\u003eStaging 3PAR Storage with Encryption\u003c/ns1:name\u003e\u003cns1:profileId\u003eb08ff38b-b1d6-413a-9fc1-20aceca2b31c\u003c/ns1:profileId\u003e\u003c/ns1:storageProfile\u003e"},"ProfileParams":null}],"MessageBusTunnelEnabled":null,"Crypto":{"CryptoKeyId":{"KeyId":"a4e7494f-da55-4917-9d90-560c5078866b","ProviderId":{"Id":"KMS"}}},"MigrateEncryption":""},"ConfigChanges":{"Modified":"config.hardware.device(2000).backing.fileName:\"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM.vmdk\"-\u003e \"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM_1.vmdk\"; \n\nconfig.hardware.device(2000).backing.contentId:\"29699e346b1037fb6d66817311f0def3\"-\u003e \"5a459ada07ea71fed8f3ba28178c333b\"; \n\nconfig.hardware.device(2000).backing.keyId:\u003cunset\u003e -\u003e (keyId = \"a4e7494f-da55-4917-9d90-560c5078866b\",providerId = (id = \"KMS\")); \n\nconfig.hardware.device(2000).iofilter:() -\u003e (\"VMW_vmwarevmcrypt_1.0.0\"); \n\nconfig.extraConfig(\"nvram\").value:\"Test VM.nvram\"-\u003e \"Test VM_1.nvram\"; \n\nconfig.keyId:\u003cunset\u003e -\u003e (keyId = \"a4e7494f-da55-4917-9d90-560c5078866b\",providerId = (id = \"KMS\")); \n\n","Added":"config.extraConfig(\"encryption.bundle\"):(key = \"encryption.bundle\",value = \"vmware:key/list/(pair/(fqid/\u003cVMWARE-NULL\u003e/KMS/a4e7494f%2dda55%2d4917%2d9d90%2d560c5078866b,HMAC%2dSHA%2d256,KPYD5DzgqxZVihEqnejnZEQfltXEmEu3w1sGVvoqtHJzbgDNqi%2fIMD2BehT%2fuykcKXaCaYeCDMIo0KswzfP3Gz1wXN3f5gyKgMhPBqGg%2fLDM4Ee5cqQ%2fbR8MbLromllqsD028Akd%2f9q%2frfn3RAkMlBsQW0H8FKYwbYriKTIKr58dL5GhlCoiZdapRXsyJRvDFdMFdUpkujYUhbMyc2lQAuR35YsXN8WT%2bynYWuEeQ9rraEndXRJg2%2fwLztvziK8bWNDLTe8h%2b3HN%2bcg%2f2aNHDfZCz9%2fnrpJrukSy6lG8Xa05MkVk4%2ftZatnwQXHGafhBg7oT6LVGQfl6muVSpxFhAaFpa0ZcHwadmhBeZDcrb4VeT%2faC0TPn2Pfr3MyXpyb8))\"); \n\n","Deleted":""}},"datacontenttype":"application/json"}'


tests = [
    {  # failure - bad data
        "cloud_event": r'',
        "cases": [
            {
                "env": {},
                "result": {
                    "code": 500,
                    "msg": "Invalid JSON > JSONDecodeError: Expecting value: line 1 column 1 (char 0)"
                }
            }
        ]
    },
    {  # failure - bad data
        "cloud_event": r'"test":"ok"',
        "cases": [
            {
                "env": {},
                "result": {
                    "code": 500,
                    "msg": "Invalid JSON > JSONDecodeError: Extra data: line 1 column 7 (char 6)"
                }
            }
        ]
    },
    {  # failure - bad data
        "cloud_event": r'{"test":"ok"}',
        "cases": [
            {
                "env": {},
                "result": {
                    "code": 500,
                    "msg": "Invalid JSON, required key not found > KeyError: 'data'"
                }
            }
        ]
    },
    {  # failure - bad data
        "cloud_event": r'{"data":"ok"}',
        "cases": [
            {
                "env": {},
                "result": {
                    "code": 500,
                    "msg": "Invalid JSON, data not iterable > TypeError: string indices must be integers"
                }
            }
        ]
    },
    {  # Eventex : vim.event.ResourceExhaustionStatusChangedEvent
        "cloud_event": r'{"id":"0707d7e0-269f-42e7-ae1c-18458ecabf3d","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/eventex","subject":"vim.event.ResourceExhaustionStatusChangedEvent","time":"2020-04-14T00:20:15.100325334Z","data":{"Key":7715,"ChainId":7715,"CreatedTime":"2020-04-14T00:20:13.76967Z","UserName":"machine-bb9a7f","Datacenter":null,"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"vCenter Log File System Resource status changed from Yellow to Green on vcsa.lab ","ChangeTag":"","EventTypeId":"vim.event.ResourceExhaustionStatusChangedEvent","Severity":"info","Message":"","Arguments":[{"Key":"resourceName","Value":"storage_util_filesystem_log"},{"Key":"oldStatus","Value":"yellow"},{"Key":"newStatus","Value":"green"},{"Key":"reason","Value":" "},{"Key":"nodeType","Value":"vcenter"},{"Key":"_sourcehost_","Value":"vcsa.lab"}],"ObjectId":"","ObjectType":"","ObjectName":"","Fault":null},"datacontenttype":"application/json"}',
        "cases": [
            {  # missing key with match_all set
                "env": {
                    "filter_arguments.2.value": r'gree',
                    "filter_nonexistentkey": r'value'
                },
                "match_all": True,
                "result": {
                    "code": 400,
                    "msg": 'some filters not matched'
                }
            },
            {  # missing key but match_all is false
                "env": {
                    "filter_arguments.2.value": r'gree',
                    "filter_nonexistentkey": r'value'
                },
                "result": {
                    "code": 400,
                    "msg": "call_function must be specified"
                }
            },
            {  # array index out of range
                "env": {
                    "filter_arguments.78.value": r'gree',
                },
                "result": {
                    "code": 400,
                    "msg": "Index out of range"
                }
            },
            {  # 'n' array value
                "env": {
                    "filter_arguments.n.value": r'gree'
                },
                "result": {
                    "code": 400,
                    "msg": "call_function must be specified"
                }
            },
            {  # 'n' array value not found
                "env": {
                    "filter_arguments.n.value": r'gree$'
                },
                "result": {
                    "code": 200,
                    "msg": 'Filter "gree$" does not match any value'
                }
            }
        ]
    },
    {  # Standard : DrsVmPoweredOnEvent
        "cloud_event": r'{"id":"c7a6c420-f25d-4e6d-95b5-e273202e1164","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DrsVmPoweredOnEvent","time":"2020-07-02T15:16:13.533866543Z","data":{"Key":130278,"ChainId":130273,"CreatedTime":"2020-07-02T15:16:11.213467Z","UserName":"Administrator","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":{"Name":"Lab","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c47"}},"Host":{"Name":"esxi03.lab","Host":{"Type":"HostSystem","Value":"host-3523"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-82"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"DRS powered on Test VM on esxi03.lab in Lab","ChangeTag":"","Template":false},"datacontenttype":"application/json"}',
        "cases": [
            {  # match some inventory paths
                "env": {
                    "filter_vm": r'/.*/vm/',
                    "filter_host": r'/.*/host/'
                },
                "match_all": True,
                "result": {
                    "code": 400,
                    "msg": 'call_function must be specified'
                }
            },
            {  # match fail some inventory paths
                "env": {
                    "filter_vm": r'/.*/vm/',
                    "filter_host": r'/.*/hos/'
                },
                "match_all": True,
                "result": {
                    "code": 200,
                    "msg": 'Filter "/.*/hos/" does not match "/DC/host/Lab/esxi03.lab"'
                }
            },
            {  # match vm name
                "env": {
                    "filter_vm.name": r'^Test VM$'
                },
                "match_all": True,
                "result": {
                    "code": 400,
                    "msg": 'call_function must be specified'
                }
            }
        ]
    },
    {  # Standard : VmReconfiguredEvent
        "cloud_event": r'{"id":"1fa118f5-c10e-4cd6-bff8-9f569adf19ad","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmReconfiguredEvent","time":"2020-08-25T16:43:38.891590532Z","data":{"Key":23765556,"ChainId":23765534,"CreatedTime":"2020-08-25T16:43:38.085242Z","UserName":"Administrator","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-21"}},"ComputeResource":{"Name":"Lab","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c26"}},"Host":{"Name":"esxi01.lab","Host":{"Type":"HostSystem","Value":"host-409"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-760"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Reconfigured Test VM on esxi01.lab in Lab.\n \nModified: \n \nconfig.hardware.device(2000).backing.fileName:\"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM.vmdk\"-\u003e \"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM_1.vmdk\"; \n\nconfig.hardware.device(2000).backing.contentId:\"29699e346b1037fb6d66817311f0def3\"-\u003e \"5a459ada07ea71fed8f3ba28178c333b\"; \n\nconfig.hardware.device(2000).backing.keyId:\u003cunset\u003e -\u003e (keyId = \"a4e7494f-da55-4917-9d90-560c5078866b\",providerId = (id = \"KMS\")); \n\nconfig.hardware.device(2000).iofilter:() -\u003e (\"VMW_vmwarevmcrypt_1.0.0\"); \n\nconfig.extraConfig(\"nvram\").value:\"Test VM.nvram\"-\u003e \"Test VM_1.nvram\"; \n\nconfig.keyId:\u003cunset\u003e -\u003e (keyId = \"a4e7494f-da55-4917-9d90-560c5078866b\",providerId = (id = \"KMS\")); \n\n Added: \n \nconfig.extraConfig(\"encryption.bundle\"):(key = \"encryption.bundle\",value = \"vmware:key/list/(pair/(fqid/\u003cVMWARE-NULL\u003e/KMS/a4e7494f-da55-4917-9d90-560c5078866b,HMAC-SHA-256,KPYD5DzgqxZVihEqnejnZEQfltXEmEu3w1sGVvoqtHJzbgDNqi/IMD2BehT/uykcKXaCaYeCDMIo0KswzfP3Gz1wXN3f5gyKgMhPBqGg/LDM4Ee5cqQ/bR8MbLromllqsD028Akd/9q/rfn3RAkMlBsQW0H8FKYwbYriKTIKr58dL5GhlCoiZdapRXsyJRvDFdMFdUpkujYUhbMyc2lQAuR35YsXN8WT+ynYWuEeQ9rraEndXRJg2/wLztvziK8bWNDLTe8h+3HN+cg/2aNHDfZCz9/nrpJrukSy6lG8Xa05MkVk4/tZatnwQXHGafhBg7oT6LVGQfl6muVSpxFhAaFpa0ZcHwadmhBeZDcrb4VeT/aC0TPn2Pfr3MyXpyb8))\"); \n\n Deleted: \n \n","ChangeTag":"","Template":false,"ConfigSpec":{"ChangeVersion":"","Name":"","Version":"","CreateDate":"2019-09-12T14:20:06.666314Z","Uuid":"","InstanceUuid":"","NpivNodeWorldWideName":null,"NpivPortWorldWideName":null,"NpivWorldWideNameType":"","NpivDesiredNodeWwns":0,"NpivDesiredPortWwns":0,"NpivTemporaryDisabled":null,"NpivOnNonRdmDisks":null,"NpivWorldWideNameOp":"","LocationId":"","GuestId":"","AlternateGuestName":"","Annotation":"","Files":{"VmPathName":"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM.vmx","SnapshotDirectory":"","SuspendDirectory":"","LogDirectory":"","FtMetadataDirectory":""},"Tools":null,"Flags":null,"ConsolePreferences":null,"PowerOpInfo":null,"NumCPUs":0,"NumCoresPerSocket":0,"MemoryMB":0,"MemoryHotAddEnabled":null,"CpuHotAddEnabled":null,"CpuHotRemoveEnabled":null,"VirtualICH7MPresent":null,"VirtualSMCPresent":null,"DeviceChange":[{"Operation":"edit","FileOperation":"","Device":{"Key":2000,"DeviceInfo":{"Label":"Hard disk 1","Summary":"83,886,080 KB"},"Backing":{"FileName":"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM.vmdk","Datastore":{"Type":"Datastore","Value":"datastore-822"},"BackingObjectId":"","DiskMode":"persistent","Split":false,"WriteThrough":false,"ThinProvisioned":false,"EagerlyScrub":true,"Uuid":"6000C291-35bd-fd2d-e77e-709fb2c53635","ContentId":"29699e346b1037fb6d66817311f0def3","ChangeId":"","Parent":null,"DeltaDiskFormat":"","DigestEnabled":false,"DeltaGrainSize":0,"DeltaDiskFormatVariant":"","Sharing":"sharingNone","KeyId":null},"Connectable":null,"SlotInfo":null,"ControllerKey":1000,"UnitNumber":0,"CapacityInKB":83886080,"CapacityInBytes":85899345920,"Shares":{"Shares":1000,"Level":"normal"},"StorageIOAllocation":{"Limit":-1,"Shares":{"Shares":1000,"Level":"normal"},"Reservation":0},"DiskObjectId":"151-2000","VFlashCacheConfigInfo":null,"Iofilter":null,"VDiskId":null,"NativeUnmanagedLinkedClone":false},"Profile":[{"ProfileId":"b08ff38b-b1d6-413a-9fc1-20aceca2b31c","ReplicationSpec":null,"ProfileData":{"ExtensionKey":"com.vmware.vim.sps","ObjectData":"\u003cns1:storageProfile xmlns:ns1=\"http://profile.policy.data.vasa.vim.vmware.com/xsd\"xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"xsi:type=\"ns1:StorageProfile\"\u003e\u003cns1:constraints\u003e\u003cns1:subProfiles\u003e\u003cns1:capability\u003e\u003ccapabilityId xmlns=\"http://capability.policy.data.vasa.vim.vmware.com/xsd\"\u003e\u003cid\u003evmwarevmcrypt@ENCRYPTION\u003c/id\u003e\u003cnamespace\u003eIOFILTERS\u003c/namespace\u003e\u003c/capabilityId\u003e\u003cconstraint xmlns=\"http://capability.policy.data.vasa.vim.vmware.com/xsd\"\u003e\u003cpropertyInstance\u003e\u003cid\u003eAllowCleartextFilters\u003c/id\u003e\u003cvalue xmlns:s2=\"http://www.w3.org/2001/XMLSchema\"xsi:type=\"s2:string\"\u003eFalse\u003c/value\u003e\u003c/propertyInstance\u003e\u003c/constraint\u003e\u003c/ns1:capability\u003e\u003cns1:name\u003eTag based placement\u003c/ns1:name\u003e\u003c/ns1:subProfiles\u003e\u003c/ns1:constraints\u003e\u003cns1:createdBy\u003eTemporary user handle\u003c/ns1:createdBy\u003e\u003cns1:creationTime\u003e2019-10-25T10:49:35.537+01:00\u003c/ns1:creationTime\u003e\u003cns1:description\u003e\u003c/ns1:description\u003e\u003cns1:generationId\u003e3\u003c/ns1:generationId\u003e\u003cns1:lastUpdatedBy\u003eTemporary user handle\u003c/ns1:lastUpdatedBy\u003e\u003cns1:lastUpdatedTime\u003e2019-11-18T15:20:24.521+00:00\u003c/ns1:lastUpdatedTime\u003e\u003cns1:name\u003eStaging 3PAR Storage with Encryption\u003c/ns1:name\u003e\u003cns1:profileId\u003eb08ff38b-b1d6-413a-9fc1-20aceca2b31c\u003c/ns1:profileId\u003e\u003c/ns1:storageProfile\u003e"},"ProfileParams":null}],"Backing":{"Parent":null,"Crypto":{"CryptoKeyId":{"KeyId":"a4e7494f-da55-4917-9d90-560c5078866b","ProviderId":{"Id":"KMS"}}}}}],"CpuAllocation":null,"MemoryAllocation":null,"LatencySensitivity":null,"CpuAffinity":null,"MemoryAffinity":null,"NetworkShaper":null,"CpuFeatureMask":null,"ExtraConfig":null,"SwapPlacement":"","BootOptions":null,"VAppConfig":null,"FtInfo":null,"RepConfig":null,"VAppConfigRemoved":null,"VAssertsEnabled":null,"ChangeTrackingEnabled":null,"Firmware":"","MaxMksConnections":0,"GuestAutoLockEnabled":null,"ManagedBy":null,"MemoryReservationLockedToMax":null,"NestedHVEnabled":null,"VPMCEnabled":null,"ScheduledHardwareUpgradeInfo":null,"VmProfile":[{"ProfileId":"b08ff38b-b1d6-413a-9fc1-20aceca2b31c","ReplicationSpec":null,"ProfileData":{"ExtensionKey":"com.vmware.vim.sps","ObjectData":"\u003cns1:storageProfile xmlns:ns1=\"http://profile.policy.data.vasa.vim.vmware.com/xsd\"xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"xsi:type=\"ns1:StorageProfile\"\u003e\u003cns1:constraints\u003e\u003cns1:subProfiles\u003e\u003cns1:capability\u003e\u003ccapabilityId xmlns=\"http://capability.policy.data.vasa.vim.vmware.com/xsd\"\u003e\u003cid\u003evmwarevmcrypt@ENCRYPTION\u003c/id\u003e\u003cnamespace\u003eIOFILTERS\u003c/namespace\u003e\u003c/capabilityId\u003e\u003cconstraint xmlns=\"http://capability.policy.data.vasa.vim.vmware.com/xsd\"\u003e\u003cpropertyInstance\u003e\u003cid\u003eAllowCleartextFilters\u003c/id\u003e\u003cvalue xmlns:s1=\"http://www.w3.org/2001/XMLSchema\"xsi:type=\"s1:string\"\u003eFalse\u003c/value\u003e\u003c/propertyInstance\u003e\u003c/constraint\u003e\u003c/ns1:capability\u003e\u003cns1:name\u003eTag based placement\u003c/ns1:name\u003e\u003c/ns1:subProfiles\u003e\u003c/ns1:constraints\u003e\u003cns1:createdBy\u003eTemporary user handle\u003c/ns1:createdBy\u003e\u003cns1:creationTime\u003e2019-10-25T10:49:35.537+01:00\u003c/ns1:creationTime\u003e\u003cns1:description\u003e\u003c/ns1:description\u003e\u003cns1:generationId\u003e3\u003c/ns1:generationId\u003e\u003cns1:lastUpdatedBy\u003eTemporary user handle\u003c/ns1:lastUpdatedBy\u003e\u003cns1:lastUpdatedTime\u003e2019-11-18T15:20:24.521+00:00\u003c/ns1:lastUpdatedTime\u003e\u003cns1:name\u003eStaging 3PAR Storage with Encryption\u003c/ns1:name\u003e\u003cns1:profileId\u003eb08ff38b-b1d6-413a-9fc1-20aceca2b31c\u003c/ns1:profileId\u003e\u003c/ns1:storageProfile\u003e"},"ProfileParams":null}],"MessageBusTunnelEnabled":null,"Crypto":{"CryptoKeyId":{"KeyId":"a4e7494f-da55-4917-9d90-560c5078866b","ProviderId":{"Id":"KMS"}}},"MigrateEncryption":""},"ConfigChanges":{"Modified":"config.hardware.device(2000).backing.fileName:\"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM.vmdk\"-\u003e \"ds:///vmfs/volumes/5da038ba-e597d618-62f1-ecf4bbd3e0d0/Test VM/Test VM_1.vmdk\"; \n\nconfig.hardware.device(2000).backing.contentId:\"29699e346b1037fb6d66817311f0def3\"-\u003e \"5a459ada07ea71fed8f3ba28178c333b\"; \n\nconfig.hardware.device(2000).backing.keyId:\u003cunset\u003e -\u003e (keyId = \"a4e7494f-da55-4917-9d90-560c5078866b\",providerId = (id = \"KMS\")); \n\nconfig.hardware.device(2000).iofilter:() -\u003e (\"VMW_vmwarevmcrypt_1.0.0\"); \n\nconfig.extraConfig(\"nvram\").value:\"Test VM.nvram\"-\u003e \"Test VM_1.nvram\"; \n\nconfig.keyId:\u003cunset\u003e -\u003e (keyId = \"a4e7494f-da55-4917-9d90-560c5078866b\",providerId = (id = \"KMS\")); \n\n","Added":"config.extraConfig(\"encryption.bundle\"):(key = \"encryption.bundle\",value = \"vmware:key/list/(pair/(fqid/\u003cVMWARE-NULL\u003e/KMS/a4e7494f%2dda55%2d4917%2d9d90%2d560c5078866b,HMAC%2dSHA%2d256,KPYD5DzgqxZVihEqnejnZEQfltXEmEu3w1sGVvoqtHJzbgDNqi%2fIMD2BehT%2fuykcKXaCaYeCDMIo0KswzfP3Gz1wXN3f5gyKgMhPBqGg%2fLDM4Ee5cqQ%2fbR8MbLromllqsD028Akd%2f9q%2frfn3RAkMlBsQW0H8FKYwbYriKTIKr58dL5GhlCoiZdapRXsyJRvDFdMFdUpkujYUhbMyc2lQAuR35YsXN8WT%2bynYWuEeQ9rraEndXRJg2%2fwLztvziK8bWNDLTe8h%2b3HN%2bcg%2f2aNHDfZCz9%2fnrpJrukSy6lG8Xa05MkVk4%2ftZatnwQXHGafhBg7oT6LVGQfl6muVSpxFhAaFpa0ZcHwadmhBeZDcrb4VeT%2faC0TPn2Pfr3MyXpyb8))\"); \n\n","Deleted":""}},"datacontenttype":"application/json"}',
        "cases": [
            {  # none existent vc objects
                "env": {
                    "filter_vm": r'/.*/vm/',
                    "filter_host": r'/.*/host/'
                },
                "match_all": True,
                "result": {
                    "code": 400,
                    "msg": 'call_function must be specified'
                }
            },
            {  # none existent vc objects
                "env": {
                    "filter_vm": r'/.*/vm/',
                    "filter_host": r'/.*/host/'
                },
                "match_all": True,
                "result": {
                    "code": 400,
                    "msg": 'call_function must be specified'
                }
            }
        ]
    }
]

test_id = 0

for test in tests:
    for case in test['cases']:
        test_id += 1
        for key, value in case['env'].items():
            os.environ[key] = value  # set env vars

        # set globals
        handler.MATCH_ALL = case.get('match_all', False)

        # run the test
        r = handler.handle(test['cloud_event'])

        if r[0] == case['result']['msg'] and r[1] == case['result']['code']:
            print(f"TEST ID {test_id} PASSED: '{r[0]}' - {r[1]}")
        else:
            print(f"TEST ID {test_id} FAILED: '{r[0]}' - {r[1]}")
            sys.exit()

        for key in case['env'].keys():
            del os.environ[key]  # clear env vars

print("ALL TESTS PASSED!")
