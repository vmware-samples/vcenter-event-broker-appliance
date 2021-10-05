---
layout: docs
toc_id: use-horizon-events
title: VMware Event Broker Appliance - Horizon Events
description: VMware Event Broker Appliance - Horizon Events
permalink: /kb/horizon-events
cta:
 title: Deploy Event-Driven Functions
 description: Extend your VMware Horizon seamlessly with our pre-built functions
 actions:
    - text: Get started quickly by deploying from the [community-sourced, pre-built functions](/examples)
    - text: Learn more about the [Event Specification](eventspec) to understand how the events are sent to the Functions
    - text: Deploy a function using these [instructions](use-functions) and learn how to [write your own function](contribute-functions).
---

# Horizon Events

VMware Horizon produces events generated in response to actions and/or operations within a Horizon deployment. These events are immutable facts documenting the entity state changes such as who initiated the change, what action was performed, which object was modified, and when was the change initiated.

Events naturally serve as auditing and troubleshooting tools, allowing an administrator to retrieve details on a specific change and react to it with event-driven automation. The VMware Event Broker Appliance unlocks this functionality for operation teams, such as VI Administrators, by writing lean code ("functions") that is triggered by VMware Horizon events.

## Overview of the Horizon events

The VMware Event Broker Appliance takes advantage of a new [Audit Events API](https://williamlam.com/2021/08/listing-all-vmware-horizon-events.html) that was recently introduced with VMware Horizon (2106). There are over 850+ events that are available and for a complete list, please take a look [here](https://github.com/lamw/horizon-event-mapping/){:target="_blank"}.
