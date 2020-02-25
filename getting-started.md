# Getting Started with vCenter Event Broker Appliance

## Appliance Deployment

### Requirements

* 2 vCPU and 8GB of memory for vCenter Event Broker Appliance
* vCenter Server 6.x or greater
* Account to login to vCenter Server (readOnly is sufficient)

**Step 1** - Download the vCenter Event Broker Appliance (OVA) from the [VMware Fling site](https://flings.vmware.com/vcenter-event-broker-appliance).

**Step 2** - Deploy the vCenter Event Broker Appliance OVA to your vCenter Server using the vSphere HTML5 Client. As part of the deployment you will be prompted to provide the following input:

*Networking*

  * Hostname - The FQDN of the vCenter Event Broker Appliance. If you do not have DNS in your environment, make sure the hostname provide is resolvable from your desktop which may require you to manually add a hosts entry. Proper DNS resolution is recommended
  * IP Address - The IP Address of the vCenter Event Broker Appliance
  * Netmask Prefix - CIDR Notation (e.g. 24 = 255.255.255.0)
  * Gateway - The Network Gateway address
  * DNS - DNS Server(s) that will be able to resolve to external sites such as Github for initial configuration. If you have multiple DNS Servers, input needs to be space separated.
  * DNS Domain - The DNS domain of your network

*Credentials*

  * Root Password - This is the OS root password for the vCenter Event Broker Appliance
  * OpenFaaS Password - This is the Admin password for OpenFaaS UI

*vSphere*

  * vCenter Server - This FQDN or IP Address of your vCenter Server that you wish to associate this vCenter Event Broker Appliance to for Event subscription
  * vCenter Username - The username to login to vCenter Server, as mentioned earlier, readOnly account is sufficient
  * vCenter Password - The password to the vCenter Username
  * Disable vCenter Server TLS Verification - If you have a self-signed SSL Certificate, you will need to check this box

*zDebug*

  * Debugging - When enabled, this will output a more verbose log file that can be used to troubleshoot failed deployments

**Step 3** - Power On the vCenter Event Broker Appliance after successful deployment. Depending on your external network connectivity, it can take a few minutes while the system is being setup. You can open the VM Console to view the progress. Once everything is completed, you should see an updated login banner for the various endpoints:

```
Appliance Status: https://[hostname]/status
Install Logs: https://[hostname]/bootstrap
OpenFaaS UI: https://[hostname]
```

**Note**: If you enable Debugging, the install logs endpoint will automatically contain the more verbose log entries.

**Step 4** - You can verify that everything was deployed correctly by opening a web browser to the OpenFaaS UI available on https://[hostname]/ and logging in with the Admin credentials (user:admin) you had specified as part of the OVA deployment.

At this point, you have successfully deployed the vCenter Event Broker Appliance and you are ready to start deploying your functions! Check the [examples](./examples/README.md) to quickly get started.
