---
layout: docs
toc_id: use-eventspec
title: VMware Event Broker Appliance - Architecture
description: VMware Event Broker Appliance Architecture
permalink: /kb/eventspec
cta:
 title: Get Started
 description: Explore the capabilities that the VMware Event Router enables
 actions:
    - text: Get started quickly by deploying from the [community-sourced, pre-built functions](/examples)
    - text: Deploy a function using these [instructions](use-functions) and learn how to [write your own function](contribute-functions).
---

# The Event Specification

The event payload structure used by the VMware Event Broker Appliance follows the [CloudEvents](https://github.com/cloudevents/sdk-go/blob/master/pkg/cloudevents/eventcontext_v1.go){:target="_blank"} v1 specification for cross-cloud portability. The current data content type which is sent as payload to a supported event processor is JSON.

The following example shows the event structure sent as JSON to a supported event processor (trimmed for better readability):

```json
{
  "id": "08179137-b8e0-4973-b05f-8f212bf5003b",
  "source": "https://vcenter-01:443/sdk",
  "specversion": "1.0",
  "type": "com.vmware.event.router/event",
  "subject": "VmPoweredOffEvent",
  "time": "2020-02-11T21:29:54.9052539Z",
  "data": {
    "Key": 9902,
    "ChainId": 9895,
    [.....]
  },
  "datacontenttype": "application/json"
}
```

`id:` The unique ID ([UUID](https://tools.ietf.org/html/rfc4122){:target="_blank"}) of the event

`source:` The vCenter emitting the embedded vSphere event (FQDN resolved when available)

`specversion:` The event specification the appliances uses (can be used for schema handling)

`type:` The canonical name of the event class in "." dot notation 

`subject:` The vCenter event name (CamelCase)

`time:` Timestamp when this event was produced by the appliance

`data:` Original vCenter event

`data.Key:` Monotonically increasing value set by vCenter (the lower the key, the older the message as being created by vCenter)

`data.CreatedTime:` When the embedded event was created by vCenter

`datacontenttype:` Encoding used (JSON)

Please see the section on function [best practices](contribute-functions.md) below how you can make use of these fields for advanced requirements.
