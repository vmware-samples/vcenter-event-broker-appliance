# Create a Salesforce Case

This function helps create a Salesforce Case when triggered by a vCenter Event. Please go through the steps below to deploy or to customize and build this function. 

## Deploy function (without modification)

Step 1 - Clone repository

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/python/create-salesforce-case
git checkout master
```

Step 2 - View the `stack.yml` and modify file as required

`stack.yml` file is the main descriptor that contains the function-specific settings. Open and edit the `stack.yml`. 

```yaml
version: 1.0
provider:
  name: openfaas                                  
  gateway: https://https://VEBA_FQDN_OR_IP # OpenFaaS URL (VEBA endpoint) 
functions:
  salesforce-fn:                           # name of the function (unique in your k8s cluster)
    lang: python3
    handler: ./handler                     # folder name which has your scripts
    image: vmware/veba-python-sfdc:latest  # docker container to pull for this deployment
    environment:
      write_debug: true
      read_debug: true
      combine_output: false                # env variable allows us to use syserr to print to console while sysout is returned back to OpenFaaS
      insecure_ssl: true                   # env variable to let us work with self signed certs
    secrets:
      - sfconfig                           # configuration for the function
    annotations:
      topic: "VmPoweredOffEvent"           # comma separated events that trigger this fn
```

Step 3 - View and edit the `sfconfig.json` 

We will use the [Kubernetes secrets](https://kubernetes.io/docs/concepts/configuration/secret/){:target="_blank"} to hold our Salesforce configuration. This secret will be mounted into the function during runtime automatically by the appliance. We only have to create the secret with a simple command through `faas-cli`.

This function takes a lot of inspiration from the Rest API POST function and reuses similar concepts. Please refer this [document](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/development/examples/python/invoke-rest-api#provide-the-api-details) to get yourself familiar with how the config is structured. Please take note of the additional fields within the `auth` parameter as Salesforce requires oAuth based flow which is added here. 

At a minimum, you'll need to provide 
* **URL** - update the URL with your salesforce instance
* **AUTH** - this function supports [un/pwd oauth flow with Salesforce](https://help.salesforce.com/articleView?id=remoteaccess_oauth_username_password_flow.htm&type=0). Please provide all the required fields. 
* **BODY** - to have the most accurate information on the ticket, you'll need to provide the ContactId, AccountId, OwnerId as indicated in the `sfconfig.json` 
* **MAPPING** - mapping is optional but recommended if you would like to pull information from the event to be used when creating the case. 

```json
{
    "url": "https://instance.my.salesforce.com/services/data/v48.0/sobjects/Case/",
    "headers": {
        "Content-Type": "application/json; charset=UTF-8",
        "Accept": "application/json; charset=UTF-8"
    },
    "auth": {
        "un": "username_required",
        "pwd": "password_required",
        "client_id": "clientid_required",
        "client_secret": "clientsecret_required",
        "redirect_uri": "https://OPENFAAS_URL"
    },
    "body": {
        "ContactId": "0034T00000BoZxKQAV",
        [....]
    },
    "mappings": [
        {
            "push": "Subject",
            "pull": "data/FullFormattedMessage"
        },
        [....]
    ]
}
```

Step 4 - Login to the OpenFaaS gateway on vCenter Event Broker Appliance and create the secret (the configuration)

```bash
export OPENFAAS_URL=https://veba-fqdn

faas-cli login --username admin --password-stdin --tls-no-verify

# now create the secret
faas-cli secret create sfconfig--from-file=sfconfig.json --tls-no-verify
```

Step 5 - Deploy function to vCenter Event Broker Appliance

```bash
faas-cli deploy -f stack.yml --tls-no-verify
```

Step 6 - Tail the logs of the salesforce-fn function

```bash 
faas-cli logs salesforce-fn --tls-no-verify
```

Step 7 - Trigger the vCenter Event and you should see output like the following in the console:

```bash
#show your function output here
```

## Customize the function

Step 1 - Change the configuration file [sfconfig.json](sfconfig.json) in this folder:

Please refer this [document](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/development/examples/python/invoke-rest-api#provide-the-api-details) to get yourself familiar with how the config is structured. This function does oAuth based authentication with Salesforce and then attempt to make an API call to Salesforce with the Authorization token. You could modify the `sfconfig.json` to attempt other Salesforce API requests. 

```json
{
    "url": "https://instance.my.salesforce.com/services/data/v48.0/sobjects/Case/",
    "headers": {
        "Content-Type": "application/json; charset=UTF-8",
        "Accept": "application/json; charset=UTF-8"
    },
    "auth": {
        "un": "username_required",
        "pwd": "password_required",
        "client_id": "clientid_required",
        "client_secret": "clientsecret_required",
        "redirect_uri": "https://OPENFAAS_URL"
    },
    "body": {
        "ContactId": "0034T00000BoZxKQAV",
        [....]
    },
    "mappings": [
        {
            "push": "Subject",
            "pull": "data/FullFormattedMessage"
        },
        [....]
    ]
}
```

Step 2 - Go ahead and store this configuration file as secret in the appliance.

```bash
# set up faas-cli for first use
export OPENFAAS_URL=https://VEBA_FQDN_OR_IP

faas-cli login -p VEBA_OPENFAAS_PASSWORD --tls-no-verify 

# now create the secret
faas-cli secret create sfconfig--from-file=sfconfig.json --tls-no-verify
```

> **Note:** Delete the local `sfconfig.json` after you're done with this exercise to not expose this sensitive information.

Step 3 - Edit the `stack.yml`

Function-specific settings are performed in the `stack.yml` file such as gateway, image, environment variables, secrets(configs) and the topics(events) that this function will subscribe to. Open and edit the `stack.yml` provided to change as per your environment/needs. If you are looking to customize the function even further, you are possibly looking to have the function be triggered for a different `topic`.  If you are changing the topic, validate that the `mapping` within the `sfconfig.json` is updated as well. 

> **Note:** A key-value annotation under `topic` defines which VM event should trigger the function. A list of VM events from vCenter can be found [here](https://code.vmware.com/doc/preview?id=4206#/doc/vim.event.VmEvent.html). Multiple topics can be specified using a `","` delimiter syntax, e.g. "`topic: "VmPoweredOnEvent,VmPoweredOffEvent"`".

Step 4 - Edit the `handler.py` and `requirements.txt`

This function uses [the username and password flow with Salesforce](https://help.salesforce.com/articleView?id=remoteaccess_oauth_username_password_flow.htm&type=0) and you might want to edit this to support other type of oauth requests. The `handler.py` file has sufficient call outs that will guide you through the process. 

If you have any library that is needed, add those to the requirements.txt file. 

### Deploy the function

After you've performed the steps and modifications above, you can go ahead and deploy the function:

```bash
faas-cli template pull # only required during the first deployment
faas-cli up --tls-no-verify
Deployed. 202 Accepted.
```

### Trigger the function

This function has been tested with VMPowerOn/Off Event. Please power on a VM to trigger the function via a `VMPowerOn/OffEvent`. You should see a Salesforce Case automatically getting created with the information set as per the mapping configuration provided. 

## Troubleshooting the function

If your function is not working as intended, tail the logs of the salesforce-fn function using faas-cli as show below. 

```
faas-cli logs salesforce-fn --tls-no-verify
```

If your function is not getting invoked, verify:

- The configuration within the `sfconfig.json`
- Whether the components can talk to each other (VMware Event Router to vCenter and OpenFaaS, VMware Event Broker Appliance to PagerDuty)
- If you have changed the `topic` in `stack.yml`, please ensure that the Function is also updated to handle the expected event data. 
- Check the logs:

```bash
# Successful log message in the OpenFaaS function
2019/01/25 23:48:55 Forking fprocess.
2019/01/25 23:48:55 Query
2019/01/25 23:48:55 Path  /

{"status": "200", "message": "Success: <response redacted>"}
2019/01/25 23:48:56 Duration: 1.551482 seconds
```
