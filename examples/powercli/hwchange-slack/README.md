# PowerCLI Function Example

## Description

This function demonstrates using PowerCLI to send VM configuration changes to Slack when the VM Reconfigure Event is triggered

There is a blog post covering this example in detail: [Audit VM configuration changes using the vCenter Event Broker
](https://www.opvizor.com/audit-vm-configuration-changes-using-the-vcenter-event-broker)

The custom PowerShell template for OpenFaaS is using [PSSlack](https://github.com/RamblingCookieMonster/PSSlack)

## Instruction

Step 1 - Setup Slack

Make sure to create a channel for the notifications and a [Slack webhook](https://my.slack.com/services/new/incoming-webhook/).


Step 2 - Update `stack.yml` and `vcconfig.json` with your enviornment information

`stack.yml` **lines: gateway, image**

```
provider:
  name: openfaas
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

`vcconfig.json`

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

