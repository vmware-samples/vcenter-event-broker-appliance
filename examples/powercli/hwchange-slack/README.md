# Send VM Configuration Changes to Slack

## Description

This function demonstrates using PowerCLI to send VM configuration changes to Slack when the VM Reconfigure Event is triggered

There is a blog post covering this example in detail: [Audit VM configuration changes using the vCenter Event Broker
](https://www.opvizor.com/audit-vm-configuration-changes-using-the-vcenter-event-broker)

The custom PowerShell template for OpenFaaS is using [PSSlack](https://github.com/RamblingCookieMonster/PSSlack)

## Consume Function Instruction

Step 1 - Setup Slack

Make sure to create a channel for the notifications and a [Slack webhook](https://my.slack.com/services/new/incoming-webhook/).


Step 2 - Update `stack.yml` and `vc-slack-config.json` with your environment information

Step 3 - Login to the OpenFaaS gateway on vCenter Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY}

faas-cli login --username admin --password-stdin --tls-no-verify
```

Step 4 - Create function secret (only required once)

```
faas-cli secret create vc-slack-config --from-file=vc-slack-config.json --tls-no-verify
```

Step 5 - Deploy function to vCenter Event Broker Appliance

```
faas-cli deploy -f stack.yml --tls-no-verify
```

## Build Function Instruction

Step 1 - Initialize function, only required during the first deployment

```
faas-cli template pull
```

Step 2 - Update `stack.yml` and `vc-slack-config.json` with your environment information. Please ensure you replace the name of the container image with your own account.

Step 3 - Build the function container

```
faas-cli build -f stack.yml
```

Step 4 - Push the function container to Docker Registry (default but can be changed to internal registry)

```
faas-cli push -f stack.yml
```

Step 5 - Login to the OpenFaaS gateway on vCenter Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY}

faas-cli login --username admin --password-stdin --tls-no-verify
```

Step 6 - Create function secret (only required once)

```
faas-cli secret create vc-slack-config --from-file=vc-slack-config.json --tls-no-verify
```

Step 7 - Deploy function to vCenter Event Broker Appliance

```
faas-cli deploy -f stack.yml --tls-no-verify
```