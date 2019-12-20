# vSphere Tagging Function

## Description

This function demonstrates using PowerCLI to apply vSphere Tag to Virtual Machine when the VM Powered On Event is triggered

## Instruction

Step 1 -  Clone this repository which contains the example functions.

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
git checkout development
cd vcenter-event-broker-appliance/examples/examples/powercli/tagging
```

Step 2 - Initialize function, only required during the first deployment

```
faas-cli template pull
```

Step 3 - Update `stack.yml` and `vcconfig.json` with your environment information

Step 4 - Build the function container

```
faas-cli build -f stack.yml
```

Step 5 - Push the function container to Docker Registry (default but can be changed to internal registry)

```
faas-cli push -f stack.yml
```

Step 6 - Deploy function to vCenter Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli login --username admin --password-stdin --tls-no-verify # login with your admin password
faas-cli secret create vcconfig --from-file=vcconfig.json --tls-no-verify # create secret, only required once
faas-cli deploy -f stack.yml --tls-no-verify
```