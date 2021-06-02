from flask import Flask, request, jsonify
from cloudevents.http import from_http
import logging, json
from vcenter import Session
from datetime import date

logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s %(levelname)s %(name)s %(threadName)s : %(message)s'
)

app = Flask(__name__)
@app.route('/', methods=['POST'])
def handler():
    try:
        event = from_http(request.headers, request.get_data(),None)

        if event._attributes.get("datacontenttype").lower() != "application/json":
            sc = 400
            msg = f'invalid datacontenttype for cloud event: {event._attributes.get("datacontenttype")}'
            app.logger.error(msg)
            message = {
                'status': sc,
                'error': msg,
            }
            resp = jsonify(message)
            resp.status_code = sc
            return resp

        # CloudEvent - simple validation
        ref_vm = event.data['Vm']['Vm']
        ref_user = event.data['UserName']
        subject = event._attributes['subject']

        vc_s = Session()
        attr_owner, attr_creation_date, attr_last_poweredon = vc_s.get_field_attributes()
        vm_obj = vc_s.get_vm(ref_vm['Value'])
        if not vm_obj:
            sc = 404
            msg = f"could not find vm with moRef: {ref_vm['Value']}"
            app.logger.error(msg)
            message = {
                'status': sc,
                'error': msg,
            }
            resp = jsonify(message)
            resp.status_code = sc
            return resp

        if subject in ["DrsVmPoweredOnEvent", "VmPoweredOnEvent"]:
            app.logger.info(f"Apply attribute > {attr_last_poweredon.name}")
            vc_s.set_custom_attr(
                entity=vm_obj,
                key=attr_last_poweredon.key,
                value=date.today().strftime("%d/%m/%Y")
            )

        if subject in ["VmCreatedEvent", "VmClonedEvent", "VmRegisteredEvent"]:
            app.logger.info(f"Apply attribute > {attr_owner.name}")
            vc_s.set_custom_attr(
                entity=vm_obj,
                key=attr_owner.key,
                value=ref_user
            )

            app.logger.info(f"Apply attribute > {attr_creation_date.name}")
            vc_s.set_custom_attr(
                entity=vm_obj,
                key=attr_creation_date.key,
                value=date.today().strftime("%d/%m/%Y")
            )
        vc_s.close()
        return {}, 204
    except KeyError as e:
        sc = 400
        msg = f'could not decode cloud event: {e}'
        app.logger.error(msg)
        message = {
            'status': sc,
            'error': msg,
        }
        resp = jsonify(message)
        resp.status_code = sc
        return resp
    
    except Exception as e:
        sc = 500
        msg = f'could not apply attributes value: {e}'
        app.logger.error(msg)
        message = {
            'status': sc,
            'error': msg,
        }
        resp = jsonify(message)
        resp.status_code = sc
        return resp


# hint: run with FLASK_ENV=development FLASK_APP=handler.py flask run
if __name__ == "__main__":
    app.run()
