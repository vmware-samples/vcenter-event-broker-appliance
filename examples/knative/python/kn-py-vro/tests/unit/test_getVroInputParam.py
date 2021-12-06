##
# These tests test the getVroInpuParam function and can be run without a vRO
##

import kn_py_vro.handler as handler
import json
from kn_py_vro.handler import getVroInputParam, camelCase


def get_data(key):
    with open("./tests/unittestdata.json", "r") as f:
        data = json.load(f)
    return (key, data[key])


def test_camelcase():
    assert camelCase('camelCaseName') == 'camelCaseName'
    assert camelCase('Camel Case Name') == 'camelCaseName'
    assert camelCase('camelCase  Name') == 'camelCaseName'
    assert camelCase('CamelCaseName') == 'camelCaseName'
    assert camelCase('CAMEL case Name') == 'camelCaseName'
    assert camelCase('camel   case name') == 'camelCaseName'
    assert camelCase('camel_case-name') == 'camelCaseName'
    assert camelCase('CAMELCaseNAME') == 'camelCaseName'
    assert camelCase('CAMEL-CASE_NAME') == 'camelCaseName'
    

def test_null_data(client):
    data = get_data('Null')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r == None


def test_int_data(client):
    data = get_data('Int')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'int'
    assert r['type'] == 'number'
    assert r['value']['number']['value'] == 56


def test_float_data(client):
    data = get_data('Float')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'float'
    assert r['type'] == 'number'
    assert r['value']['number']['value'] == 1.56


def test_string_data(client):
    data = get_data('String')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'string'
    assert r['type'] == 'string'
    assert r['value']['string']['value'] == 'String data'


def test_string_number_data(client):
    data = get_data('String Number')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'stringNumber'
    assert r['type'] == 'number'
    assert r['value']['number']['value'] == 5.6


def test_string_date_data(client):
    data = get_data('String Date')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'stringDate'
    assert r['type'] == 'Date'
    assert r['value']['date']['value'] == '2021-05-04T07:33:33.773581268Z'


def test_boolean_data(client):
    data = get_data('Boolean')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'boolean'
    assert r['type'] == 'boolean'
    assert r['value']['boolean']['value'] == False


def test_sdk_dict_vm_data(client):
    """
    A dict representing a VC SDK object parsed as a native vRO SDK object
    """

    data = get_data('SDK Dict Vm')
    handler.vchost = 'vcsa.lab.local'
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'sdkDictVm'
    assert r['type'] == 'VC:VirtualMachine'
    assert r['value']['sdk-object']['type'] == 'VC:VirtualMachine'
    assert r['value']['sdk-object']['id'] == 'vcsa.lab.local,id:vm-596'


def test_sdk_dict_host_data(client):
    """
    A dict representing a VC SDK object parsed as a native vRO SDK object
    """

    data = get_data('SDK Dict Host')
    handler.vchost = 'vcsa.lab.local'
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'sdkDictHost'
    assert r['type'] == 'VC:HostSystem'
    assert r['value']['sdk-object']['type'] == 'VC:HostSystem'
    assert r['value']['sdk-object']['id'] == 'vcsa.lab.local,id:host-34'


def test_raw_dict_data(client):
    """
    A normal dict should be parsed as a vRO Properties object
    """

    data = get_data('Raw Dict')
    handler.vchost = 'vcsa.lab.local'
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'rawDict'
    assert r['type'] == 'Properties'
    p = r['value']['properties']['property'][0]
    assert p['key'] == "Test String"
    assert p['value']['string']['value'] == "foo bar"
    p = r['value']['properties']['property'][1]
    assert p['key'] == "Test Number"
    assert p['value']['number']['value'] == 0.56
    p = r['value']['properties']['property'][2]
    assert p['key'] == "Test List"
    assert p['value']['array']['elements'][0]['number']['value'] == 1
    assert p['value']['array']['elements'][1]['string']['value'] == "two"
    assert p['value']['array']['elements'][2]['properties']['property'][0]['key'] == "Three"
    assert p['value']['array']['elements'][2]['properties']['property'][0]['value']['number']['value'] == 4.1
    p = r['value']['properties']['property'][3]
    assert p['key'] == "Test Dict"
    assert p['value']['properties']['property'][0]['key'] == "One"
    assert p['value']['properties']['property'][0]['value']['number']['value'] == 1
    assert p['value']['properties']['property'][1]['key'] == "Two"
    assert p['value']['properties']['property'][1]['value']['number']['value'] == -2
    assert p['value']['properties']['property'][2]['key'] == "Three"
    assert p['value']['properties']['property'][2]['value']['number']['value'] == 0.3
    

def test_list_kv_dicts_data(client):
    """
    A list of dicts each with a "key" and "value" keys should be parsed as a vRO Properties object
    """
    data = get_data('List Key Value Dicts')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'listKeyValueDicts'
    assert r['type'] == 'Properties'
    p = r['value']['properties']['property'][0]
    assert p['key'] == "Test string"
    assert p['value']['string']['value'] == "String value"
    p = r['value']['properties']['property'][1]
    assert p['key'] == "Test number"
    assert p['value']['number']['value'] == 123
    p = r['value']['properties']['property'][2]
    assert p['key'] == "Test date"
    assert p['value']['date']['value'] == "2021-05-04T07:33:33.773581268Z"


def test_list_tuple_data(client):
    """
    Tuples/lists with a length of 2 and the first item being a string
    should be parsed as a key/value vRO Properties object
    """

    data = get_data('List Tuple')
    with client.application.app_context():
        r = getVroInputParam(data)
    
    assert r['name'] == 'listTuple'
    assert r['type'] == 'Properties'
    p = r['value']['properties']['property'][0]
    assert p['key'] == "String key"
    assert p['value']['string']['value'] == "String value"
    p = r['value']['properties']['property'][1]
    assert p['key'] == "Number key"
    assert p['value']['number']['value'] == 1234
    p = r['value']['properties']['property'][2]
    assert p['key'] == "Date key"
    assert p['value']['date']['value'] == "2021-05-04T07:33:33.773581268Z"


def test_list_string_data(client):
    """
    An all string list should be parsed as a vRO array of strings
    """

    data = get_data('List String')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'listString'
    assert r['type'] == 'Array/string'
    assert r['value']['array']['elements'][0]['string']['value'] == "Value 1"
    assert r['value']['array']['elements'][1]['string']['value'] == "Value 2"
    assert r['value']['array']['elements'][2]['string']['value'] == "foo"
    assert r['value']['array']['elements'][3]['string']['value'] == "bar"


def test_list_number_data(client):
    """
    An all number list should be parsed as a vRO array of numbers
    """
    
    data = get_data('List Number')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'listNumber'
    assert r['type'] == 'Array/number'
    assert r['value']['array']['elements'][0]['number']['value'] == 123
    assert r['value']['array']['elements'][1]['number']['value'] == -456
    assert r['value']['array']['elements'][2]['number']['value'] == 0
    assert r['value']['array']['elements'][3]['number']['value'] == 1.2


def test_list_mixed_data(client):
    """
    A list of differing datatypes should be parsed as a vRO generic Array
    """

    data = get_data('List Mixed')
    with client.application.app_context():
        r = getVroInputParam(data)

    assert r['name'] == 'listMixed'
    assert r['type'] == 'Array'
    assert r['value']['array']['elements'][0]['string']['value'] == "Value 1"
    assert r['value']['array']['elements'][1]['number']['value'] == 123
    assert r['value']['array']['elements'][2]['properties']['property'][0]['key'] == "foo"
    assert r['value']['array']['elements'][2]['properties']['property'][0]['value']['string']['value'] == "bar"
