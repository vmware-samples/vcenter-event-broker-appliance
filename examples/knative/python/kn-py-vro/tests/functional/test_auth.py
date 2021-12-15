##
# These tests require a working vRO appliance with valid
# credentials that can authenticate by both HTTP Basic Auth
# and OAUTH2
##

import pytest
import json
from tests.functional.conftest import get_conf, get_data
from werkzeug.exceptions import InternalServerError
    
    
def test_no_config(client):
    with pytest.raises(InternalServerError) as excinfo:
        rv = client.post('/')
    assert "Mandatory configuration key not found: 'VROCONFIG_SECRET'" in str(excinfo.value)


def test_missing_username(monkeypatch, client):
    conf = get_conf()
    del conf['USERNAME']
    monkeypatch.setenv("VROCONFIG_SECRET", json.dumps(conf))
    with pytest.raises(InternalServerError) as excinfo:
        rv = client.post('/', data=get_data('testevent.json'))
    #assert rv.status_code == 500
    assert "Mandatory configuration key not found: 'USERNAME'" in str(excinfo.value) 


def test_bad_username(monkeypatch, client):
    conf = get_conf()
    conf['USERNAME'] = 'TestBadUserName'
    monkeypatch.setenv("VROCONFIG_SECRET", json.dumps(conf))
    rv = client.post('/', data=get_data('testevent.json'))
    assert rv.status_code == 500
    assert "Error with authentication request:" in str(rv.data)


def test_basic_auth(monkeypatch, client):
    conf = get_conf()
    conf['AUTH_TYPE'] = 'BASIC'
    monkeypatch.setenv("VROCONFIG_SECRET", json.dumps(conf))
    rv = client.post('/', data=get_data('testevent.json'))