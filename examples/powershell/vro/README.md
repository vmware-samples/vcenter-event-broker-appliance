# vRealize Orchestrator Function

## Description

This function demonstrates using PowerShell to trigger vRealize Orchestrator workflow using vRO REST API

## Prerequisites

* You have deployed the example vSphere Tagging vRO Workflow package from https://github.com/kclinden/vro-vsphere-tagging
* You have retrieved the required vRO Workflow ID (please see this blog post [here](https://www.virtuallyghetto.com/2020/03/using-vro-rest-api-to-execute-a-workflow-with-sdk-objects.html) for more details)

## Instruction Consuming Function

Step 1 - Initialize function, only required during the first deployment

```
faas-cli template pull
```

Step 2 - Update `stack.yml` and `vro-secrets.json` with your environment information

> **Note:** If you are building your own function, you will need to update the `image:` property in the stack.yaml to point to your own Dockerhub account and Docker Image (e.g. `<dockerhubid>/<dockerimagename>`)

Step 3 - Deploy function to VMware Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli login --username admin --password-stdin --tls-no-verify # login with your admin password
faas-cli secret create vro-secrets --from-file=vro-secrets.json --tls-no-verify # create secret, only required once
faas-cli deploy -f stack.yml --tls-no-verify
```

Step 4 - To remove the function and secret from VMware Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli remove -f stack.yml --tls-no-verify
faas-cli secret remove vro-secrets --tls-no-verify
```

## Instruction Building Function

Follow Step 1 from above and then any changes made to your function, you will need to run these additional two steps before proceeding to Step 2 from above.

Step 1 - Build the function container

```
faas-cli build -f stack.yml
```

Step 2 - Push the function container to Docker Registry (default but can be changed to internal registry)

```
faas-cli push -f stack.yml
```
