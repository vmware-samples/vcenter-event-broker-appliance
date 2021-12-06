from flask import request, jsonify, Blueprint
from flask import current_app as app
from werkzeug.exceptions import InternalServerError
from requests.exceptions import RequestException
from cloudevents.http import from_http, to_json
from cloudevents.exceptions import MissingRequiredFields
import logging
import sys
import json
import os
import urllib3
import requests
import traceback
import time
import regex as re
from dateutil.parser import isoparse

#####
vchost = None
vrotoken = None 
vroinsecure = False
bearer_token = None

class bgc:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'
#####

logging.basicConfig(level=logging.DEBUG,format='%(asctime)s %(levelname)s %(name)s %(threadName)s : %(message)s')


def homogeneous_type(seq):
    """
    Returns the vRO data type of the passed properties if they are all the
    same, otherwise returns False

    Args:
        seq (dict): The vRO properties to be tested

    Returns:
        str/bool: The vRO data type or False
    """
    iseq = iter(seq)
    first_type = list(next(iseq).keys())[0]
    return first_type if all( (list(x.keys())[0] is first_type) for x in iseq ) else False


def camelCase(string):
    """
    Return the passed string as camelCase

    Args:
        string (str): any string

    Returns:
        str: camelCase representation of the string
    """
    
    # Split the string in to words based on uppercase letters, spaces, underscores and hyphens
    words = re.split(r'(?<=[a-z])(?=[A-Z])|(?<=[A-Z])(?=[A-Z][a-z])|[\s_-]+', string, flags = re.VERSION1)
    # Join words camelcased
    s = "".join(word[0].upper() + word[1:].lower() for word in words)
    return s[0].lower() + s[1:]


def getVroInputParam(item):
    """
    Takes an object from the event router data and turns it in to a vRO input parameter
    
    Args:
        item (tuple): event router event data parameter name, value pair
    
    Returns:
        object: A vRO input parameter
    """

    # Validate the input data
    if type(item) is not tuple:
        app.logger.error(f'{bgc.FAIL}Passed item is not a tuple! Item > {item}')
        return None
    if len(item) != 2:
        app.logger.error(f'{bgc.FAIL}Passed item has a length of {len(item)} - must be 2')
        return None
    if type(item[0]) != str:
        app.logger.error(f'{bgc.FAIL}Passed item[0] has a type of {type(item[0])} - must be str')
        return None

    name, value = item
    param = {
        "scope": "local"
    }
    # Determine the data type of the object and create a vRO input parameter
    if type(value) == int or type(value) == float:
        param['type'] = "number"
        param['value'] = {"number": {"value": value}}

    elif isinstance(value, str):
        try:  # for strings, try and parse to a date first...
            isoparse(value)
            param['type'] = "Date"
            param['value'] = {"date": {"value": value}}
        except ValueError:  # ...if that doesn't work, see if it's a number
            try:
                float(value)
                param['type'] = "number"
                param['value'] = {"number": {"value": float(value)}}
            except ValueError:  # ...and if that doesn't work just build a string
                param['type'] = "string"
                param['value'] = {"string": {"value": value}}

    elif type(value) == bool:  # booleans are easy, jsut convert to string representation
        param['type'] = "boolean"
        param['value'] = {"boolean": {"value": value}}

    elif type(value) == dict:  # handle dicts as a vRO Properties data type, unless they're a VC SDK object
        for v in value.values():
            # search the items of the dict and look for contained dicts with a 'Type' and
            # 'Value' property as they're probably SDK objects
            if type(v) == dict and 'Type' in v and 'Value' in v:
                param['type'] = f"VC:{v['Type']}"
                param['value'] = {
                    "sdk-object": {
                        "type": param['type'],
                        "id": f'{vchost},id:{v["Value"]}'
                    }
                }
        if not 'type' in param:  # if no vc managed object was found, parse the data as Properties object
            param['type'] = "Properties"
            # This line parses the dict value as a list of key/value items
            param['value'] = getVroInputParam((name, list(value.items())))['value']

    elif type(value) in (list, tuple):
        # lists or tuples have a number of options:
        # - if all the contained items are the same data type, then pass to vRO as an array of that type
        # - if all the contained items are of different types, then pass to vRO as a generic 'Array' type
        # - if all the contained items are dicts containing 'key'/'value' properties then pass to vRO as Properties
        # - if all the contained items are list/tuple with length 2 and a string first value, also treat as vRO Properties 
        properties = []
        for v in value:
            if type(v) == dict and 'Key' in v and 'Value' in v:
                param['type'] = "Properties"
                # pass the key and value to this function recursively to get a correctly structured value
                vroParam = getVroInputParam((v["Key"], v["Value"]))
                properties.append({"key": v["Key"], "value": vroParam["value"]})
            elif type(v) in (tuple, list) and len(v) == 2 and type(v[0]) == str:
                vroParam = getVroInputParam(tuple(v))
                properties.append({"key": v[0], "value": vroParam["value"]})
            else:
                properties.append(getVroInputParam(("dummyName", v))["value"])

        t = homogeneous_type(properties)  # get the parsed datatype of all values, or false if there are diffs
        
        if not t:  # The datatypes in the source data have differences, so treat as a generic vRO Array
            param['type'] = "Array"
            param['value'] = {"array": {"elements": properties}}
        elif t == 'key': # Key/value pairs are in the data, so treat as vRO Properties
            param['type'] = "Properties"
            param['value'] = {"properties": {"property": properties}}
        else: # anything else is a list of identical datatypes, so treat as an array of that type
            param['type'] = f'Array/{list(properties[0].keys())[0]}'
            param['value'] = {"array": {"elements": properties}}

    elif value is None:  # handle empty/null data
        param = None

    else:  # Unhandled data type - try and wrangle to a string
        app.logger.debug(f'{bgc.WARNING}Unhandled data type "{type(value).__name__}" - forcing to string{bgc.ENDC}')
        param['type'] = "string"
        try:
            param['value'] = {"string": {"value": str(value)}}
        except:
            app.logger.debug(f'{bgc.FAIL}Failed to convert to string - returning None!')
            param = None

    if type(param) == dict:  # set the parameter name for valid types
        param["name"] = camelCase(name)

    return param


def get_response(message, code):
    """
    Get's a formated response object

    Args:
        message (string): The response message
        code (integer): The result code of the response

    Returns:
        Object: correctly formatted response object
    """
    if 200 <= code < 300:
        app.logger.info(message)
    else:
        app.logger.error(f'{bgc.FAIL}{message}{bgc.ENDC}')
    msg = {
        'status': code,
        'message': message,
    }
    resp = jsonify(msg)
    resp.status_code = code
    return resp


#app = Flask(__name__)
bp = Blueprint('handler', __name__)
@bp.route('/', methods=['POST'])
def handle():
    """
    Handles a request to the function

    Returns:
        Object: HTTP response object
    """
    global vchost, bearer_token

    # Load the Events that function gets from vCenter through the Event Router
    try:
        app.logger.debug(f'{bgc.HEADER}Reading Cloud Event: {bgc.ENDC}')
        cevent = from_http(request.headers, request.get_data(),None)

        if cevent["datacontenttype"].lower() != "application/json":
            return get_response(f'invalid datacontenttype for cloud event: {cevent["datacontenttype"]}', 400)

        cevent_as_json = json.loads(to_json(cevent).decode("utf-8"))

        # Pretty Print the CloudEvent data
        app.logger.debug(f'{bgc.OKBLUE}Event > {bgc.ENDC}{json.dumps(cevent_as_json, indent=4, sort_keys=True)}')

        # Get the source vCenter hostname
        source = cevent['source']
        vchost = urllib3.util.url.parse_url(source).host

    except MissingRequiredFields as err:
        return get_response(f'Failed to read CloudEvent data > MissingRequiredFields: {err}', 400)
    except KeyError as err:
        traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
        return get_response(f'Invalid CloudEvent, required key not found > KeyError: {err} > {cevent}', 500)
    except TypeError as err:
        traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
        return get_response(f'Invalid CloudEvent, data not iterable > TypeError: {err}', 500)
    except AttributeError as err:
        traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
        return get_response(f'Invalid CloudEvent, required attribute not found > AttributeError: {err} > {cevent}', 500)

    app.logger.debug(f'{bgc.HEADER}Validation passed! Building vRO request:{bgc.ENDC}')

    # Build the vRO input parameter object using the event data
    body = {"parameters": []}
    for item in cevent.data.items():
        # Get the vRO input parameter correctly formatted for the received event data
        res = getVroInputParam(item)
        if res:
            body["parameters"].append(res)

    # Add CloudEvent Attributes and raw JSON CloudEvent as individual vRO input parameters
    body["parameters"].append(getVroInputParam(("cloudEventAttributes", list(cevent._attributes.items()))))
    body["parameters"].append(getVroInputParam(("rawEventData", json.dumps(cevent_as_json))))

    # Add the parameter list built above as input to vRO
    vro_inputs = ''
    for p in body["parameters"]:
        vro_inputs += f'Param: {p["name"]} - Type: {p["type"]}\n'
    body["parameters"].append(getVroInputParam(("availableInputs", vro_inputs)))
    app.logger.debug(f'REST body: {json.dumps(body, indent=4)}')

    if app.debug:
        app.logger.debug(f'{bgc.HEADER}Passing the following params to vRO:{bgc.ENDC}')
        for p in body["parameters"]:
            app.logger.debug(f'Param: {p["name"]} - Type: {p["type"]}')

    app.logger.debug(f'Workflow ID: {wfId}')

    vroUrl = f'https://{vrohost}:{vroport}/vco/api/workflows/{wfId}/executions'

    app.logger.debug(f'{bgc.HEADER}Attemping HTTP POST: {bgc.ENDC}')
    retries = 1
    success = False
    while not success:
        # POST to vRO REST API
        if vroauth == "BASIC":
            r = requests.post(vroUrl,
                          auth=(vrouser.encode('utf-8'), vropass.encode('utf-8')),
                          json=body,
                          verify=not vroinsecure
                          )
        if vroauth == "OAUTH":
            r = requests.post(vroUrl,
                          headers={"Authorization":f'Bearer {bearer_token}'},
                          json=body,
                          verify=not vroinsecure
                          )
            if r.status_code == 401:
                bearer_token = get_token()
                if not isinstance(bearer_token, str):
                    return bearer_token  # if the result isn't a string, it is a response object as a result of a failure
        if r.ok:
            success = True
        if retries < 0:
            return get_response('vRO REST Request failed > Exception: {0}'.format(r.reason), r.status_code)
        retries -= 1
        time.sleep(2)

    app.logger.debug(f'{bgc.OKBLUE}POST Successful...{bgc.ENDC}')

    try:
        vro_res = json.loads(r.text)
    except json.decoder.JSONDecodeError:
        traceback.print_exc(limit=1, file=sys.stderr)  # providing traceback since it helps debug the exact key that failed
        return get_response(f'Response is not valid JSON\n{r.text}', r.status_code)

    if r.ok:
        app.logger.debug(f'Successfully executed vRO workflow: {vro_res["name"]}')

        app.logger.debug(f'{bgc.OKBLUE}vRO Response: {bgc.ENDC}')
        app.logger.debug(json.dumps(vro_res, indent=4))
        return get_response(f'Workflow Token {vro_res["id"]}: {vro_res["state"]}', r.status_code)
    else:
        app.logger.debug(f'{bgc.FAIL}Failed to execute workflow:{bgc.ENDC}')
        return get_response(vro_res["message"], vro_res["status"])


@bp.before_app_first_request
def before_first_request():
    """
    Runs before the first request in order to perform function initialisation 
    """
    global vroconfig, vrohost, vroport, vroauth, vrouser, vropass, vrotoken, vroinsecure, wfId, bearer_token

    # Get function config values
    try:
        vroconfig = json.loads(os.environ['VROCONFIG_SECRET'])
        vrohost = vroconfig['SERVER']
        vroport = vroconfig['PORT']
        vroauth = vroconfig['AUTH_TYPE'].upper()
        if vroauth not in ('BASIC', 'OAUTH'):
            raise KeyError("AUTH_TYPE must be BASIC or OAUTH")
        if vroauth == 'BASIC' or not vroconfig.get('REFRESH_TOKEN'):
            vrouser = vroconfig['USERNAME']
            vropass = vroconfig['PASSWORD']
            vrotoken = None
        else:
            vrotoken = vroconfig['REFRESH_TOKEN']
        vroinsecure = vroconfig['INSECURE_SSL']
        wfId = vroconfig['WORKFLOW_ID']
    except json.decoder.JSONDecodeError as err:
        app.logger.error(f'{bgc.FAIL}Could not read vro configuration:{bgc.ENDC} {err}')
        raise InternalServerError(description=f'Could not read vro configuration: {err}')
    except KeyError as err:
        app.logger.error(f'{bgc.FAIL}Mandatory configuration key not found:{bgc.ENDC} {err}')
        raise InternalServerError(description=f'Mandatory configuration key not found: {err}')
        
    if(vroinsecure):
        # Surpress SSL warnings
        app.logger.debug(f'{bgc.WARNING}Using insecure SSL!!{bgc.ENDC}')
        urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    
    bearer_token = get_token()


@bp.errorhandler(InternalServerError)
def handle_internal_server_error(e):
    return get_response(str(e.original_exception), e.original_exception.code), e.original_exception.code


def get_token():
    """
    If using OAUTH this requests a new bearer token for later use

    Returns:
        Either a bearer token (as a string), or a response object in the case of a failure
    """
    global vrotoken

    try:
        if vroauth == "OAUTH":
            if not vrotoken:
                app.logger.debug(f'{bgc.OKBLUE}Attempting to fetch refresh token for user {vrouser}{bgc.ENDC}')
                r = requests.post(f'https://{vrohost}:{vroport}/csp/gateway/am/api/login?access_token',
                        json={"username":vrouser, "password":vropass},
                        verify=not vroinsecure)
                r.raise_for_status()
                vrotoken = json.loads(r.text)['refresh_token']
                auth_url = f'https://{vrohost}:{vroport}'
            else:
                auth_url = 'https://api.mgmt.cloud.vmware.com'
            app.logger.debug(f'{bgc.OKBLUE}Attempting to fetch bearer token from {auth_url}{bgc.ENDC}')
            r = requests.post(f'{auth_url}/iaas/api/login',
                        json={"refreshToken":vrotoken},
                        verify=not vroinsecure)
            r.raise_for_status()
            bearer_token = json.loads(r.text)['token']
            return bearer_token

    except KeyError as err:
        return get_response(f'Authentication response invalid: {err}', 500)
    except ConnectionError as err:
        return get_response(f'Connection error during authentication: {err}', 500)
    except RequestException as err:
        return get_response(f'Error with authentication request: {err}', 500)


if __name__ == '__main__':
    app.run()