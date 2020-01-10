# AWS EventBridge Function

## Description

This function demonstrates using PowerShell and the EventBridge cmdlets to forward a vCenter Server Event to AWS EventBridge

## Prerequisites

* Already created custom [EventBridge Bus](https://docs.aws.amazon.com/eventbridge/latest/userguide/create-event-bus.html) and [EventBridge Rule](https://docs.aws.amazon.com/eventbridge/latest/userguide/create-rule-partner-events.html)
* AWS Secret and Access Key with `AmazonEventBridgeFullAccess` policy

## Instruction

Step 1 - Initialize function, only required during the first deployment

```
faas-cli template pull
```

Step 2 - Update `stack.yml` and `eventbridge-secrets.json` with your environment information

Step 3 - Build the function container

```
faas-cli build -f stack.yml
```

Step 4 - Push the function container to Docker Registry (default but can be changed to internal registry)

```
faas-cli push -f stack.yml
```

Step 5 - Deploy function to vCenter Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli login --username admin --password-stdin --tls-no-verify # login with your admin password
faas-cli secret create eventbridge-secrets --from-file=eventbridge-secrets.json --tls-no-verify # create secret, only required once
faas-cli deploy -f stack.yml --tls-no-verify
```

Step 6 - To remove the function and secret from vCenter Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli remove -f stack.yml --tls-no-verify
faas-cli secret remove eventbridge-secrets --tls-no-verify
```