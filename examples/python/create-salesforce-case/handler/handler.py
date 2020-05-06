import sys, json, os
import urllib,urllib3
import requests
import traceback
import dpath.util
from uuid import uuid4
from json import JSONDecodeError
from .util import Logger, FaaSResponse
#if you add other libraries, make sure to add them to the requirements.txt file

if(os.getenv("insecure_ssl")):
    # Surpress SSL warnings
    urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
l = Logger()

###
# Any Global variables or  Paths and Endpoints
###
META_CONFIG='/var/openfaas/secrets/sfconfig'
### --------

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
            auth = {
                "grant_type": "password",
                "client_id": ref['client_id'],
                "client_secret": ref['client_secret'],
                "username": ref['un'], 
                "password": ref['pwd']
            }
        except (TypeError,KeyError) as err:
            l.log('FAIL', f'Unexpected auth param provided, assuming no auth > "{ref}"')
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
    
    def make_authorization_url(self):
        # Generate a random string for the state parameter
        # Save it for use later to prevent xsrf attacks
        #state = str(uuid4())
        params = self.getauth()
        #l.log('INFO', f'> Auth: {params}') #don't want auth printed
        
        url = "https://login.salesforce.com/services/oauth2/token?" + urllib.parse.urlencode(params)
        return url

    # Salesforce REST API POST w/ OAUTH implementation        
    def post(self):
        """
        Function to make the POST call to the endpoint
        
        Returns:
            [FaaSResponse] -- [Formatted message for OpenFaaS]
        """

        try:
            authn_url = self.make_authorization_url()
            l.log('TITLE', f'> Attempting Authentication request:')
            access_resp = self.session.post(authn_url)
            access_resp.raise_for_status()
            access_body = json.loads(access_resp.text)
            l.log('SUCCESS', f'> Successfully Authenticated: {json.dumps(access_body, indent=4, sort_keys=True)}')
        except requests.HTTPError as err:
            return FaaSResponse('500', 'Error occurred while authenticating to Salesforce > HTTPError: {0}'.format(err))

        try:   
            l.log('TITLE', f'> Attempting POST call to Salesforce')
            urlPath = self.geturl()
            l.log('INFO', f'> URL: {urlPath}')
            headerObj = self.getheaders()
            headerObj["Authorization"] = f"Bearer {access_body['access_token']}"
            l.log('INFO', f'> Headers: {json.dumps(headerObj, indent=4)}')
            bodyObj = self.getbody()
            l.log('INFO', f'> Body: {json.dumps(bodyObj, indent=4)}')

            resp = self.session.post(urlPath, json=bodyObj, headers=headerObj)
            resp.raise_for_status()
            try:
                resp_body = json.loads(resp.text)
                l.log('SUCCESS', f'> Response: {json.dumps(resp_body, indent=4, sort_keys=True)}')
            except json.JSONDecodeError as err:
                l.log('SUCCESS', f'> Response: {resp.text}') #some apis don't return json

            return FaaSResponse('200', 'Success: {resp.text}')
        except requests.exceptions.ConnectionError as err:
            return FaaSResponse('500', 'Error connecting to Salesforce > ConnectionError: {0}'.format(err))
        except requests.HTTPError as err:
            return FaaSResponse('500', 'Error connecting to Salesforce > HTTPError: {0}'.format(err))

def handle(req):

    # Load the Events that function gets from vCenter through the Event Router
    l.log('TITLE', 'Reading Cloud Event: ')
    l.log('INFO', f'Event > {req}')
    try:
        cevent = json.loads(req)
    except json.JSONDecodeError as err:
        res = FaaSResponse('400','Invalid JSON > JSONDecodeError: {0}'.format(err))
        print(json.dumps(vars(res)))
        return

    # Load the Config File
    l.log('TITLE', 'Reading Configuration file: ')
    l.log('INFO', f'Config File > {META_CONFIG}')
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
    l.log('TITLE', 'Validating Input data and mapping: ')
    l.log('INFO', f'Event > {json.dumps(cevent, indent=4, sort_keys=True)}')
    l.log('INFO', f'Config > {json.dumps(metaconfig, indent=4, sort_keys=True)}')
    try:
        #CloudEvent - simple validation
        event = cevent['data']

        #Config - checking for required fields
        url = metaconfig['url']                #not validating if an actual URL is provided
        auth = metaconfig['auth']              #auth key needs to be present
        client_id = auth['client_id']          ##Client ID needs to be present for oauth
        client_secret = auth['client_secret']  ##client_secret needs to be present for oauth
        un = auth['un']                        ##un needs to be present for oauth
        pwd = auth['pwd']                      ##pwd needs to be present for oauth
        headers = metaconfig['headers']        #not validating sanctity of headers
        body = metaconfig['body']              #json only supported
        mappings = metaconfig['mappings']      #mapping can be empty array but the key needs to be present in the config

        #supports 1-1 event-config mapping. next iteration can take complex 1-*(seperated by comma) mapping that will allow building out strings with values from the event
        for mapping in mappings:
            pushvalue = mapping['push']
            l.log('INFO', f'Config has key "{pushvalue}" >>> "{dpath.util.get(body, pushvalue)}"')
            pullvalue = mapping['pull']
            l.log('INFO', f'Event has key "{pullvalue}" >>> "{dpath.util.get(cevent, pullvalue)}"')
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
    l.log('TITLE', 'Attempting connection to Salesforce')
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
