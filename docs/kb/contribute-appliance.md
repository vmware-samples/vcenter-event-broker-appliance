---
layout: docs
toc_id: contribute-appliance
title: Building the VMware Event Broker Appliance 
description: Building the VMware Event Broker Appliance
permalink: /kb/contribute-appliance
cta:
 title: Have a question? 
 description: Please check our [Frequently Asked Questions](/faq) first.
---

## Getting Started Build Guide for VMware Event Broker Appliance

## Requirements

* 6 vCPU and 8GB of memory for VMware Event Broker Appliance
* ESXi host v6.7 or greater
  * Datastore with at least 60GB of free space
  * SSH must be enabled on the host
  * Enable GuestIPHack on the host by running `esxcli system settings advanced set -o /Net/GuestIPHack -i 1`
* The following must be installed on your development machine:
  * [VMware OVFTool](https://www.vmware.com/support/developer/ovf/){:target="_blank"}
  * [Docker Client](https://docs.docker.com/v17.09/engine/installation/){:target="_blank"}
  * [Packer](https://www.packer.io/intro/getting-started/install.html){:target="_blank"} (v1.6.3 or greater)
  * [jq](https://stedolan.github.io/jq/){:target="_blank"}
* Development machine must have the firewall disabled for the duration of the build
> **Note:** It has been seen that Packer can bind to an IPv6 on the development machine - you may wish to disable IPv6!
* Development machine must be on the same L2 subnet as the target VM portgroup defined in `builder_host_portgroup` below


Step 1 - Clone the VMware Event Broker Appliance Git repository

```
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance.git
```

Step 2 - Edit the `photon-builder.json` file to configure the vSphere endpoint for building the VMware Event Broker Appliance

```
{
  "builder_host": "192.168.30.10",
  "builder_host_username": "root",
  "builder_host_password": "VMware1!",
  "builder_host_datastore": "vsanDatastore",
  "builder_host_portgroup": "VM Network"
}
```

> **Note:** If you need to change the default root password on the VMware Event Broker Appliance, take a look at `photon-version.json`

Step 3 - The `veba-bom.json` will need to be updated to specify the branch you wish to build the vCenter Event Broker Appliance code from whether that is from master, a release- branch or from development. Below are two examples of how to correctly set the versions needed prior to building.

> **Note:** The default BOM version in development will be the development branch. No changes will be necessary unless you wish to build from a release or master branch.

Example 1 (build from master branch):
```
".veba.version" => "v0.5.0"
"vmware-event-router.version" => "v0.5.0"
"vmware-event-router.containers[0].version" => "v0.5.0"
```

Example 2 (build from development branch):
```
".veba.version" => "development"
"vmware-event-router.version" => "development"
"vmware-event-router.containers[0].version" => "development"
```

* master branch will be reflected using a stable tag, e.g. v0.5.0
* Any release- branch will be reflected using release-<version> omitting v for backwards-compat reasons, e.g. release-0.5.0
* Router image tags, based on the branch where changes are pushed to, will use <image>:v0.5.0 for master, <image>:release-0.5.0 for release-<version> and <image>:development on every push to development branch. In addition, the master and development container images will also be tagged with the corresponding COMMIT_ID of the pushed commit.

Step 4 - Start the build by running the build script

```
./build.sh
````

If you wish to automatically deploy the VMware Event Broker Appliance after successfully building the OVA, please take a look at the script samples located in the test directory.
