---
layout: docs
toc_id: intro-about
title: VMware Event Broker Appliance - Introduction
description: VMware Event Broker Appliance - Introduction
permalink: /kb/
cta:
 title: Getting Started
 description: Get started with VMware Event Broker Appliance and extend your vSphere SDDC in under 60 minutes
 actions:
    - text: Install the [Appliance with Knative](install-knative) to extend your SDDC with our [community-sourced functions](/examples)
---

# VMware Event Broker Appliance

The [VMware Event Broker Appliance](https://flings.vmware.com/vmware-event-broker-appliance#summary) Fling enables customers to unlock the hidden potential of events in their SDDC to easily create [event-driven automation](https://octo.vmware.com/vsphere-power-event-driven-automation/). The VMware Event Broker Appliance includes support for vCenter Server and VMware Horizon events as well as any valid `CloudEvent` through the native webhook event provider. Easily triggering custom or prebuilt actions to deliver powerful integrations within your datacenter across public cloud has never been more easier before. A detailed list of use cases and possibilities with VMware Event Broker Appliance is available [here](https://vmweventbroker.io)

VMware Event Broker Appliance is provided as a virtual appliance that can be deployed to any vSphere-based infrastructure, including an on-premises and/or any public cloud environment running on vSphere such as VMware Cloud on AWS or VMware Cloud on DellEMC.

With this solution, end-users, partners and independent software vendors only have to write minimal business logic without going through a steep learning curve understanding the vSphere or Horizon APIs. As such, we believe this solution not only offers a better user experience in solving existing problems for VI/Cloud Admins, SRE/Operators, Automation Engineers and 3rd Party Vendors. More importantly, it will enable new integration use cases and workflows to grow the VMware ecosystem and community, similar to what AWS has achieved with AWS Lambda.

## Use Cases

VMware Event Broker Appliance enables customers to quickly get started with pre-built functions and enable the following use cases:

### Notification:
- Receive alerts and real time updates using your preferred communication channel such as SMS, Slack, Microsoft Teams, etc.
- Real time updates for specific vSphere Inventory objects which matter to you and your organization
- Monitor the health, availability & capacity of SDDC resources

### Automation:
- Apply configuration or customization changes based on specific VM or Host life cycle activities as an example within the SDDC (e.g. apply security settings to a VM or vSphere Tag to Host)
- Scheduled jobs to validate health of an environment such as a long running snapshot on a VM

### Integration:
- Consume 2nd/3rd party solutions that provide remote APIs to associate with specific infrastructure events
- Automated ticket creation using platforms such as Pager Duty, ServiceNow, Jira Service Desk, Salesforce based specific incidents such as workload and/or hardware failure as an example
- Easily extend and consume public cloud services such as AWS EventBridge

### Remediation:
- Detect and automatically perform specific tasks based on certain types of events (e.g. add or request additional capacity)
- Enables Operations and SRE teams to codify existing and well known run books for automated resolution

### Audit:
- Track all configuration changes for objects like a VM and automatically update a change management database (CMDB)
- Forward all authentication and authorization events to your security team for compliance and/or intrusion detection
- Replay configuration changes to aide in troubleshooting or debugging purposes

### Analytics:
- Reduce the number of connections and/or users to vCenter Server by providing access to events in an external system like CMDB or data warehouse solution
- Enable teams to better understand workload and infrastructure behaviors by identifying trends observed in the events data including duration of events, users generating specific operations or the commonly used workflows

### Community Use Cases

Please see [this list here](https://github.com/vmware-samples/vcenter-event-broker-appliance/blob/master/USECASES.md) for a collection of use cases from the VMware Event Broker Appliance community.