## Description

This example will disable alarm actions on a host while it is in maintenance mode.  It deploys two functions that use the same PowerCLI script.  The first function subscribes to the `entered.maintenance.mode` event to run when a host is put into maintenance mode and disable alarms.  The second function subscribes to the `exit.maintenance.mode` event to re-enable alarms when the host exits maintenance mode.  There is an accompanying blog post with more details:  [Automate Host Maintenance with the vCenter Event Broker Appliance
](https://doogleit.github.io/2019/11/automate-host-maintenance-with-the-vcenter-event-broker-appliance/)

## Deployment

1. Update the vcconfig.json file with your vCenter information and create the secret.  If you already have this secret created from one of the other PowerCLI examples you can skip this step.

```json
{
    "VC" : "vcenter-hostname",
    "VC_USERNAME" : "veba@vsphere.local",
    "VC_PASSWORD" : "FillMeIn"
}
```
```shell
faas-cli secret create vcconfig --from-file=vcconfig.json --tls-no-verify
```

2. Update the gateway in the stack.yml file with your vCenter Event Broker Appliance address and deploy the functions.
```yaml
provider:
  name: openfaas
  gateway: https://veba.yourdomain.com
...
```
```shell
faas-cli deploy -f stack.yml --tls-no-verify
```

