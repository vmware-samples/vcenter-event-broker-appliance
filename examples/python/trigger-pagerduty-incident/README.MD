### Get the example function

Clone this repository which contains the example functions. 

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/python/trigger-pagerduty-incident
git checkout master
```

### Customize the function

For security reasons to not expose sensitive data we will create a Kubernetes [secret](https://kubernetes.io/docs/concepts/configuration/secret/) which will hold the routingkey information for the PagerDuty Event API v2. This secret will be mounted into the function during runtime and is taken care of by the appliance. We only have to create the secret with a simple command through `faas-cli`.

First, change the configuration file [pdconfig.json](pdconfig.json) holding your sensitive pagerduty information in this folder:

```json
{
    "routing_key": "<replace with your routing key>",
    "event_action": "trigger" 
}
```

Now go ahead and store this configuration file as secret in the appliance.

```bash
# set up faas-cli for first use
export OPENFAAS_URL=https://VEBA_FQDN_OR_IP
faas-cli login -p VEBA_OPENFAAS_PASSWORD --tls-no-verify # vCenter Event Broker Appliance is configured with authentication, pass in the password used during the vCenter Event Broker Appliance deployment process

# now create the secret
faas-cli secret create pdconfig --from-file=pdconfig.json --tls-no-verify
```

> **Note:** Delete the local `pdconfig.json` after you're done with this exercise to not expose this sensitive information.

Lastly, define the vCenter event which will trigger this function. Such function-specific settings are performed in the `stack.yml` file. Open and edit the `stack.yml` provided with in the Python example code. Change `gateway` and `topic` as per your environment/needs.

> **Note:** A key-value annotation under `topic` defines which VM event should trigger the function. A list of VM events from vCenter can be found [here](https://code.vmware.com/doc/preview?id=4206#/doc/vim.event.VmEvent.html). Multiple topics can be specified using a `","` delimiter syntax, e.g. "`topic: "VmPoweredOnEvent,VmPoweredOffEvent"`".

```yaml
provider:
  name: openfaas
  gateway: https://VEBA_FQDN_OR_IP # replace with your VMware Event Broker Appliance environment
functions:
  pdinvoke-fn:
    lang: python3
    handler: ./handler
    image: vmware/veba-pagerduty-invoke:latest
    environment:
      write_debug: true #function writes verbose entries to the log when set to true, also requires combine_output to be set to false to avoid debug messages from showing up in the response
      read_debug: true
      combine_output: false #prevents error logs from showing up on the response output
      insecure_ssl: true #set to true if you have a trusted TLS certificate on the gateway
    secrets:
      - pdconfig # update file with your Pagerduty integration key - https://v2.developer.pagerduty.com/docs/send-an-event-events-api-v2
    annotations:
      topic: VmPoweredOnEvent,VmPoweredOffEvent # or DrsVmPoweredOnEvent in a DRS-enabled cluster
```

> **Note:** If you are running a vSphere DRS-enabled cluster the topic annotation above should be `DrsVmPoweredOnEvent`. Otherwise the function would never be triggered.

### Deploy the function

After you've performed the steps and modifications above, you can go ahead and deploy the function:

```bash
faas-cli template pull # only required during the first deployment
faas-cli deploy -f stack.yml --tls-no-verify
Deployed. 202 Accepted.
```

### Trigger the function

Turn on a virtual machine, e.g. in vCenter or via `govc` CLI, to trigger the function via a `(DRS)VmPoweredOnEvent`. You should now receive a PagerDuty notification through the configured notification channel

> **Note:** If you are not seeing the PagerDuty alert upon the event being triggered, validate that the routing key in the `pdconfig.json` is correct and see the [troubleshooting](#troubleshooting) section below.

## Troubleshooting

If your PagerDuty event is not getting invoked, verify:

- Routing Key in the `pdconfig.json`
- Whether the components can talk to each other (VMware Event Router to vCenter and OpenFaaS, VMware Event Broker Appliance to PagerDuty)
- If you have changed the `topic` in `stack.yml`, please ensure that the Function is also updated to handle the expected event data. 
- Check the logs:

```bash
faas-cli logs pdinvoke-fn --follow --tls-no-verify 

# Successful log message in the OpenFaaS function
2019/01/25 23:48:55 Forking fprocess.
2019/01/25 23:48:55 Query
2019/01/25 23:48:55 Path  /

{"status": "200", "message": "successfully triggered event action with dedup_key: <redacted>"}
2019/01/25 23:48:56 Duration: 1.551482 seconds
```
