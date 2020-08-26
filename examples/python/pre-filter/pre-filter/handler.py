import ssl
import sys
import json
import os
import requests
import toml
import atexit
import re
import traceback
import socket
from pyVim import connect
from pyVmomi import vim
from pyVmomi import vmodl
from requests.structures import CaseInsensitiveDict


# GLOBAL_VARS
service_instance = None
filters = {}
DEBUG = False
MATCH_ALL = os.getenv("match_all", default='false').lower() == 'true'
INSECURE_SSL = os.getenv("insecure_ssl", default='false').lower() == 'true'

# CONFIG
VC_CONFIG = '/var/openfaas/secrets/vcconfig'


class bgc:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'


if(os.getenv("write_debug", default='false').lower() == 'true'):
    sys.stderr.write(f"{bgc.WARNING}WARNING!! DEBUG has been enabled for this function. Sensitive information could be printed to sysout{bgc.ENDC} \n")
    DEBUG = True


def debug(s):
    if DEBUG:
        sys.stderr.write(s + " \n")  # stderr only get logged on the console logs
        sys.stderr.flush()


def typeName(obj):
    """
    Gets the object name of the passed instance as a string

    Args:
        obj (object): Instance of object to get the type name of

    Returns:
        str: name of the passed objects type
    """
    return obj.__class__.__name__


def faasFunctionName():
    """
    Gets the name of this FaaS function by parsing the k8s pod name

    Returns:
        str: The name of the the FaaS function
    """
    return "-".join(socket.gethostname().split("-")[:-2])


def init():
    """
    Loads the function config and sets up a connection to vCenter

    Returns:
        A tuple in the format of (reason, error_code) is returned on error,
        otherwise resturns None
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
    if(INSECURE_SSL):
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
    Gets the short type name of the passed managed object
    e.g. VirtualMachine

    Args:
        mo (vim.ManagedEntity): The vCenter Managed Object

    Returns:
        str: The type name of the passed managed object
    """
    return typeName(mo).rpartition(".")[2]


def getManagedObject(obj):
    """
    Converts an object as received from the event router in to a pyvmomi managed object

    Args:
        obj (object): object received from the event router

    Returns:
        vim.ManagedEntity: The vCenter Managed Object
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
            debug(f'Found Managed object > {moref} has name {mo.name} and type {getManagedObjectTypeName(mo)}')
            return mo
        except vmodl.fault.ManagedObjectNotFound as err:
            debug(f'{bgc.FAIL}{err.msg}{bgc.ENDC}')
    return None


def getViObjectPath(obj):
    """
    Gets the full path to the passed managed object

    Args:
        obj (vim.ManagedObject): VC managed object

    Returns:
        str: The inventory path to the object
    """
    path = ""
    while obj != service_instance.content.rootFolder:
        path = f'/{obj.name}{path}'
        obj = obj.parent
    return path


def filterVcObject(obj, filter):
    """
    Takes a VC managed object and tests its inventory path against a regex filter

    Args:
        obj (vim.ManagedObject): VC managed Object
        filter (str): Regex filter string

    Returns:
        Returns a boolean on a successful match. If no match then returns
        a tuple in the format of (reason, result_code)
    """
    if obj and filter:
        moType = getManagedObjectTypeName(obj)
        objPath = getViObjectPath(obj)
        debug(f'{bgc.OKBLUE}{moType} Path > {bgc.ENDC}{objPath}')
        return applyFilter(objPath, filter)
    else:
        debug(f'{bgc.WARNING}Object or filter not specified, skipping > Obj: {obj}, Filter: {filter}{bgc.ENDC}')
    return True


def applyFilter(value, filter):
    """
    Applies the regex filter to the passed value

    Args:
        value (str): data to apply filter to
        filter (str): Regex filter string

    Returns:
        Returns a boolean on a successful match. If no match then returns
        a tuple in the format of (reason, result_code)
    """
    try:
        if not re.search(filter, value):
            debug(f'{bgc.WARNING}Filter "{filter}" does not match\n{bgc.OKBLUE}Value >{bgc.ENDC} "{value}".')
            return f'Filter "{filter}" does not match "{value}"', 200
        else:
            debug(f'{bgc.OKBLUE}Match > {bgc.ENDC}Filter "{filter}" matched "{value}"')
            return True
    except re.error as err:
        debug(f'{bgc.FAIL}Invalid regex pattern specified - {err.msg} at pos {err.pos}{bgc.ENDC}')
        return f'Invalid regex pattern specified - {err.msg} at pos {err.pos}', 500
    except TypeError:
        debug(f'{bgc.FAIL}Cannot apply filter against non-string object. Type is: {typeName(value)}{bgc.ENDC}')
        return f'Cannot apply filter against non-string object. Type is: {typeName(value)}', 200


def filterItem(key, filter, data):
    """
    Tries to filter against the passed value

    Args:
        key (str): key name of the filter
        filter (str): the regex filter to use
        data (object): event data to filter

    Returns:
        Returns true on a successful match, or false if the key isn't found
        On error, retuns a tuple in the format of (reason, result_code)
    """
    # Split the key on the first . and capture both oarts
    thisKey, _, remainingKey = key.partition('.')

    # If there is more key to pricess then recurse in
    if remainingKey is not '':
        if type(data) == dict:
            try:
                return filterItem(remainingKey, filter, CaseInsensitiveDict(data)[thisKey])
            except KeyError as err:
                debug(f'{bgc.WARNING}Value not found for key {err}{bgc.ENDC}')
                return False
        elif type(data) == list:
            # If 'n' is passed then loop through the list
            if thisKey is 'n':
                for item in data:
                    res = filterItem(remainingKey, filter, item)
                    if res is True:
                        return True
                return f'Filter "{filter}" does not match any value', 200
            # Otherwise get the passed item by index
            elif thisKey.isnumeric():
                try:
                    return filterItem(remainingKey, filter, data[int(thisKey)])
                except IndexError:
                    debug(f'{bgc.WARNING}Index out of range{bgc.ENDC}')
                    return "Index out of range", 400
            # If we git here there is a key error
            else:
                debug(f'{bgc.WARNING}Invalid array key specified: {thisKey}{bgc.ENDC}')
                return False
    # There's no more key to process so apply the filter
    else:
        try:
            if type(data) == dict:
                value = CaseInsensitiveDict(data)[thisKey]
            elif type(data) == list:
                value = data[int(thisKey)]
            else:
                value = data
        except KeyError as err:
            debug(f'{bgc.WARNING}Value not found for key {err}{bgc.ENDC}')
            return False
        if type(value) == dict:
            mo = getManagedObject(CaseInsensitiveDict(value)[thisKey])
            return filterVcObject(mo, filter)
        else:
            res = applyFilter(value, filter)
            return res


def _getKeys(name, value):
    """
    Recurses in to the passed value building up a list of keys and related values

    Args:
        name (str): key name in dot notation
        value (value): value to recurse in to

    Returns:
        list: list of strings describing event data params and values
    """
    keys = []
    if type(value) == dict:  # for a dict, traverse in to it first...
        for subName, subValue in value.items():
            subName = f'{name}.{subName}'.lower()
            keys.extend(_getKeys(subName, subValue))
        # and if it's a MoRef then get the object's path too
        if len(value.keys()) == 2 and 'Type' in value.keys() and 'Value' in value.keys():
            mo = getManagedObject(CaseInsensitiveDict(value))
            if isinstance(mo, vim.ManagedEntity):
                keys.append(f"{name.rpartition('.')[2]} = {getViObjectPath(mo)}")
            else:  # No object found
                keys.append(f"{name.rpartition('.')[2]} = {bgc.FAIL}<object not found in vCenter>{bgc.ENDC}")
    elif type(value) == list:  # for a list, loop through the contained items using a numeric index
        i = 0
        for subValue in value:
            subName = f'{name}.{i}'.lower()
            keys.extend(_getKeys(subName, subValue))
            i = i + 1
    else:  # for non-iterable items append the name and value
        keys.append(f"{name} = {value}")
    return keys


def getEventKeys(data):
    """
    Builds and returns a list of all available params in the event data and their values

    Args:
        data (object): The CloudEvent data object

    Returns:
        list: list of strings describing event data params and values
    """
    keys = []
    for name, value in data.items():
        keys.extend(_getKeys(name, value))
    return keys


def handle(req):
    """
    Handles a request to the function

    Args:
        req (str): request body

    Returns:
        tuple: Result of function in the format of (result_message, result_code)
    """
    global filters

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
        _ = event['Key']
    except KeyError as err:
        traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
        return f'Invalid JSON, required key not found > KeyError: {err}', 500
    except TypeError as err:
        traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
        return f'Invalid JSON, data not iterable > TypeError: {err}', 500

    debug(f'{bgc.HEADER}Validation passed! Applying object filters:{bgc.ENDC}')

    # Load filter_... environment variables and strip prefix
    filters = {}
    for k, v in os.environ.items():
        if k.startswith("filter_"):
            filters[k[7:]] = v

    # Log all available event keys
    if DEBUG:
        for key in getEventKeys(event):
            debug(f"Event Data > {key}")

    # Loop through each defined filter and apply it if possible
    for key, filter in filters.copy().items():
        debug(f'Key > {key}')
        res = filterItem(key, filter, event)
        if isinstance(res, tuple):  # Tuple is returned if the object didn't match the filter
            return res
        if res is True:
            filters.pop(key)

    # Test if all filters have passed or not
    if len(filters) > 0 and MATCH_ALL is True:
        debug(f'{bgc.WARNING}Some filters not matched when match_all is true - "{", ".join(filters.keys())}". Exiting{bgc.ENDC}')
        return 'some filters not matched', 400

    # All filters matched succesfully so lets call the chained function
    func = os.getenv("call_function", default="")
    if func is not "":
        debug(f'{bgc.HEADER}All filters matched. Calling chained function {func}{bgc.ENDC}')

        # Add this function's name to a faasstack array in the cloud event data
        # This allows the chained function to know who called it
        if 'faasstack' in cevent:
            cevent['faasstack'].append(faasFunctionName())
        else:
            cevent['faasstack'] = [faasFunctionName()]

        # Send a post request to the chained function with the original event data
        try:
            res = requests.post(f'http://gateway.openfaas:8080/function/{func}', json.dumps(cevent))
            return res.text, res.status_code
        except Exception as err:
            traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
            sys.stderr.flush()
            return f'Unexpected error occurred calling chained function "{func}" > Exception: {err}', 500

    debug(f'{bgc.FAIL}call_function must be specified!{bgc.ENDC}')
    return "call_function must be specified", 400
