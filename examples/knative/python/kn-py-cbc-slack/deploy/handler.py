from flask import Flask, request, jsonify
from cloudevents.http import from_http
from cbc_sdk import CBCloudAPI
from cbc_sdk.workload.vm_workloads_search import ComputeResource
from slack_sdk.webhook import WebhookClient
import logging,json,os,time
logging.basicConfig(level=logging.INFO,format='%(asctime)s %(levelname)s %(name)s %(threadName)s : %(message)s')

app = Flask(__name__)
@app.route("/", methods=["POST"])

def home():
 # Extract  VM Name From CloudEvent Data
    event = from_http(request.headers, request.get_data(),None)
    app.logger.debug(f"Full event contents {event}")
    data  = event.data
    attrs = event._attributes
    app.logger.info(f"Found event ID {event['id']} triggered by {event['subject']} event for VM {data['Vm']['Name']}")

    #Output Carbon Black Cloud SDK Authentication Environment Variables
    app.logger.debug(f"Environment variable CBC_URL value is " + os.environ['CBC_URL'])
    app.logger.debug(f"Environment variable CBC_TOKEN value is " + os.environ['CBC_TOKEN'])
    app.logger.debug(f"Environment variable CBC_ORG_KEY value is " + os.environ['CBC_ORG_KEY'])
    #Output Sensor Version Environment Variable
    app.logger.debug(f"Environment variable SENSOR_VER value is " + os.environ['SENSOR_VER'])
    #Output Slack WebHook URL Environment Variable
    app.logger.debug(f"Environment variable SLACK_URL value is " + os.environ['SLACK_URL'])

    # Establish SDK session to workload API
    workloadApi = CBCloudAPI()
    app.logger.debug(f"Carbon Black Cloud API {workloadApi}")
    
    # Setup Slack comms 
    url = os.environ['SLACK_URL']
    slackWebhook = WebhookClient(url)

    # Search CBC workload API for VM which event relates to
    vmName = str(data['Vm']['Name'])
    app.logger.debug(f"Virtual Machine name to search for is {vmName}")

    cbcComputeResourceQuery = workloadApi.select(ComputeResource).set_name([vmName])
    app.logger.debug(f"CBC compute resource query object is {cbcComputeResourceQuery}")

    # Get VM objectc from CBC instane, if not immediatly available wait and retry
    tries = 5
    while tries >=0:
        for vm in cbcComputeResourceQuery:
            app.logger.debug(f"CBC compute resource object ID is " + vm.id)
            if vm.id is None or tries == 0:
                app.logger.debug(f"Resource not found in CBS instance,  retrying")
                time.sleep(5)
                tries -= 1
            else:
                app.logger.debug(f"Found VM in CBS instance")
                tries = 0
    if vm.id is None:
        app.logger.debug(f"VM not found in CBS instance")
        return "", 404

    # Query VM objet sensor install eligibility status, if not immediatly eligible wait and retry
    tries = 5
    while tries >=0:
      for vm in cbcComputeResourceQuery:
        if vm.eligibility == 'ELIGIBLE':
            app.logger.debug(f"VM is eligible for sensor installation")
            workloadApi.select(ComputeResource,vm.id).install_sensor(os.environ['SENSOR_VER'],config_file=os.environ['CBC_CONFIG_INI'])
            app.logger.debug(f"Attempting installation of sensors on VM")
            slackResponse = slackWebhook.send(text=f":carbonblack: New VM - Carbon Black Installed :carbonblack: \n *CLOUDEVENT_ID*: \n {attrs['id']}\n\n Source:  {attrs['source']}\n Type:  {attrs['type']}\n *Subject*:  *{attrs['subject']}*\n Time:  {attrs['time']}\n\n *EVENT DATA*:\n key:  {data['Key']}\n user:  {data['UserName']}\n datacenter:  {data['Datacenter']['Name']}\n Host:  {data['Host']['Name']}\n VM:  {data['Vm']['Name']}\n\n Message: {data['FullFormattedMessage']}")
            return "", 202
        else:
            app.logger.debug(f"VM not eligible for sensor installation,  retrying " + vm.eligibility)
            time.sleep(30)
            tries -= 1

if __name__ == "__main__":
    app.run()

