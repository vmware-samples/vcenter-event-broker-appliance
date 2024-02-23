---
layout: docs
toc_id: use-eventspec
title: VMware Event Broker Appliance - Event Schema
description: VMware Event Broker Appliance Event Schema
permalink: /kb/eventspec
cta:
 title: Get Started
 description: Explore the capabilities that the VMware Event Router enables
 actions:
    - text: Get started quickly by deploying from the [community-sourced, pre-built functions](/examples)
    - text: Deploy a function using these [instructions](use-functions) and learn how to [write your own function](contribute-functions).
---

# The Event Specification

The event payload structure used by the VMware Event Broker Appliance uses the
[CloudEvents](https://cloudevents.io/){:target="_blank"} v1 specification for
cross-cloud portability.

Events produced by the supported event `providers`, e.g. `vcenter` and `horizon`
are JSON-encoded and injected into the CloudEvents `data` attribute. The current
data content-type, which is sent as payload to a supported event processor, is
`application/json`.

Based on defined `triggers`, the `broker` in the VMware Event Broker Appliance
sends these events to registered event `processors` (i.e. functions). By
default, the `broker` sends CloudEvents via the HTTP protocol using `binary`
encoding. That is, the key CloudEvent attributes, e.g. `id`, `source`, `type`,
etc. are set via HTTP headers. The HTTP body contains the event as emitted by
the event `provider`, e.g. `vcenter`.

Please use one of the provided CloudEvents [SDKs](https://cloudevents.io/) to
ease the consumption and handling of these events.

<!-- omit in toc -->
## Table of Contents

- [The Event Specification](#the-event-specification)
  - [Example](#example)
    - [HTTP Headers](#http-headers)
      - [Description](#description)
    - [HTTP Body](#http-body)

## Example

The following example shows a converted CloudEvent published by the `vcenter`
event provider (optimized for readability) using CloudEvent HTTP `binary` mode
transport encoding.

### HTTP Headers

Key HTTP headers used:

```json
{
  "Ce-Id": "08179137-b8e0-4973-b05f-8f212bf5003b",
  "Ce-Source": "https://vcenter-01:443/sdk",
  "Ce-Specversion": "1.0",
  "Ce-Time": "2021-09-27T19:02:54.063Z",
  "Ce-Type": "com.vmware.vsphere.<event>.v0",
  "Content-Type": "application/json",
}
```

#### Description

- `id:` The unique ID ([UUID v4](https://tools.ietf.org/html/rfc4122){:target="_blank"}) of the event
- `source:` The vCenter emitting the embedded vSphere event (FQDN resolved when available)
- `specversion:` The CloudEvent specification the used
- `type:` The canonical name of the event class in "." dot notation
- `time:` Timestamp when this event was produced by the event `provider` (`vcenter`)
- `content-type:` Data (payload) encoding scheme used (JSON)

### HTTP Body

The event as emitted by vCenter:

```json
{
  "Key": 280034,
  "ChainId": 280032,
  "CreatedTime": "2024-01-17T13:02:05.426999Z",
  "UserName": "VSPHERE.LOCAL\\Administrator",
  "Datacenter": {
    "Name": "jarvis-lab",
    "Datacenter": {
      "Type": "Datacenter",
      "Value": "datacenter-1001"
    }
  },
  "ComputeResource": {
    "Name": "mk-1",
    "ComputeResource": {
      "Type": "ClusterComputeResource",
      "Value": "domain-c1015"
    }
  },
  "Host": {
    "Name": "esxi02.jarvis.lab",
    "Host": {
      "Type": "HostSystem",
      "Value": "host-1012"
    }
  },
  "Vm": {
    "Name": "workload1",
    "Vm": {
      "Type": "VirtualMachine",
      "Value": "vm-2073"
    }
  },
  "Ds": null,
  "Net": null,
  "Dvs": null,
  "FullFormattedMessage": "DRS powered on workload1 on esxi02.jarvis.lab in jarvis-lab",
  "ChangeTag": "",
  "Template": false
}
```
