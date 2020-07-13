# HA Restarted VMs Email Notification

## Description

This function demonstrates using PowerShell and PowerCLI to send an HTML formatted email notification containing a list of all VMs which were restarted due to a host failure in an HA enabled cluster.  The body of the email contains the VM names, time the VMs were restarted, and the contents of the VM notes/description field.

## Consume Function Instruction

Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/powercli/ha-restarted-vms
git checkout master
```

Step 2 - Update `stack.yml` and `vcconfig-ha-restarted-vms.json` with your environment information

<B>Note:</B> Be sure to configure all of the appropriate SMTP fields such as the SMTP_SERVER and SMTP_PORT for your particular SMTP server.  If authentication is not needed for your SMTP server, you can leave SMTP_USERNAME and SMTP_PASSWORD blank.

Step 3 - Login to the OpenFaaS gateway on vCenter Event Broker Appliance

```
VEBA_GATEWAY=https://veba.abc.com
export OPENFAAS_URL=${VEBA_GATEWAY}

faas-cli login --username admin --password-stdin --tls-no-verify
```

Step 4 - Create function secret (only required once)

```
faas-cli secret create vcconfig-ha-restarted-vms --from-file=vcconfig-ha-restarted-vms.json --tls-no-verify
```

Step 5 - Deploy function to vCenter Event Broker Appliance

```
faas-cli deploy -f stack.yml --tls-no-verify
```

## Build Function Instruction

Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/powercli/ha-restarted-vms
git checkout master
```

Step 2 - Verify the template/powercli directory and its contents exists - Needed for build

```
ls -R template/

template/:
powercli

template/powercli:
Dockerfile  function  template.yml

template/powercli/function:
script.ps1
```

Step 3 - Update `stack.yml` and `vcconfig-ha-restarted-vms.json` with your environment information. Please ensure you replace the name of the container image with your own account.

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
VEBA_GATEWAY=https://veba.abc.com
export OPENFAAS_URL=${VEBA_GATEWAY}

faas-cli login --username admin --password-stdin --tls-no-verify
```

Step 7 - Create function secret (only required once)

```
faas-cli secret create vcconfig-ha-restarted-vms --from-file=vcconfig-ha-restarted-vms.json --tls-no-verify
```

Step 8 - Deploy function to vCenter Event Broker Appliance

```
faas-cli deploy -f stack.yml --tls-no-verify
```