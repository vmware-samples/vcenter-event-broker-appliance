import ssl
import sys
import json
import os
import requests
import toml
import atexit
import re
import traceback
from pyVim import connect
from pyVmomi import vim
from pyVmomi import vmodl


# GLOBAL_VARS
DEBUG = False
# CONFIG
VC_CONFIG = '/var/openfaas/secrets/vcconfig'
service_instance = None


class bgc:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'


if(os.getenv("write_debug")):
    sys.stderr.write(f"{bgc.WARNING}WARNING!! DEBUG has been enabled for this function. Sensitive information could be printed to sysout{bgc.ENDC} \n")
    DEBUG = True


def debug(s):
    if DEBUG:
        sys.stderr.write(s + " \n")  # Syserr only get logged on the console logs
        sys.stderr.flush()


def init():
    """
    Load the config and set up a connection to vc
    """
    global service_instance

    # Load the Config File
    debug(f'{bgc.HEADER}Reading Configuration files: {bgc.ENDC}')
    debug(f'{bgc.OKBLUE}VC Config File > {bgc.ENDC}{VC_CONFIG}')
    try:
        with open(VC_CONFIG, 'r') as vcconfigfile:
            vcconfig = toml.load(vcconfigfile)
            vchost = vcconfig['vcenter']['server']
            vcuser = vcconfig['vcenter']['user']
            vcpass = vcconfig['vcenter']['pass']
    except OSError as err:
        return f'Could not read vcenter configuration: {err}', 500
    except KeyError as err:
        return f'Mandatory configuration key not found: {err}', 500

    try:  # If we already have a valid session then return
        sessionid = service_instance.content.sessionManager.currentSession.key
        if service_instance.content.sessionManager.SessionIsActive(sessionid, vcuser):
            return
    except Exception:
        debug(f'{bgc.WARNING}Init VC Connection...{bgc.ENDC}')

    sslContext = ssl.SSLContext(ssl.PROTOCOL_SSLv23)
    if(os.getenv("insecure_ssl")):
        sslContext.verify_mode = ssl.CERT_NONE

    debug(f'{bgc.OKBLUE}Initialising vCenter connection...{bgc.ENDC}')
    try:
        service_instance = connect.SmartConnect(host=vchost,
                                                user=vcuser,
                                                pwd=vcpass,
                                                port=443,
                                                sslContext=sslContext)
        atexit.register(connect.Disconnect, service_instance)
    except IOError as err:
        return f'Error connecting to vCenter: {err}', 500

    if not service_instance:
        return 'Unable to connect to vCenter host with supplied credentials', 400


def getManagedObjectTypeName(mo):
    """
    Returns the short type name of the passed managed object
    e.g. VirtualMachine
    Args:
        mo (vim.ManagedEntity)
    """
    return mo.__class__.__name__.rpartition(".")[2]


def getManagedObject(obj):
    """
    Convert an object as received from the event router in to a pyvmomi managed object
    Args:
        obj (object): object received from the event router
    """
    mo = None
    try:
        moref = obj['Value']
        type = obj['Type']
    except KeyError as err:
        traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
        return f'Invalid JSON, required key not found > KeyError: {err}', 500

    if hasattr(vim, type):
        typeClass = getattr(vim, type)
        mo = typeClass(moref)
        mo._stub = service_instance._stub
        try:
            debug(f'{bgc.OKBLUE}Managed object > {bgc.ENDC}{moref} has name {mo.name} and type {getManagedObjectTypeName(mo)}')
            return mo
        except vmodl.fault.ManagedObjectNotFound as err:
            debug(f'{bgc.FAIL}{err.msg}{bgc.ENDC}')
    return None


def getViObjectPath(obj):
    """
    Gets the full path to the passed managed object
    Args:
        obj (vim.ManagedObject): VC managed object
    """
    path = ""
    while obj != service_instance.content.rootFolder:
        path = f'/{obj.name}{path}'
        obj = obj.parent
    return path


def filterVcObject(obj, filter):
    """
    Takes a VC managed object and tests it against a filter
    If it doesn't match, a tuple (reason, code) describing why
    is returned, otherwise returns true.
    Args:
        obj (vim.ManagedObject): VC managed Object
        filter (str): Regex filter string
    """

    if obj and filter:
        moType = getManagedObjectTypeName(obj)
        objPath = getViObjectPath(obj)
        debug(f'{bgc.OKBLUE}{moType} Path > {bgc.ENDC}{objPath}')
        try:
            if not re.search(filter, objPath):
                debug(f'{bgc.WARNING}Filter "{filter}" does not match {moType} path "{objPath}". Exiting{bgc.ENDC}')
                return f'Filter "{filter}" does not match {moType} path "{objPath}"', 200
            else:
                debug(f'{bgc.OKBLUE}Match > {bgc.ENDC}Filter matched {moType} path')
        except re.error as err:
            debug(f'{bgc.FAIL}Invalid regex pattern specified - {err.msg} at pos {err.pos}{bgc.ENDC}')
            return f'Invalid regex pattern specified - {err.msg} at pos {err.pos}', 500

    debug(f'{bgc.WARNING}Object or filter not specified, skipping > Obj: {obj}, Filter: {filter}{bgc.ENDC}')


def handle(req):
    """
    Handle a request to the function
    Args:
        req (str): request body
    """

    # Initialise a connection to vCenter
    res = init()
    if isinstance(res, tuple):  # Tuple is returned if an error occurred
        return res
    vcinfo = service_instance.RetrieveServiceContent().about
    debug(f'Connected to {vcinfo.fullName} ({vcinfo.instanceUuid})')

    # Load the Events that function gets from vCenter through the Event Router
    debug(f'{bgc.HEADER}Reading Cloud Event: {bgc.ENDC}')
    debug(f'{bgc.OKBLUE}Event > {bgc.ENDC}{req}')
    try:
        cevent = json.loads(req)
    except json.JSONDecodeError as err:
        return f'Invalid JSON > JSONDecodeError: {err}', 500

    debug(f'{bgc.HEADER}Validating Input data: {bgc.ENDC}')
    debug(f'{bgc.OKBLUE}Event > {bgc.ENDC}{json.dumps(cevent, indent=4, sort_keys=True)}')
    try:
        # CloudEvent - simple validation
        event = cevent['data']
    except KeyError as err:
        traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
        return f'Invalid JSON, required key not found > KeyError: {err}', 500
    except AttributeError as err:
        traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
        return f'Invalid JSON, data not iterable > AttributeError: {err}', 500

    debug(f'{bgc.HEADER}Validation passed! Applying object filters:{bgc.ENDC}')
    # Loop through the event data parameters and find managed objects
    for name, value in event.items():
        if type(value) == dict:  # a dict is probably a VC managed object reference
            for moName, moRef in value.items():
                # Search the items of the dict and look for contained dicts with a 'Type' property
                if type(moRef) == dict and 'Type' in moRef:
                    # Get the relevant filter_... environment variable
                    objFilter = eval(f'os.getenv("filter_{name.lower()}", default=".*")')
                    # Get a vc managed object
                    mo = getManagedObject(moRef)
                    # Test the filter
                    moType = getManagedObjectTypeName(mo)
                    debug(f'{bgc.OKBLUE}Apply Filter > {bgc.ENDC}"{name.lower()}" object ({moType}): "filter_{name.lower()}" = {objFilter}')
                    res = filterVcObject(mo, objFilter)
                    if isinstance(res, tuple):  # Tuple is returned if the object didn't match the filter
                        return res

    func = os.getenv("call_function", False)
    if func:
        debug(f'{bgc.HEADER}All filters matched. Calling chained function {func}{bgc.ENDC}')

        try:
            res = requests.post(f'http://gateway.openfaas:8080/function/{func}', req)
            return res.text, res.status_code
        except Exception as err:
            traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
            sys.stderr.flush()
            return f'Unexpected error occurred calling chained function "{func}" > Exception: {err}', 500

    debug(f'{bgc.FAIL}call_function must be specified!{bgc.ENDC}')
    return "call_function must be specified", 400


#
# Unit Test - helps testing the function locally
# Uncomment r=handle('...') to test the function with the event samples provided below test without deploying to OpenFaaS
#
if __name__ == '__main__':
    VC_CONFIG = 'vc-secrets.toml'
    DEBUG = True
    os.environ['insecure_ssl'] = 'true'
    os.environ['filter_vm'] = '/\w*/vm/Infrastructure/'
    os.environ['call_function'] = 'veba-echo'
    #
    # FAILURE CASES :Invalid Inputs
    #
    # handle('')
    # handle('"test":"ok"')
    # handle('{"test":"ok"}')
    # handle('{"data":"ok"}')

    #
    # SUCCESS CASES :Invalid vc objects
    #
    # handle('{"id":"c7a6c420-f25d-4e6d-95b5-e273202e1164","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DrsVmPoweredOnEvent","time":"2020-07-02T15:16:13.533866543Z","data":{"Key":130278,"ChainId":130273,"CreatedTime":"2020-07-02T15:16:11.213467Z","UserName":"Administrator","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":{"Name":"Lab","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c47"}},"Host":{"Name":"esxi03.lab","Host":{"Type":"HostSystem","Value":"host-9999"}},"Vm":{"Name":"Bad VM","Vm":{"Type":"VirtualMachine","Value":"vm-9999"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"DRS powered on Bad VM on esxi01.lab in Lab","ChangeTag":"","Template":false},"datacontenttype":"application/json"}')

    #
    # SUCCESS CASES
    #
    # Standard : UserLogoutSessionEvent
    # handle('{"id":"17e1027a-c865-4354-9c21-e8da3df4bff9","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"UserLogoutSessionEvent","time":"2020-04-14T00:28:36.455112549Z","data":{"Key":7775,"ChainId":7775,"CreatedTime":"2020-04-14T00:28:35.221698Z","UserName":"machine-b8eb9a7f","Datacenter":null,"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"User machine-b8ebe7eb9a7f@127.0.0.1 logged out (login time: Tuesday, 14 April, 2020 12:28:35 AM, number of API invocations: 34, user agent: pyvmomi Python/3.7.5 (Linux; 4.19.84-1.ph3; x86_64))","ChangeTag":"","IpAddress":"127.0.0.1","UserAgent":"pyvmomi Python/3.7.5 (Linux; 4.19.84-1.ph3; x86_64)","CallCount":34,"SessionId":"52edf160927","LoginTime":"2020-04-14T00:28:35.071817Z"},"datacontenttype":"application/json"}')
    # Eventex : vim.event.ResourceExhaustionStatusChangedEvent
    # handle('{"id":"0707d7e0-269f-42e7-ae1c-18458ecabf3d","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/eventex","subject":"vim.event.ResourceExhaustionStatusChangedEvent","time":"2020-04-14T00:20:15.100325334Z","data":{"Key":7715,"ChainId":7715,"CreatedTime":"2020-04-14T00:20:13.76967Z","UserName":"machine-bb9a7f","Datacenter":null,"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"vCenter Log File System Resource status changed from Yellow to Green on vcsa.lab ","ChangeTag":"","EventTypeId":"vim.event.ResourceExhaustionStatusChangedEvent","Severity":"info","Message":"","Arguments":[{"Key":"resourceName","Value":"storage_util_filesystem_log"},{"Key":"oldStatus","Value":"yellow"},{"Key":"newStatus","Value":"green"},{"Key":"reason","Value":" "},{"Key":"nodeType","Value":"vcenter"},{"Key":"_sourcehost_","Value":"vcsa.lab"}],"ObjectId":"","ObjectType":"","ObjectName":"","Fault":null},"datacontenttype":"application/json"}')
    # Standard : DrsVmPoweredOnEvent
    handle('{"id":"c7a6c420-f25d-4e6d-95b5-e273202e1164","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DrsVmPoweredOnEvent","time":"2020-07-02T15:16:13.533866543Z","data":{"Key":130278,"ChainId":130273,"CreatedTime":"2020-07-02T15:16:11.213467Z","UserName":"Administrator","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":{"Name":"Lab","ComputeResource":{"Type":"ClusterComputeResource","Value":"domain-c47"}},"Host":{"Name":"esxi03.lab","Host":{"Type":"HostSystem","Value":"host-3523"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-82"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"DRS powered on Test VM on esxi03.lab in Lab","ChangeTag":"","Template":false},"datacontenttype":"application/json"}')
    # Standard : VmPoweredOffEvent
    # handle('{"id":"d77a3767-1727-49a3-ac33-ddbdef294150","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmPoweredOffEvent","time":"2020-04-14T00:33:30.838669841Z","data":{"Key":7825,"ChainId":7821,"CreatedTime":"2020-04-14T00:33:30.252792Z","UserName":"Administrator","Datacenter":{"Name":"PKLAB","Datacenter":{"Type":"Datacenter","Value":"datacenter-3"}},"ComputeResource":{"Name":"esxi01.lab","ComputeResource":{"Type":"ComputeResource","Value":"domain-s29"}},"Host":{"Name":"esxi01.lab","Host":{"Type":"HostSystem","Value":"host-31"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-33"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Test VM on  esxi01.lab in PKLAB is powered off","ChangeTag":"","Template":false},"datacontenttype":"application/json"}')
    # Standard : DvsPortLinkUpEvent
    # handle('{"id":"a10f8571-fc2a-40db-8df6-8284cecf5720","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DvsPortLinkUpEvent","time":"2020-07-02T15:16:13.43892986Z","data":{"Key":130277,"ChainId":130277,"CreatedTime":"2020-07-02T15:16:11.207727Z","UserName":"","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":{"Name":"Lab Switch","Dvs":{"Type":"VmwareDistributedVirtualSwitch","Value":"dvs-22"}},"FullFormattedMessage":"The dvPort 2 link was up in the vSphere Distributed Switch Lab Switch in Lab","ChangeTag":"","PortKey":"2","RuntimeInfo":null},"datacontenttype":"application/json"}')
    # Standard : DatastoreRenamedEvent
    # handle('{"id":"369b403a-6729-4b0b-893e-01383c8307ba","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DatastoreRenamedEvent","time":"2020-07-02T21:44:11.09338265Z","data":{"Key":130669,"ChainId":130669,"CreatedTime":"2020-07-02T21:44:08.578289Z","UserName":"","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Renamed datastore from esxi04-local to esxi04-localZ in Lab","ChangeTag":"","Datastore":{"Name":"esxi04-localZ","Datastore":{"Type":"Datastore","Value":"datastore-3313"}},"OldName":"esxi04-local","NewName":"esxi04-localZ"},"datacontenttype":"application/json"}')
    # Standard : DVPortgroupRenamedEvent
    # handle('{"id":"aab77fd1-41ed-4b51-89d3-ef3924b09de1","source":"https://vcsa01.lab/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"DVPortgroupRenamedEvent","time":"2020-07-03T19:36:38.474640186Z","data":{"Key":132376,"ChainId":132375,"CreatedTime":"2020-07-03T19:36:32.525906Z","UserName":"Administrator","Datacenter":{"Name":"Lab","Datacenter":{"Type":"Datacenter","Value":"datacenter-2"}},"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":{"Name":"vMotion AZ","Network":{"Type":"DistributedVirtualPortgroup","Value":"dvportgroup-3357"}},"Dvs":{"Name":"10G Switch A","Dvs":{"Type":"VmwareDistributedVirtualSwitch","Value":"dvs-3355"}},"FullFormattedMessage":"dvPort group vMotion A in Lab was renamed to vMotion AZ","ChangeTag":"","OldName":"vMotion A","NewName":"vMotion AZ"},"datacontenttype":"application/json"}')
