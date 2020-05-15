# vSphere Datastore Usage Email Notification

## Description

This function demonstrates using PowerShell to send an email notification when warning/error threshold is reach for Datastore Usage Alarm in vSphere

## Consume Function Instruction

Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/powercli/datastore-usage-email
git checkout master
```

Step 2 - Update `stack.yml` and `vc-datastore-config.json` with your environment information

> **Note:**
 Leave SMTP_USERNAME and SMTP_PASSWORD blank if you do not want to use authenticated SMTP

The function supports pulling a To: email address from a custom attribute in vCenter. This allows administrators with vCenter access to configure email notifications without having to alter the JSON script configuration. To enable this feature, do the following:
- Create a custom attribute in vCenter, assign it to a datastore and give it an email address. For example, create a custom attribute named `NotifyEmail` and assign it a value of `admin@foo.com`
- In `vc-datastore-config.json`, add the name of the custom attribute `NotifyEmail` as the value for `DATASTORE_CUSTOM_PROP_EMAIL_TO`
- Add a vCenter URL and credentials to `VC`, `VC_USERNAME`, and `VC_PASSWORD`

When this is configured, an email address found in the custom attribute will be added to the email already defined in the `EMAIL_TO` section of the JSON file. 

Step 3 - Login to the OpenFaaS gateway on VMware Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY}

faas-cli login --username admin --password-stdin --tls-no-verify
```

Step 4 - Create function secret (only required once)

```
faas-cli secret create vc-datastore-config --from-file=vc-datastore-config.json --tls-no-verify
```

Step 5 - Deploy function to VMware Event Broker Appliance

```
faas-cli deploy -f stack.yml --tls-no-verify
```

## Build Function Instruction

Step 1 - Clone repo

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/powercli/datastore-usage-email
git checkout master
```

Step 2 - Initialize function, only required during the first deployment

```
faas-cli template pull
```

Step 3 - Update `stack.yml` and `vc-datastore-config.json` with your environment information. Please ensure you replace the name of the container image with your own account.

Step 4 - Build the function container

```
faas-cli build -f stack.yml
```

Step 5 - Push the function container to Docker Registry (default but can be changed to internal registry)

```
faas-cli push -f stack.yml
```

Step 6 - Login to the OpenFaaS gateway on VMware Event Broker Appliance

```
VEBA_GATEWAY=https://veba.primp-industries.com
export OPENFAAS_URL=${VEBA_GATEWAY}

faas-cli login --username admin --password-stdin --tls-no-verify
```

Step 7 - Create function secret (only required once)

```
faas-cli secret create vc-datastore-config --from-file=vc-datastore-config.json --tls-no-verify
```

Step 8 - Deploy function to VMware Event Broker Appliance

```
faas-cli deploy -f stack.yml --tls-no-verify
```