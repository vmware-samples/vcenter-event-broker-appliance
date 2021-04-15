# Function to make a POST request to any REST API

This function aims to provide a easy way to make HTTP Post request to an API Endpoint that might help with a lot of integration scenarios. Eg. Post to a Slack channel, Create a PagerDuty or ServiceNow incident. 

> **NOTE:** This function currently supports endpoints that allow basic authentication (un/pwd) or token based authentication that can be passed with the headers

- [Deploy](#deploy)
  * [Get the example function](#get-the-example-function)
  * [Customize the function](#customize-the-function)
    + [Understanding the Metaconfig-[SYSTEM].json](#understanding-the-metaconfig--system-json)
      - [Provide the API Details](#provide-the-api-details)
      - [Mapping the Events and Request body](#mapping-the-events-and-request-body)
    + [Updating the Stack.yml](#updating-the-stackyml)
    + [Updating the Handler.py (advanced)](#updating-the-handlerpy--advanced-)
  * [Deploy the function](#deploy-the-function)
    + [Create the secret](#create-the-secret)
    + [Build function (only if handler.py is changed)](#build-function--only-if-handlerpy-is-changed-)
    + [Deploy the function](#deploy-the-function-1)
  * [Trigger the function](#trigger-the-function)
- [Troubleshooting](#troubleshooting)

# Deploy

## Get the example function

Clone this repository which contains the example functions. 

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/python/invoke-rest-api
git checkout master
```

## Customize the function

There are three key files, that you might have to modify if you are looking to customize this function and make a post api call to an external system. 

```bash
  /invoke-rest-api/metaconfig-[SYSTEM].json #sample copies provided for PagerDuty, Slack, JIRA, Zendesk, ServiceDesk and ServiceNow
  /invoke-rest-api/stack.yml 
  /invoke-rest-api/handler/handler.py
```

### Understanding the Metaconfig-[SYSTEM].json 
First, change the configuration file `metaconfig-[SYSTEM].json` (samples for Slack, PagerDuty and ServiceNow are provided with this example) holding both sensitive and configuration information needed for this function in this folder. 

The `metaconfig-[SYSTEM].json` (as you can see below) has inputs needed to make a POST API request such as - `url, auth, headers` and `body`. These fields are self explanatory for an API and here are some of the considerations for this functions

#### Provide the API Details
* **url** - Http(s) endpoint to the System's API Endpoint. The function only does a POST call and this endpoint should be able to accept POST http request
* **auth** - Most APIs provide some sort of API key to use and don't seem to require authentication credentials. While some others require basic authentication which can be send in the header. Where the API requires username and password explicitly, i've added that as an option in the config. The function requires the presence of the auth key in the config (it can be an empty dict if the API does not require any username/password for authentication)
* **headers** - JSON key value pair of any headers that the API requires
* **body** - currently support only json body (which most APIs should support)

> **NOTE:** All these keys are required attributes. They can be empty if not required but if they are missing in the config then the function will fail. 


#### Mapping the Events and Request body
The `mapping` information provides a way to pull data from the Event that vCenter sends and replaces that within the request body before making the POST call.

* `"pull": "data/FullFormattedMessage"` attempts to get the value of `FullFormattedMessage` within the `data` dict of the cloud Event that we receive from VMware Event Router
* `"push": "payload/summary"` updates the `summary` field within the `payload` key of the request body as provided in this configuration file

> **Note:** This function has been developed to handle VMPoweredOn(/Off)Event by default, which you can see in the provided samples. Please edit the mapping for other Events accordingly.

```json
{
  "url": "https://events.pagerduty.com/v2/enqueue",
  "auth": {
    "un": "<username if required, leave blank if not required>",
    "pwd": "<username if required, leave blank if not required>"
  },
  "headers": {
    "Authorization":"Bearer <TOKEN>",
    .....
  },
  "body": {
    "payload": {
      "summary": "Example alert on host1.example.com",
      .....
    }
  },
  "mappings": [
    {
      "push": "payload/summary",
      "pull": "data/FullFormattedMessage"
    }
    .....
  ]
}
```

### Updating the Stack.yml
Function-specific settings are performed in the `stack.yml` file such as gateway, image, environment variables, secrets(configs) and the topics(events) that this function will subscribe to. Open and edit the `stack.yml` provided to change as per your environment/needs.

> **Note:** A key-value annotation under `topic` defines which VM event should trigger the function. A list of VM events from vCenter can be found [here](https://code.vmware.com/doc/preview?id=4206#/doc/vim.event.VmEvent.html). Multiple topics can be specified using a `","` delimiter syntax, e.g. "`topic: "VmPoweredOnEvent,VmPoweredOffEvent"`".

```yaml
provider:
  name: openfaas
  gateway: https://VEBA_FQDN_OR_IP              # replace with your vCenter Event Broker Appliance URL
functions:
  restpost-fn:
    lang: python3
    handler: ./handler
    image: vmware/veba-python-restpost:latest
    environment:
      write_debug: true                         # additional debugging messages are printed
      combine_output: false                     # required to prevent debug messages from showing up in faas response
      read_debug: true
      insecure_ssl: true                        # set to false if you have a trusted TLS certificate on VEBA 
    secrets:
      - metaconfig                              # leave as is, you will need to edit the function if this is changed
    annotations:
      topic: VmPoweredOnEvent,VmPoweredOffEvent # DrsVmPoweredOnEvent in a DRS-enabled cluster
```

> **Note:** If you are running a vSphere DRS-enabled cluster the topic annotation above should be `DrsVmPoweredOnEvent`. Otherwise the function would never be triggered.

### Updating the Handler.py (advanced)
You might have to edit this file if you are looking to possibly have multiple copies of this function running to make api calls to different system or to improve the function. 

To have multiple copies of this function running, you'll need multiple metaconfigs for each system and end up creating multiple secrets for each config. The `handler.py` for each function will have to be updated to reference their respective secret. You can do this by updating the below line in the file

```
META_CONFIG='/var/openfaas/secrets/metaconfig-[SYSTEM]'
```

For others that are looking to update the function and make improvements, please have at it! 

## Deploy the function

For the most part (if you didn't have to edit the handler.py), you'll have to create the secret and deploy the function. 

### Create the secret
Let's store the configuration file as secret in the appliance.

```bash
# set up faas-cli for first use
export OPENFAAS_URL=https://VEBA_FQDN_OR_IP
faas-cli login -p VEBA_OPENFAAS_PASSWORD --tls-no-verify # vCenter Event Broker Appliance is configured with authentication, pass in the password used during the vCenter Event Broker Appliance deployment process

# now create the secret
faas-cli secret create metaconfig --from-file=metaconfig-[SYSTEM].json --tls-no-verify
```

> **Note:** Delete the local `metaconfig-[SYSTEM].json` after you're done with this exercise to not expose any sensitive information.

### Build function (only if handler.py is changed)

> **NOTE:** This is only required if you changed the `handler.py` 

 Under the hoods, the functions are deployed as a container. Usually these containers are built and made readily available for you with the example function. However, when you make changes the function, you'll have to build the function and build the container. 

```bash 
faas-cli template pull

faas-cli build

faas-cli push #optional if you are pushing to DockerHub 
```

> **NOTE:** Make sure the `image` tag in the `stack.yml` is updated to reference the correct image. 

### Deploy the function
After you've performed the steps and modifications above, you can go ahead and deploy the function:

```bash
faas-cli deploy -f stack.yml --tls-no-verify
Deployed. 202 Accepted.
```

## Trigger the function

Turn on a virtual machine, e.g. in vCenter or via `govc` CLI, to trigger the function via a `(DRS)VmPoweredOnEvent`. Verify that the API was correctly called. 

> **Note:** If the API doesn't get called verify that you correctly followed each step above, IPs/FQDNs and credentials are correct and see the [troubleshooting](#troubleshooting) section below.

# Troubleshooting

If the API doesn't get called, verify:

- Validate that the API call works with Postman or cURL
- Validate the configurations provided within the `metaconfig-[SYSTEM].json`
- Validate the `stack.yml` file and the topic being subscribed to
- Verify if the components can talk to each other (VMware Event Router to vCenter and OpenFaaS, VMware Event Broker Appliance to API)
- Check the logs:

```bash
faas-cli logs restpost-fn --follow --tls-no-verify 

# Successful log message in the OpenFaaS tagging function
2019/01/25 23:48:55 Forking fprocess.
2019/01/25 23:48:55 Query
2019/01/25 23:48:55 Path  /
......
{"status": "200", "message": "Successfully executed the REST API call"}
2019/01/25 23:48:56 Duration: 1.551482 seconds
```