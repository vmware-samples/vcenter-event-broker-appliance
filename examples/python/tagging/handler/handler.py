import sys, json, os
import urllib3
import requests
import toml

# Surpress SSL warnings
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

### VAPI REST endpoints
VAPI_SESSION_PATH='/rest/com/vmware/cis/session'
VAPI_TAG_PATH='/rest/com/vmware/cis/tagging/tag-association/id:'
VC_CONFIG='/var/openfaas/secrets/vcconfig'

### Simple VAPI REST tagging implementation
class FaaSResponse:
    """FaaSResponse is a helper class to construct a properly formatted message returned by this function.
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

class Tagger:
    """Tagger is a vSphere REST API tagging client used to connect and tag objects in vCenter."""    

    def __init__(self,conn):
        """
        
        Arguments:

            conn {Session} -- connection to vCenter REST API
        """        

        try:
            with open(VC_CONFIG, 'r') as vcconfigfile:
                vcconfig = toml.load(vcconfigfile)
                self.vc=vcconfig['vcenter']['server']
                self.username=vcconfig['vcenter']['user']
                self.password=vcconfig['vcenter']['password']
                self.tagurn=vcconfig['tag']['urn']
                self.action=vcconfig['tag']['action'].lower()
        except OSError as e:
            print(f'could not read vcenter configuration: {e}')
            sys.exit(1)
        except KeyError as e:
            print(f'mandatory configuration key not found: {e}')
            sys.exit(1)
        self.session=conn

    # vCenter connection handling    
    def connect(self):
        """performs a login to vCenter
        
        Returns:
            FaaSResponse -- status code and message
        """        
        try:
            resp = self.session.post('https://'+self.vc+VAPI_SESSION_PATH,auth=(self.username,self.password))
            resp.raise_for_status()
            return FaaSResponse('200', 'successfully connected to vCenter')
        except (requests.HTTPError, requests.ConnectionError) as err:
            return FaaSResponse('500', 'could not connect to vCenter {0}'.format(err))

    # VAPI REST tagging implementation        
    def tag(self,obj):
        """tags an object in vCenter
        
        Arguments:

            obj {dict} -- ManagedObjectReference

        Returns:
            FaaSResponse -- status code and message
        """        
        try:
            resp = self.session.post('https://'+self.vc+VAPI_TAG_PATH+self.tagurn+'?~action='+self.action,json=obj)
            resp.raise_for_status()
            print(resp.text)
            return FaaSResponse('200', 'successfully {0}ed tag on: {1}'.format(self.action, obj['object_id']['id']))
        except requests.HTTPError as err:
            return FaaSResponse('500', 'could not tag object {0}'.format(err))

def handle(req):
    # Validate input
    try:
        body = json.loads(req)
    except ValueError as err:
        res = FaaSResponse('400','invalid JSON {0}'.format(err))
        print(json.dumps(vars(res)))
        return
        
    # Assert managed object reference (e.g. to a VM) exists
    # For debugging: validate the JSON blob we received - uncomment if needed
    # print(j)
    try:
        ref = (body['data']['Vm']['Vm'])
    except KeyError as err:
        res = FaaSResponse('400','JSON does not contain ManagedObjectReference {0}'.format(err))
        print(json.dumps(vars(res)))
        return

    # Convert MoRef to an object VAPI REST tagging endpoint requires
    obj = {
        'object_id': {
            'id': ref['Value'],
            'type': ref['Type']
        }
    }

    # Open session to VAPI REST and obtain session token
    s=requests.Session()
    s.verify=False
    t = Tagger(s)
    res = t.connect()
    if res.status != '200':
        print(json.dumps(vars(res)))
        return

    # Perform tagging action on the object
    res = t.tag(obj)
    if res.status != '200':
        print(json.dumps(vars(res)))
        return

    # Close session to VC
    s.close()

    print(json.dumps(vars(res)))
    return

