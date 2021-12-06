##
# These tests confirm the correct handling of input data and require 
# a working vRO appliance with valid credentials that can authenticate
# by OAUTH2
##

import json
from tests.functional.conftest import get_conf, get_data


def test_success(mock_env_config, client):
    rv = client.post('/', data=get_data('testevent.json'))
    assert rv.status_code == 202
    assert "Workflow Token" in rv.get_json()['message']
    assert "running" in rv.get_json()['message']


def test_no_data(mock_env_config, client):
    rv = client.post('/', data='')
    assert rv.status_code == 400


def test_bad_data(mock_env_config, client):
    rv = client.post('/', data='{"InvalidJson"}')
    assert rv.status_code == 400
    assert "The following can not be parsed as json:" in rv.get_json()['message']
    

def test_bad_workflow_id(monkeypatch, client):
    conf = get_conf()
    conf['WORKFLOW_ID'] = 'TestBadWfId'
    monkeypatch.setenv("VROCONFIG_SECRET", json.dumps(conf))
    rv = client.post('/', data=get_data('testevent.json'))
    assert rv.status_code == 404
    assert "vRO REST Request failed > Exception: Not Found" in rv.get_json()['message']
