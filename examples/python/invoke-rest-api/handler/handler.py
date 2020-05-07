import sys, json, os
import urllib3
import requests
import dpath.util
import traceback

# GLOBAL_VARS
DEBUG=False
class bgc:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

if(os.getenv("insecure_ssl")):
    # Surpress SSL warnings
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
if(os.getenv("write_debug")):
    sys.stderr.write(f"{bgc.WARNING}WARNING!! DEBUG has been enabled for this function. Sensitive information could be printed to sysout{bgc.ENDC} \n")
    DEBUG=True

def debug(s):
    if DEBUG:
        sys.stderr.write(s+" \n") #Syserr only get logged on the console logs

#
### Paths and Endpoints
### More examples and details here - https://v2.developer.pagerduty.com/docs/send-an-event-events-api-v2
#
META_CONFIG='/var/openfaas/secrets/metaconfig'

class FaaSResponse:
    """
    FaaSResponse is a helper class to construct a properly formatted message returned by this function.
    By default, OpenFaaS will marshal this response message as JSON.
    """    
    def __init__(self, status, message):
        """
        Arguments:
            status {str} -- the response status code
            message {str} -- the response message
        """    
        self.status=status
        self.message=message

class RESTful:
    """
    RESTful is a class which aims to make Rest API calls easily without writing any code
    """    

    def __init__(self, conn, config, event):
        """
        Constructor for RESTful class
        
        Arguments:
            conn {session} -- [Request Connection]
            config {dict} -- [Config with URL, Body and Mapping for Rest API call]
            event {dict} -- [Cloud Event from vCenter]
        """   
        self.session = conn
        self.config = config
        self.event = event
    
    def geturl(self):
        """
        Getter for the URL to make the call
        
        Returns:
            [string] -- [URL for the Rest API endpoint]
        """
        url = self.config['url']
        return url
    
    def getheaders(self): 
        header = self.config['headers']
        return header
    
    def getauth(self): 
        ref = self.config['auth']
        try:
            auth = (ref['un'], ref['pwd'])
        except (TypeError,KeyError) as err:
            debug(f'Unexpected auth param provided, assuming no auth > "{ref}"')
            auth = None
        return auth
    
    def getbody(self): 
        """
        Getter for the Request Body to make the call
        
        Returns:
            [dict] -- [JSON constructed body]
        """
        body = self.config['body']
        mappings = self.config['mappings']
        for mapping in mappings:
            pushvalue = mapping['push']
            pullvalue = mapping['pull']
            #debug(f'Attempting mapping of Body:{pushvalue} with CloudEvent:{pullvalue}')
            #debug(f'Replacing >>> {dpath.util.get(body, pushvalue)} with ::: {dpath.util.get(self.event, pullvalue)}')
            dpath.util.set(body, pushvalue, dpath.util.get(self.event, pullvalue))
        return body

    # PagerDuty REST API implementation        
    def post(self):
        """
        Function to make the POST call to the endpoint
        
        Returns:
            [FaaSResponse] -- [Formatted message for OpenFaaS]
        """
        
        urlPath = self.geturl()
        debug(f'{bgc.OKBLUE}> URL: {bgc.ENDC}{urlPath}')
        authObj = self.getauth()
        #debug(f'{bgc.OKBLUE}> Auth: {bgc.ENDC}{authObj}') #don't want auth printed
        headerObj = self.getheaders()
        debug(f'{bgc.OKBLUE}> Headers: {bgc.ENDC}{json.dumps(headerObj, indent=4)}')
        bodyObj = self.getbody()
        debug(f'{bgc.OKBLUE}> Body: {bgc.ENDC}{json.dumps(bodyObj, indent=4)}')
        try:
            resp = self.session.post(urlPath, auth=authObj, json=bodyObj, headers=headerObj)
            resp.raise_for_status()
            try:
                resp_body = json.loads(resp.text)
                debug(f'{bgc.OKBLUE}> Response: {bgc.ENDC}{json.dumps(resp_body, indent=4, sort_keys=True)}')
            except json.JSONDecodeError as err:
                debug(f'{bgc.OKBLUE}> Response: {bgc.ENDC}{resp.text}') #some apis don't return json
            
            return FaaSResponse('200', f'Response:{resp.text}')
        except requests.HTTPError as err:
            return FaaSResponse('500', 'Could not executed REST API > HTTPError: {0}'.format(err))

def handle(req):
    
    # Load the Events that function gets from vCenter through the Event Router
    debug(f'{bgc.HEADER}Reading Cloud Event: {bgc.ENDC}')
    debug(f'{bgc.OKBLUE}Event > {bgc.ENDC}{req}')
    try:
        cevent = json.loads(req)
    except json.JSONDecodeError as err:
        res = FaaSResponse('400','Invalid JSON > JSONDecodeError: {0}'.format(err))
        print(json.dumps(vars(res)))
        return

    # Load the Config File
    debug(f'{bgc.HEADER}Reading Configuration file: {bgc.ENDC}')
    debug(f'{bgc.OKBLUE}Config File > {bgc.ENDC}{META_CONFIG}')
    try: 
        with open(META_CONFIG, 'r') as prodconfig:
            metaconfig = json.load(prodconfig)
    except json.JSONDecodeError as err:
        res = FaaSResponse('400','Invalid JSON > JSONDecodeError: {0}'.format(err))
        print(json.dumps(vars(res)))
        return
    except OSError as err:
        res = FaaSResponse('500','Could not read configuration > OSError: {0}'.format(err))
        print(json.dumps(vars(res)))
        return

    #Validate CloudEvent and Configuration for mandatory fields
    debug(f'{bgc.HEADER}Validating Input data and mapping: {bgc.ENDC}')
    debug(f'{bgc.OKBLUE}Event > {bgc.ENDC}{json.dumps(cevent, indent=4, sort_keys=True)}')
    debug(f'{bgc.OKBLUE}Config > {bgc.ENDC}{json.dumps(metaconfig, indent=4, sort_keys=True)}')
    try:
        #CloudEvent - simple validation
        event = cevent['data']
        
        #Config - checking for required fields
        url = metaconfig['url'] #not validating if an actual URL is provided
        auth = metaconfig['auth'] #auth can be empty for no auth but a required key
        headers = metaconfig['headers'] #not validating sanctity of headers
        body = metaconfig['body'] #json only supported
        mappings = metaconfig['mappings'] #mapping can be empty array but the key needs to be present in the config
        
        #supports 1-1 event-config mapping. next iteration can take complex 1-*(seperated by comma) mapping that will allow building out strings with values from the event
        for mapping in mappings:
            pushvalue = mapping['push']
            debug(f'Config has key "{pushvalue}" >>> "{dpath.util.get(body, pushvalue)}"')
            pullvalue = mapping['pull']
            debug(f'Event has key "{pullvalue}" >>> "{dpath.util.get(cevent, pullvalue)}"')
    except KeyError as err:
        res = FaaSResponse('400','Invalid JSON, required key not found > KeyError: {0}'.format(err))
        traceback.print_exc(limit=1, file=sys.stderr) #providing traceback since it helps debug the exact key that failed
        print(json.dumps(vars(res)))
        return
    except ValueError as err:
        res = FaaSResponse('400','Invalid mapping, multiple keys found > ValueError: {0}'.format(err))
        traceback.print_exc(limit=1, file=sys.stderr) #providing traceback since it helps debug the exact key that failed
        print(json.dumps(vars(res)))
        return

    # Make the Rest Api Call
    s=requests.Session()
    if(os.getenv("insecure_ssl")):
        s.verify=False
    
    # with the metaconfig - which is the configuration file with the URL and body to make the call
    # and with the cloud event - which is the event generated from vCenter
    # we are going to build the request body and make the rest api call
    debug(f'{bgc.HEADER}Attemping HTTP POST: {bgc.ENDC}')
    try:
        restful = RESTful(s, metaconfig, cevent)
        res = restful.post()
        print(json.dumps(vars(res)))
    except Exception as err:
        res = FaaSResponse('500','Unexpected error occurred > Exception: {0}'.format(err))
        traceback.print_exc(limit=1, file=sys.stderr) #providing traceback since it helps debug the exact key that failed
        print(json.dumps(vars(res)))
        
    s.close()

    return

#
## Unit Test - helps testing the function locally
## Uncomment META_CONFIG - update the path to the file accordingly
## Uncomment handle('...') to test the function with the event samples provided below test without deploying to OpenFaaS
#
#META_CONFIG='metaconfig-pduty.json'

#
## FAILURE CASES :Invalid Inputs
#
#handle('')
#handle('"test":"ok"')
#handle('{"test":"ok"}')
#handle('{"data":"ok"}')

#
## FAILURE CASES :Unhandled Events
# 
# Standard : UserLogoutSessionEvent
#handle('{"id":"17e1027a-c865-4354-9c21-e8da3df4bff9","source":"https://vcsa.pdotk.local/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"UserLogoutSessionEvent","time":"2020-04-14T00:28:36.455112549Z","data":{"Key":7775,"ChainId":7775,"CreatedTime":"2020-04-14T00:28:35.221698Z","UserName":"machine-b8eb9a7f","Datacenter":null,"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"User machine-b8ebe7eb9a7f@127.0.0.1 logged out (login time: Tuesday, 14 April, 2020 12:28:35 AM, number of API invocations: 34, user agent: pyvmomi Python/3.7.5 (Linux; 4.19.84-1.ph3; x86_64))","ChangeTag":"","IpAddress":"127.0.0.1","UserAgent":"pyvmomi Python/3.7.5 (Linux; 4.19.84-1.ph3; x86_64)","CallCount":34,"SessionId":"52edf160927","LoginTime":"2020-04-14T00:28:35.071817Z"},"datacontenttype":"application/json"}')
# Eventex : vim.event.ResourceExhaustionStatusChangedEvent
#handle('{"id":"0707d7e0-269f-42e7-ae1c-18458ecabf3d","source":"https://vcsa.pdotk.local/sdk","specversion":"1.0","type":"com.vmware.event.router/eventex","subject":"vim.event.ResourceExhaustionStatusChangedEvent","time":"2020-04-14T00:20:15.100325334Z","data":{"Key":7715,"ChainId":7715,"CreatedTime":"2020-04-14T00:20:13.76967Z","UserName":"machine-bb9a7f","Datacenter":null,"ComputeResource":null,"Host":null,"Vm":null,"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"vCenter Log File System Resource status changed from Yellow to Green on vcsa.pdotk.local  ","ChangeTag":"","EventTypeId":"vim.event.ResourceExhaustionStatusChangedEvent","Severity":"info","Message":"","Arguments":[{"Key":"resourceName","Value":"storage_util_filesystem_log"},{"Key":"oldStatus","Value":"yellow"},{"Key":"newStatus","Value":"green"},{"Key":"reason","Value":" "},{"Key":"nodeType","Value":"vcenter"},{"Key":"_sourcehost_","Value":"vcsa.pdotk.local"}],"ObjectId":"","ObjectType":"","ObjectName":"","Fault":null},"datacontenttype":"application/json"}')

#
## SUCCESS CASES
#
# Standard : VmPoweredOnEvent
#handle('{"id":"453120cd-3d19-4c43-aadc-df0cdbce3887","source":"https://vcsa.pdotk.local/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmPoweredOnEvent","time":"2020-04-13T23:46:10.402531287Z","data":{"Key":7441,"ChainId":7438,"CreatedTime":"2020-04-13T23:46:09.387283Z","UserName":"Administrator","Datacenter":{"Name":"PKLAB","Datacenter":{"Type":"Datacenter","Value":"datacenter-3"}},"ComputeResource":{"Name":"esxi01.pdotk.local","ComputeResource":{"Type":"ComputeResource","Value":"domain-s29"}},"Host":{"Name":"esxi01.pdotk.local","Host":{"Type":"HostSystem","Value":"host-31"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-33"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Test VM on esxi01.pdotk.local in PKLAB has powered on","ChangeTag":"","Template":false},"datacontenttype":"application/json"}')
# Standard : VmPoweredOffEvent
#handle('{"id":"d77a3767-1727-49a3-ac33-ddbdef294150","source":"https://vcsa.pdotk.local/sdk","specversion":"1.0","type":"com.vmware.event.router/event","subject":"VmPoweredOffEvent","time":"2020-04-14T00:33:30.838669841Z","data":{"Key":7825,"ChainId":7821,"CreatedTime":"2020-04-14T00:33:30.252792Z","UserName":"Administrator","Datacenter":{"Name":"PKLAB","Datacenter":{"Type":"Datacenter","Value":"datacenter-3"}},"ComputeResource":{"Name":"esxi01.pdotk.local","ComputeResource":{"Type":"ComputeResource","Value":"domain-s29"}},"Host":{"Name":"esxi01.pdotk.local","Host":{"Type":"HostSystem","Value":"host-31"}},"Vm":{"Name":"Test VM","Vm":{"Type":"VirtualMachine","Value":"vm-33"}},"Ds":null,"Net":null,"Dvs":null,"FullFormattedMessage":"Test VM on  esxi01.pdotk.local in PKLAB is powered off","ChangeTag":"","Template":false},"datacontenttype":"application/json"}')
