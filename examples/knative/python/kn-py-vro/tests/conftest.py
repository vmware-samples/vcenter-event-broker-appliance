##
# Root conftest
##

import pytest
from kn_py_vro import create_app


@pytest.fixture
def client():
    app = create_app({'TESTING': True, 'FLASK_ENV': True})
    with app.test_client() as client:
        yield client