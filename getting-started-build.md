## Getting Started Build Guide for vCenter Event Broker Appliance

## Requirements

* 2 vCPU and 8GB of memory for vCenter Event Broker Appliance
* vCenter Server or Standalone ESXi host 6.x or greater
* [VMware OVFTool](https://www.vmware.com/support/developer/ovf/)
* [Docker Client](https://docs.docker.com/v17.09/engine/installation/)
* [OpenFaaS CLI](https://github.com/openfaas/faas-cli)
* [Packer](https://www.packer.io/intro/getting-started/install.html)


Step 1 - Clone the vCenter Event Broker Appliance Git repository

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance.git
```

Step 2 - Edit the `photon-builder.json` file to configure the vSphere endpoint for building the vCenter Event Broker Appliance

```
{
  "builder_host": "192.168.30.10",
  "builder_host_username": "root",
  "builder_host_password": "VMware1!",
  "builder_host_datastore": "vsanDatastore",
  "builder_host_portgroup": "VM Network"
}
```

**Note:** If you need to change the default root password on the vCenter Event Broker Appliance, take a look at `photon-version.json`

Step 3 - Start the build by running the build script

```
./build.sh
````

If you wish to automatically deploy the vCenter Event Broker Appliance after successfully building the OVA. You can edit the `photon-dev.xml.template` file and change the `ovftool_deploy_*` variables and run `./build.sh dev` instead.
