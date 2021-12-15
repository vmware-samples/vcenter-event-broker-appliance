from flask import Flask
from kn_py_vro.handler import bp as handler_bp


def create_app(test_config=None):
    app = Flask(__name__)
    
    if test_config is not None:
        app.config.update(test_config)
    
    app.register_blueprint(handler_bp)

    return app