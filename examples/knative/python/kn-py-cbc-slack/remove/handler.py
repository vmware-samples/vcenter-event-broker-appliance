from flask import Flask, request, jsonify
from cloudevents.http import from_http
from cbc_sdk import CBCloudAPI
from cbc_sdk.platform import Device
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
    try:
        workloadApi = CBCloudAPI()
    except:
        return "Could not establish authenticated connection to Carbon Black Cloud", 400
    app.logger.debug(f"Carbon Black Cloud API {workloadApi}")

    # Setup Slack comms
    url = os.environ['SLACK_URL']
    slackWebhook = WebhookClient(url)

    # Search CBC workload API for VM which event relates to
    vmName = str(data['Vm']['Name'])
    app.logger.debug(f"Virtual Machine name to search for is {vmName}")

    try:
        cbcComputeResourceQuery = workloadApi.select(Device).where(name=vmName) 
    except:
        return "Could not form API query using the Virtual Server Name", 400
    app.logger.debug(f"CBC compute resource query object is {cbcComputeResourceQuery}")

    try:
        for vm in cbcComputeResourceQuery:
            try:
                workloadApi.select(Device, vm.id).uninstall_sensor()
            except:
                return "Call to Uninstall Sensor Failed", 400
            time.sleep(5)
            try:
                workloadApi.select(Device, vm.id).delete_sensor()
            except:
                return "Call to De-register Sensor Failed", 400
            try:
                slackResponse = slackWebhook.send(text=f":siren: VM Removed from CBC :siren: \n *CLOUDEVENT_ID*: \n  {attrs['id']}\n\n Source:  {attrs['source']}\n Type:  {attrs['type']}\n *Subject*:  *{attrs['subject']}*\n Time:  {attrs['time']}\n\n *EVENT DATA*:\n key:  {data['Key']}\n user:  {data['UserName']}\n datacenter:  {data['Datacenter']['Name']}\n Host:  {data['Host']['Name']}\n VM:  {data['Vm']['Name']}\n\n Message: {data['FullFormattedMessage']}")
            except:
                return "Request to post Slack message failed", 400
        return "", 204
    except:
        return "Could not execute API query using the Virtual Server Name", 400

if __name__ == "__main__":
    app.run()
