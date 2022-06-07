---
layout: docs
toc_id: use-vcenter-events
title: VMware Event Broker Appliance - vCenter Events
description: VMware Event Broker Appliance - vCenter Events
permalink: /kb/vcenter-events
cta:
 title: Deploy Event-Driven Functions
 description: Extend your vCenter seamlessly with our pre-built functions
 actions:
    - text: Get started quickly by deploying from the [community-sourced, pre-built functions](/examples)
    - text: Learn more about the [Event Specification](eventspec) to understand how the events are sent to the Functions
    - text: Deploy a function using these [instructions](use-functions) and learn how to [write your own function](contribute-functions).
---

# vCenter Events

vCenter produces events that get generated in response to actions taken on an entity such as VM, Host, Datastore, etc. These events contain immutable facts documenting the entity state changes such as who initiated the change, what action was performed, which object was modified, and when was the change initiated.

Events naturally serve as auditing and troubleshooting tools, allowing an administrator to retrieve details on a specific change. Event Driven Automation builds on the construct of events and enables advanced distributed design patterns driven through Events. VMware Event Broker Appliance aims to enable this for VMware SDDC by enabling VI Administrators to write lean functions (script or code) that are triggered by vCenter Events.

## Overview of the vCenter events

vCenter Events are categorized by the Objects and the actions that are allowed on these objects and are documented under the vSphere API [7.0U3 reference](https://vdc-download.vmware.com/vmwb-repository/dcr-public/bf660c0a-f060-46e8-a94d-4b5e6ffc77ad/208bc706-e281-49b6-a0ce-b402ec19ef82/SDK/vsphere-ws/docs/ReferenceGuide/vim.event.Event.html){:target="_blank"}.

* Event
  * ClusterEvent
    * ClusterCreatedEvent, ClusterDestroyedEvent, ClusterOvercommittedEvent...
  * DatastoreEvent
    * DatastoreCapacityIncreasedEvent, DatastoreDestroyedEvent, DatastoreDuplicatedEvent...
  * DatacenterEvent
    * DatacenterCreatedEvent, DatacenterRenamedEvent
  * HostEvent
    * HostShutdownEvent, HostAddedEvent, EnteringMaintenanceModeEvent...
  * VMEvent
    * VmNoNetworkAccessEvent, VmOrphanedEvent, VmPoweredOffEvent...
  * ...

There are over 1900+ events available on an out of the box install of vCenter that are provided [here](https://github.com/lamw/vcenter-event-mapping/){:target="_blank"}. You can also get the complete list of events for your specific vCenter using the following powershell script below.

```powershell
$vcNames = "hostname"

Connect-VIServer -Server $vcNames

$vcenterVersion = ($global:DefaultVIServer.ExtensionData.Content.About.ApiVersion)

$eventMgr = Get-View $global:DefaultVIServer.ExtensionData.Content.EventManager

$results = @()
foreach ($event in $eventMgr.Description.EventInfo) {
    if($event.key -eq "EventEx" -or $event.key -eq "ExtendedEvent") {
        #echo $event
        $eventId = ($event.FullFormat.toString()) -replace "\|.*",""
        $eventType = $event.key
    } else {
        $eventId = $event.key
        $eventType = "Standard"
    }
    $eventCategory = $event.Category
    $eventDescription = $event.Description

    $tmp = [PSCustomObject] @{
        EventId = $eventId;
        EventCategory = $eventCategory
        EventType = $eventType;
        EventDescription = $($eventDescription.Replace("<","").Replace(">",""));
    }

    $results += $tmp
}

Write-Host "Number of Events: $($results.count)"
$results | Sort-Object -Property EventId | ConvertTo-Csv | Out-File -FilePath vcenter-$vcenterVersion-events.csv

Write-Host "Disconnecting from vCenter Server ..."
Disconnect-VIServer * -Confirm:$false
```
