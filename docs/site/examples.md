---
layout: page
id: functions
title: Prebuilt Functions
description: Community-sourced and validated prebuilt functions for OpenFaaS with VEBA
permalink: /examples
images:
  powercli: /assets/img/languages/powercli.png
  python: /assets/img/languages/python.png
  go: /assets/img/languages/go.png
  powershell: /assets/img/languages/powershell.png
examples:
  - title: vSphere Tagging
    usecases: 
    - item: automation
    id: vsphere-tag
    description: Automatically tag a VM upon a vCenter event (ex. a VM can be tagged during a poweron event)
    links:
     - language: python
       image: {{ page.images.python }}
       url: "/tree/master/examples/python/tagging"
     - language: powercli
       url: "/tree/master/examples/powercli/tagging"
     - language: golang
       url: "/tree/master/examples/go/tagging"

  - title: Send VM Configuration Changes to Slack
    usecases: 
    - item: integration
    - item: notification
    id: config-changes-to-slack
    description: Notify a Slack channel upon a VM configuration change event
    links: 
    - language: powercli
      url: "/tree/master/examples/powercli/hwchange-slack"

  - title: Disable Alarms for Host Maintenance
    usecases: 
    - item: automation
    id: disable-host-maintenance-alarms
    description: Disable alarm actions on a host when it has entered maintenance mode and will re-enable alarm actions on a host after it has exited maintenance mode
    links: 
    - language: powercli
      url: "/tree/master/examples/powercli/hostmaint-alarms"

  - title: ESX Maximum transmission unit fixer
    usecases: 
    - item: automation
    - item: remediation
    id: esx-mtu-fixer
    description: Remediation function which will be triggered when a VM is powered on to ensure that the Maximum Transmission Unit (MTU) of the VM Kernel Adapter on all ESX hosts is at least 1500
    links: 
    - language: python
      url: "/tree/master/examples/python/esx-mtu-fixer"

  - title: Datastore Usage Notification
    usecases: 
    - item: notification
    id: datastore-usage-notification
    description: Send an email notification when warning/error threshold is reach for Datastore Usage Alarm in vSphere
    links: 
    - language: powercli
      url: "/tree/master/examples/powercli/datastore-usage-email"

  - title: vRealize Orchestrator
    usecases: 
    - item: integration
    - item: remediation
    id: vrealize-workflow
    description: Trigger vRealize Orchestrator workflow using vRO REST API
    links: 
    - language: powershell
      url: "/tree/master/examples/powershell/vro"

  - title: Echo VEBA Event
    usecases: 
    - item: other
    id: echo-function
    description: Function helps users understand the structure and data of a given vCenter Event which will be useful when creating brand new Functions.
    links: 
    - language: powershell
      url: "/tree/master/examples/python/echo"

  - title: Trigger PagerDuty incident
    usecases: 
    - item: integration
    - item: notification
    - item: remediation
    id: invoke-pagerduty
    description: Trigger a PagerDuty incident upon a vCenter Event
    links: 
    - language: python
      url: "/tree/master/examples/python/trigger-pagerduty-incident"

  - title: POST to any REST API
    usecases: 
    - item: automation
    - item: integration
    - item: notification
    - item: remediation
    id: post-res-api
    description: Function allows making a single post api request to any endpoint - tested with Slack, ServiceNow and PagerDuty
    links: 
    - language: python
      url: "/tree/master/examples/python/invoke-rest-api"

  - title: HA Restarted VMs Notification
    usecases: 
    - item: notification
    id: ha-restarted-vms
    description: Send an email listing all of the VMs which were restarted due to a host failure in an HA enabled cluster.
    links: 
    - language: powercli
      url: "/tree/master/examples/powercli/ha-restarted-vms"

  - title: VMware Cloud on AWS SDDC Provisioned and Deletion Slack Notification
    usecases:
    - item: notification
    id: vmware-cloud-ngw-slack
    description: Send Slack message when a VMware Cloud on AWS SDDC is Provisioned or Deleted.
    links:
    - language: powershell
      url: "/tree/master/examples/powershell/vmware-cloud-ngw-slack"

  - title: VMware Cloud on AWS SDDC Provisioned and Deletion Microsoft Teams Notification
    usecases:
    - item: notification
    id: vmware-cloud-ngw-teams
    description: Send Microsoft Team message when a VMware Cloud on AWS SDDC is Provisioned or Deleted.
    links:
    - language: powershell
      url: "/tree/master/examples/powershell/vmware-cloud-ngw-teams"
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