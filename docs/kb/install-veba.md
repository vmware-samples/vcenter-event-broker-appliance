---
layout: docs
toc_id: install-veba
title: VMware Event Broker Appliance
description: Deploying VMware Event Broker Appliance
permalink: /kb/install-veba
cta:
 title: Deploy a Function
 description: At this point, you have successfully deployed the VMware Event Broker Appliance and you are ready to start deploying your functions!
 actions:
  - text: Check the [Knative Echo Function](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/examples/knative/powershell/kn-ps-echo){:target="_blank"} to quickly get started
---

# Deploy VMware Event Broker Appliance

Customers looking to seamlessly extend their vCenter by either deploying our prebuilt functions or writing your own functions can get started quickly by deploying VMware Event Broker Appliance with Knative as the Event Processor

<!-- omit in toc -->
## Table of Contents

- [Deploy VMware Event Broker Appliance](#deploy-vmware-event-broker-appliance)
  - [Requirements](#requirements)
  - [Step 1 - Download OVA](#step-1---download-ova)
  - [Step 2 - Deploy OVA](#step-2---deploy-ova)
    - [Networking (Required)](#networking-required)
    - [Proxy Settings (Optional)](#proxy-settings-optional)
    - [OS Credentials (Required)](#os-credentials-required)
    - [vSphere (Required)](#vsphere-required)
    - [Horizon (Optional)](#horizon-optional)
    - [Webhook (Optional)](#webhook-optional)
    - [Custom TLS Certificate Configuration (Optional)](#custom-tls-certificate-configuration-optional)
    - [Syslog Server Configuration (Optional)](#syslog-server-configuration-optional)
    - [zAdvanced (Optional)](#zadvanced-optional)
  - [Step 3 - Verification](#step-3---verification)

## Requirements

- 6 vCPU and 8GB of memory for VMware Event Broker Appliance
- vCenter Server 7.x or greater
  - **The VEBA UI requires vCenter Server 7.0 or greater**
- vCenter TCP/443 accessible from Appliance IP address
- Account to login to vCenter Server (readOnly is sufficient)

## Step 1 - Download OVA

Download the VMware Event Broker Appliance (OVA) from the [VMware Fling site](https://vmwa.re/flings){:target="_blank"}.

## Step 2 - Deploy OVA

Deploy the VMware Event Broker Appliance OVA to your vCenter Server using the vSphere HTML5 Client. As part of the deployment you will be prompted to provide the following input:

### Networking (Required)

- Hostname - The FQDN of the VMware Event Broker Appliance. If you do not have DNS in your environment, make sure the hostname provide is resolvable from your desktop which may require you to manually add a hosts entry. Proper DNS resolution is recommended
- IP Address - The IP Address of the VMware Event Broker Appliance
- Network Prefix - Network CIDR Selection (e.g. 24 = 255.255.255.0)
- Gateway - The Network Gateway address
- DNS - DNS Server(s) that will be able to resolve to external sites such as Github for initial configuration. If you have multiple DNS Servers, input needs to be **space separated**.
- DNS Domain - The DNS domain of your network
- NTP Server - NTP Server(s) for proper time synchronization. If you have multiple NTP Servers, input needs to be **space separated**.

### Proxy Settings (Optional)

- HTTP Proxy Server - HTTP Proxy Server followed by the port (e.g. http://proxy.provider.com:3128)
- HTTPS Proxy - HTTPS Proxy Server followed by the port (e.g. http(s)://proxy.provider.com:3128)
- Proxy Username - Optional Username for Proxy Server
- Proxy Password - Optional Password for Proxy Server
- No Proxy - Exclude internal domain suffix. Comma separated (localhost, 127.0.0.1, domain.local)

### OS Credentials (Required)

- Root Password - This is the OS root password for the VMware Event Broker Appliance
- Enable SSH - Check the box to allow SSH to the Appliance (SSH to the appliance is disabled by default)
- Endpoint Username - Specify the username to authenticate against the VEBA endpoints (e.g. /bootstrap, /events, /top, etc.)
- Endpoint Password - Specify the password to authenticate against the VEBA endpoints.

### vSphere (Required)

- vCenter Server - This FQDN or IP Address of your vCenter Server that you wish to associate this VMware Event Broker Appliance to for Event subscription
- vCenter Username - The username to login to vCenter Server, as mentioned earlier, readOnly account is sufficient
- vCenter Password - The password to the vCenter Username
- vCenter Username to register VEBA UI (Optional) - Username to register VMware Event Broker UI to vCenter Server for Knative Processor
- vCenter Password to register VEBA UI (Optional) - Password to register VMware Event Broker UI to vCenter Server for Knative Processor
- Disable vCenter Server TLS Verification - If you have a self-signed SSL Certificate, you will need to check this box
- vCenter Checkpointing Age - Maximum allowed age (seconds) for replaying events determined by last successful event in checkpoint (default 300s)
- vCenter Checkpointing Period - Period (seconds) between saving checkpoints (default 10s)

**Note:** The minimum vSphere Privileges that is required for proper VEBA UI functionality are: **Register Extension**, **Update Extension** (Installing Plugins) and **Manage Plugins** (Updating Plugins)

**Note:** For more information about the Checkpointing feature, see here: [Event Delivery](./intro-tanzu-sources.md#event-provider-delivery-guarantees).

### Horizon (Optional)

- Enable Horizon Event Provider - Enable Horizon Event Provider
- Horizon Server - IP Address or Hostname of Horizon Server
- Horizon Domain Name - Active Directory Domain the username to login to the Horizon Server belongs to (e.g. corp)
- Horizon Username - Username to login to Horizon Server (UPN-style not allowed)
- Horizon Password - Password to login to Horizon Server
- Disable Horizon Server TLS Verification - Disable TLS Verification for Horizon Server (required for self-sign certificate)

> **Note:** The minimum Horizon Role that is required to retrieve events is the `"Collect Operation Logs"` Role (located under Logs)

### Webhook (Optional)

- Enable Webhook Event Provider - Enable Webhook Event Provider
- Basic Auth Username (Optional) - Username to login to webhook endpoint
- Basic Auth Password (Optional) - Password to login to webhook endpoint

### Custom TLS Certificate Configuration (Optional)

- Custom VMware Event Broker Appliance TLS Certificate Private Key (Base64) - Base64 encoded custom TLS certificate (.PEM) for the VMware Event Broker Appliance
- Custom VMware Event Broker Appliance TLS Certificate Authority Certificate (Base64) - Base64 encoded custom TLS certificate (.CER) for the VMware Event Broker Appliance

### Syslog Server Configuration (Optional)

- Hostname or IP Address - Specify the Hostname (FQDN) or IP Address of the Syslog Server
- Port - Syslog Server Port
- Protocol - Choose the Transport Protocol (TCP, TLS or UDP)
- Format - Choose the Syslog Protocol Format (RFC5424 or RFC3164)

### zAdvanced (Optional)

- Debugging - When enabled, this will output a more verbose log file that can be used to troubleshoot failed deployments
- POD CIDR Network - Customize POD CIDR Network (Default 10.99.0.0/20). Must not overlap with the appliance IP address

## Step 3 - Verification

Power On the VMware Event Broker Appliance after successful deployment. Depending on your external network connectivity, it can take a few minutes while the system is being setup. You can open the VM Console to view the progress. Once everything is completed, you should see an updated login banner for the various endpoints:

```code
Appliance Configuration

Install Logs: https://[hostname]/bootstrap
Resource Utilization: https://[hostname]/top
Events: https://[hostname]/events
Webhook: https://[hostname]/webhook

Appliance Provider Stats

Webhook: https://[hostname]/stats/webhook
```

> NOTE: If you enable Debugging, the install logs endpoint will automatically contain the more verbose log entries.

You can verify that everything was deployed correctly by opening a web browser and accessing one of the endpoints (`/bootstrap`, `/events`, `/top`, etc.) along with the associated admin password you had specified as part of the OVA deployment.
