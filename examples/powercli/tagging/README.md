# vSphere Tagging Function

## Description

This function demonstrates using PowerCLI to apply vSphere Tag to Virtual Machine when the VM Powered On Event is triggered

## Instruction

Step 1 - Initialize function, only required during the first deployment

```
faas-cli template pull
```

Step 2 - Update `stack.yml` and `vcconfig.json` with your environment information. TAG_FILTER lets you tag only VMs with a specific tag alread on them. For example, if TAG_FILTER is set to 'tag-on-power', only VMs that are already tagged with the tag 'tag-on-power' will be tagged with TAG_NAME when powered on.

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
faas-cli secret create vcconfig --from-file=vcconfig.json --tls-no-verify # create secret, only required once
faas-cli deploy -f stack.yml --tls-no-verify
```