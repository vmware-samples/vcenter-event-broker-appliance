##
# Functional conftest
##

import pytest
import json


@pytest.fixture
def mock_env_config(monkeypatch):
    config = json.dumps(get_conf())
    monkeypatch.setenv("VROCONFIG_SECRET", config)    


def get_conf():
    file = open("./kn-py-vro_secret.json")
    config = file.read()
    file.close()
    return json.loads(config)


def get_data(name):
    file = open(f'tests/{name}')
    data = file.read()
    file.close()
    return data