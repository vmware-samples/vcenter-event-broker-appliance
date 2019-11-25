# PowerCLI Function Example

## Description

This function demonstrates using PowerCLI to send VM configuration changes to Slack when the VM Reconfigure Event is triggered.

Make sure to use the template in this repo as it builds a container image containing [PSSlack](https://github.com/RamblingCookieMonster/PSSlack) 

You can also find a blog post that includes all steps here: 

[Audit VM Configuration changes using the vCenter Event Broker Appliance](https://www.opvizor.com/audit-vm-configuration-changes-using-the-vcenter-event-broker)


## Instruction

Step 1 - Create Slack channel and get a Slack webhook for that

[Slack webhook](https://my.slack.com/services/new/incoming-webhook/)


Step 2 - Update `stack.yml` and `vcconfig.json` with your environment information

**stack.yml - gateway, image**
```
provider:
  name: faas
  gateway: https://veba.mynetwork.local
functions:
  powercli-reconfigure:
    lang: powercli
    handler: ./handler
    image: opvizorpa/powercli-slack:latest
    environment:
      write_debug: true
      read_debug: true
      function_debug: false
    secrets:
      - vcconfig
    annotations:
      topic: vm.reconfigured
```

**vcconfig.json**
```
{
    "VC" : "my-vCenter",
    "VC_USERNAME" : "user@vsphere.local",
    "VC_PASSWORD" : "userpassword",
    "SLACK_URL"   : "https://my.slack.com/services/new/incoming-webhook/",
    "SLACK_CHANNEL" : "vcevent"
}

```


Step 3 - Build the function container

```
faas-cli build -f stack.yml
```

Step 4 - Push the function container to Docker Registry (default but can be changed to internal registry)

```
faas-cli push -f stack.yml
```

Step 5 - Deploy function to VEBA

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli login --username admin --password-stdin --tls-no-verify # login with your admin password
faas-cli secret create vcconfig --from-file=vcconfig.json --tls-no-verify # create secret, only required once
faas-cli deploy -f stack.yml --tls-no-verify
```
