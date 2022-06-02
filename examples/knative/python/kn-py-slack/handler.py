from flask import Flask, request, jsonify
import os
import requests
from cloudevents.http import from_http
import logging,json

logging.basicConfig(level=logging.DEBUG,format='%(asctime)s %(levelname)s %(name)s %(threadName)s : %(message)s')

app = Flask(__name__)
#Change the value to match the secret key in the VEBA appliance where you enter the Slack webook url information
url = os.environ.get('SLACK_SECRET')

@app.route('/', methods=['POST'])
def slack():
    
    try:
        event = from_http(request.headers, request.get_data(),None)
        
        data = event.data
        attrs = event._attributes
               
        #this section uses the Slack formatting to present the events in the format you would like.  You can modify as needed to add, remove or re-order the elements in a message
        payload = { "text": f"*CLOUDEVENT_ID*:\n  {attrs['id']}\n\n Source:  {attrs['source']}\n Type:  {attrs['type']}\n *Subject*:  *{attrs['subject']}*\n Time:  {attrs['time']}\n\n *EVENT DATA*:\n key:  {data['Key']}\n user:  {data['UserName']}\n datacenter:  {data['Datacenter']['Name']}\n Host:  {data['Host']['Name']}\n VM:  {data['Vm']['Name']}\n\n Message: {data['FullFormattedMessage']}" }  
                    
        requests.post(url=url, json=payload)
                    
        # app.logger.info(f'"***cloud event*** {json.dumps(e)}')
        return {}, 200
    
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
        msg = f'could not send message: {e}'
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
