# Disable Alarms for Host Maintenance

## Description

This example will disable alarm actions on a host while it is in maintenance mode.  It deploys two functions that use the same PowerCLI script.  The first function subscribes to the `entered.maintenance.mode` event to run when a host is put into maintenance mode and disable alarms.  The second function subscribes to the `exit.maintenance.mode` event to re-enable alarms when the host exits maintenance mode.  There is an accompanying blog post with more details:  [Automate Host Maintenance with the vCenter Event Broker Appliance
](https://doogleit.github.io/2019/11/automate-host-maintenance-with-the-vcenter-event-broker-appliance/)

## Consume Function Instruction

Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/powercli/hostmaint-alarms
git checkout master
```

Step 2 - Update `stack.yml` and `vc-hostmaint-config.json` with your environment information

Step 3 - Login to the OpenFaaS gateway on vCenter Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY}

faas-cli login --username admin --password-stdin --tls-no-verify
```

Step 4 - Create function secret (only required once)

```
faas-cli secret create vc-hostmaint-config --from-file=vc-hostmaint-config.json --tls-no-verify
```

Step 5 - Deploy function to vCenter Event Broker Appliance

```
faas-cli deploy -f stack.yml --tls-no-verify
```

## Build Function Instruction

Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/powercli/hostmaint-alarms
git checkout master
```

Step 2 - Initialize function, only required during the first deployment

```
faas-cli template pull
```

Step 3 - Update `stack.yml` and `vc-hostmaint-config.json` with your environment information. Please ensure you replace the name of the container image with your own account.

Step 4 - Build the function container

```
faas-cli build -f stack.yml
```

Step 5 - Push the function container to Docker Registry (default but can be changed to internal registry)

```
faas-cli push -f stack.yml
```

Step 6 - Login to the OpenFaaS gateway on vCenter Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY}

faas-cli login --username admin --password-stdin --tls-no-verify
```

Step 7 - Create function secret (only required once)

```
faas-cli secret create vc-hostmaint-config --from-file=vc-hostmaint-config.json --tls-no-verify
```

Step 8 - Deploy function to vCenter Event Broker Appliance

```
faas-cli deploy -f stack.yml --tls-no-verify
```