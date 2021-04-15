---
layout: function
id: functions
type: openfaas
title: Prebuilt Functions
description: Community-sourced and validated prebuilt functions for OpenFaaS with VEBA.
permalink: /examples
images:
  powercli: /assets/img/languages/powercli.png
  python: /assets/img/languages/python.png
  go: /assets/img/languages/go.png
  powershell: /assets/img/languages/powershell.png
examples:
  - title: Virtual Machine Reconfiguration Via Tags
    usecases:
    - item: automation
    description: Many configurations such as enabling memory/cpu hot add, changing the number of CPUs, or changing amount of memory can be set only when a VM is powered off (if the hot plug/hot add settings are not yet enabled). Attach tags containing desired configuration settings to a VM and have it automatically reconfigure the next time it powers down.
    links:
    - language: go
      url: "/tree/master/examples/openfaas/go/vm-reconfig-via-tag"

  - title: vSphere Tagging
    usecases: 
    - item: automation
    id: vsphere-tag
    description: Automatically tag a VM upon a vCenter event (ex. a VM can be tagged during a poweron event).
    links:
     - language: python
       image: {{ page.images.python }}
       url: "/tree/master/examples/openfaas/python/tagging"
     - language: powercli
       url: "/tree/master/examples/openfaas/powercli/tagging"
     - language: go
       url: "/tree/master/examples/openfaas/go/tagging"

  - title: Send VM Configuration Changes to Slack
    usecases: 
    - item: integration
    - item: notification
    id: config-changes-to-slack
    description: Notify a Slack channel upon a VM configuration change event.
    links: 
    - language: powercli
      url: "/tree/master/examples/openfaas/powercli/hwchange-slack"

  - title: Disable Alarms for Host Maintenance
    usecases: 
    - item: automation
    id: disable-host-maintenance-alarms
    description: Disable alarm actions on a host when it has entered maintenance mode and will re-enable alarm actions on a host after it has exited maintenance mode.
    links: 
    - language: powercli
      url: "/tree/master/examples/openfaas/powercli/hostmaint-alarms"

  - title: ESX Maximum transmission unit fixer
    usecases: 
    - item: automation
    - item: remediation
    id: esx-mtu-fixer
    description: Remediation function which will be triggered when a VM is powered on to ensure that the Maximum Transmission Unit (MTU) of the VM Kernel Adapter on all ESX hosts is at least 1500.
    links: 
    - language: python
      url: "/tree/master/examples/openfaas/python/esx-mtu-fixer"

  - title: Datastore Usage Notification
    usecases: 
    - item: notification
    id: datastore-usage-notification
    description: Send an email notification when warning/error threshold is reach for Datastore Usage Alarm in vSphere.
    links: 
    - language: powercli
      url: "/tree/master/examples/openfaas/powercli/datastore-usage-email"

  - title: vRealize Orchestrator
    usecases: 
    - item: integration
    - item: remediation
    id: vrealize-workflow
    description: Trigger vRealize Orchestrator workflow using vRO REST API.
    links: 
    - language: powershell
      url: "/tree/master/examples/openfaas/powershell/vro"

  - title: Echo Cloud Event
    usecases: 
    - item: other
    id: echo-function
    description: Function helps users understand the structure and data of a given vCenter Event which will be useful when creating brand new Functions.
    links: 
    - language: python
      url: "/tree/master/examples/openfaas/python/echo"

  - title: Trigger PagerDuty incident
    usecases: 
    - item: integration
    - item: notification
    - item: remediation
    id: invoke-pagerduty
    description: Trigger a PagerDuty incident upon a vCenter Event.
    links: 
    - language: python
      url: "/tree/master/examples/openfaas/python/trigger-pagerduty-incident"
    - language: go
      url: "/tree/master/examples/openfaas/go/pagerduty-trigger"

  - title: POST to any REST API
    usecases: 
    - item: automation
    - item: integration
    - item: notification
    - item: remediation
    id: post-res-api
    description: Function allows making a single post api request to any endpoint - tested with Slack, ServiceNow and PagerDuty.
    links: 
    - language: python
      url: "/tree/master/examples/openfaas/python/invoke-rest-api"

  - title: HA Restarted VMs Notification
    usecases: 
    - item: notification
    id: ha-restarted-vms
    description: Send an email listing all of the VMs which were restarted due to a host failure in an HA enabled cluster.
    links: 
    - language: powercli
      url: "/tree/master/examples/openfaas/powercli/ha-restarted-vms"

  - title: VMware Cloud on AWS SDDC Provisioned and Deletion Slack Notification
    usecases:
    - item: notification
    id: vmware-cloud-ngw-slack
    description: Send Slack message when a VMware Cloud on AWS SDDC is Provisioned or Deleted.
    links:
    - language: powershell
      url: "/tree/master/examples/openfaas/powershell/vmware-cloud-ngw-slack"

  - title: VMware Cloud on AWS SDDC Provisioned and Deletion Microsoft Teams Notification
    usecases:
    - item: notification
    id: vmware-cloud-ngw-teams
    description: Send Microsoft Team message when a VMware Cloud on AWS SDDC is Provisioned or Deleted.
    links:
    - language: powershell
      url: "/tree/master/examples/openfaas/powershell/vmware-cloud-ngw-teams"

  - title: vCenter Managed Object Pre-Filter
    usecases: 
    - item: other
    id: pre-filter
    description: Function to limit the scope of other functions by allowing filtering of events by vCenter Inventory paths using standard regex.
    links: 
    - language: python
      url: "/tree/master/examples/openfaas/python/pre-filter" #relative path to the function

  - title: Auto-refresh of a vSphere Client UI plugin
    usecases:
    - item: automation
    - item: integration
    id: plugin-auto-refresh
    description: Plugin auto-refresh function triggers automatic refresh of the UI of a vSphere Client plugin after data changes.
    links:
    - language: java
      url: "/tree/master/examples/openfaas/java/plugin-auto-refresh"

  - title: Automatic Backup of Virtual Machines via Veeam Backup & Replication
    usecases:
    - item: automation
    - item: integration
    id: veeam-vm-backup
    description: Veeam-vm-backup function uses 3rd party solution Veeam to provide automatic backup for any virtual machine when when the VM state changes.
    links:
    - language: java
      url: "/tree/master/examples/openfaas/java/veeam-vm-backup"
---

A complete and updated list of ready to use functions curated by the VMware Event Broker community is listed below. 

# Get started with our prebuilt functions

These functions are prebuilt, available in ready to deploy container and `stack.yml` files for you to deploy as is. Should you need to modify the functions to fit your needs, the `README.md` files provided within each function folder will provide all the information you need to customize, build and deploy the function on your VMware Event Broker appliance. 

> **Note:** These functions are provided and tested to be used with the VMware Event Broker Appliance deployed with [OpenFaaS](/kb/install-openfaas) as the event stream processor. 


 <div class="examples wd-100">
    <h2>Functions</h2>
    {% for ex in page.examples %}
    <div id="{{ ex.id }}" class="row title">
      <h3>{{ex.title}}</h3>
      <div class="language">
      {% for link in ex.links %}
        <a href="{{ link.url | prepend: site.gh_repo}}" target="_blank" class="col-md-2 pr-0">
            <img src="{{ '/assets/img/languages/' | append: link.language | append: '.png' | relative_url}}" alt="{{ link.language }}">
            <span class="m-0">{{ link.language }}</span>
        </a>
      {% endfor %}
      </div>
    </div>
    {{ ex.description | markdownify }}
    <div class="usecases">
      {% for usecase in ex.usecases %}
      <span class="{{usecase.item}}">{{usecase.item}}</span>
      {% endfor %}
    </div>
    
    {% endfor %}
</div>

## Contributions

These functions serve as an easy way to use the appliance and as an inspiration for how to write functions in different languages. If you have an idea for a function and are looking to write your own, start with our documentation [here](/kb/contribute-functions). 

Check our [contributing guidelines](\community#contributing) and join [Team #VEBA](/#team-veba) by submitting a pull request for your function to be showcased on this list. 