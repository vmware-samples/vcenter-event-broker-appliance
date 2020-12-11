---
layout: docs
toc_id: intro-event-router
title: VMware Event Router - Introduction
description: VMware Event Router Introduction
permalink: /kb/event-router
cta:
 title: Get Started
 description: Explore the capabilities that the VMware Event Router enables
 actions:
    - text: Install the [Appliance with OpenFaaS](install-openfaas) to extend your SDDC with our [community-sourced functions](/examples)
    - text: Install the [Appliance with AWS EventBridge](install-eventbridge) to extend your SDDC leveraging native AWS capabilities.
    - text: Learn more about the [Events in vCenter](vcenter-events) and how to find the right event for your usecase
    - text: Learn more about Functions in this overview [here](functions).
---

# VMware Event Router

The VMware Event Router is responsible for connecting to event `stream` sources, such as VMware vCenter, and forward events to an event `processor`. To allow for extensibility and different event sources/processors event sources and processors are abstracted via Go `interfaces`.

Currently, one VMware Event Router is deployed per appliance (1:1 mapping). Only one vCenter event stream can be processed per appliance.  Also, only one event stream (source) and one processor can be configured. The list of supported event sources and processors can be found below.We are evaluating options to support multiple event sources (vCenter servers) and processors per appliance (scale up) or alternatively support multi-node appliance deployments (scale out), which might be required in large deployments (performance, throughput).

> **Note:** We have not done any extensive performance and scalability testing to understand the limits of the single appliance model.

## Supported Event Sources

- [VMware vCenter Server](https://www.vmware.com/products/vcenter-server.html){:target="_blank"}
- vCenter Simulator [vcsim](https://github.com/vmware/govmomi/tree/master/vcsim){:target="_blank"} (for testing purposes only)

## Supported Event Processors

- [Knative](https://knative.dev/)
- [OpenFaaS](https://www.openfaas.com/){:target="_blank"}
- [AWS EventBridge](https://aws.amazon.com/eventbridge/?nc1=h_ls){:target="_blank"}

# Event Handling

As described in the [architecture section](intro-architecture.md), due to the microservices architecture used in the VMware Event Broker Appliance one always has to consider message delivery problems such as timeouts, delays, reordering, loss. These challenges are fundamental to [distributed systems](https://github.com/papers-we-love/papers-we-love/blob/master/distributed_systems/a-note-on-distributed-computing.pdf){:target="_blank"} and must be understood and considered by function authors.

## Event Types supported

For the supported event stream source, e.g. VMware vCenter, all events provided by that source can be used. Since event types are environment specific (vSphere version, extensions), a list of events for vCenter as an event source can be generated as described in this [blog post](https://www.virtuallyghetto.com/2019/12/listing-all-events-for-vcenter-server.html){:target="_blank"}.

## Message Delivery Guarantees

Consider the following most basic form of messaging between two systems:

[PRODUCER]------[MESSAGE]----->[CONSUMER]  
[PRODUCER]<---[MESSAGE_ACK]---[CONSUMER]

Even though this example looks simple, a lot of things can go wrong when transferring a message over the network (vs in-process communication):

- The message might never be received by the consumer
- The message might arrive out of order (previous message not shown here)
- The message might be delayed during transport
- The message might be duplicated during transport
- The consumer might be slow acknowledging the message
- The consumer might receive the message and then crash before acknowledging it
- The consumer acknowledges the message but this message is lost/delayed/arrives out of order
- The producer crashes immediately after receiving the acknowledgement 

> **Note:** For our example, it doesn't really matter whether the packet (message) actually leaves the machine or the destination (consumer) is on the same host. Of course, having a physical network in between the actors increases the chances of [messaging failures](https://queue.acm.org/detail.cfm?id=2655736){:target="_blank"}. The network protocol in use was intentionally left unspecified. 

One of the following message delivery semantics is typically used to describe the messaging characteristics of a  distributed system such as the VMware Event Broker Appliance:

- At-most-once semantics: a message will be delivered once or not at all to the consumer
- At-least-once semantics: a message will be delivered once or multiple times to the consumer
- Exactly-once semantics: a message will be delivered exactly once to the consumer

> **Note:** Exactly-once semantics is not supported by all messaging systems as it requires significant engineering effort to implement. It is considered the gold standard in messaging while at the same time being a highly [debated](https://medium.com/@jaykreps/exactly-once-support-in-apache-kafka-55e1fdd0a35f){:target="_blank"} topic.

As of today the VMware Event Broker Appliance guarantees *at-most-once* as well as *at-least-once* event delivery semantics for the vCenter event provider (using checkpoints).

**Event Delivery Guarantees:**

- At-least-once event delivery
  - with the [vCenter event provider](https://vmweventbroker.io/kb/contribute-eventrouter) option `checkpoint: true`
- At-most-once event delivery
  - with the [vCenter event provider](https://vmweventbroker.io/kb/contribute-eventrouter) option `checkpoint: false`
  - with the [vcsim event provider](https://vmweventbroker.io/kb/contribute-eventrouter)

The VMware Event Broker Appliance currently does not persist (to disk) or retry event delivery in case of failure during function invocation or upstream (external system, such as Slack) communication issues. For introspection and debugging purposes invocations are logged to standard output by the OpenFaaS vcenter-connector ("sync" invocation mode) or OpenFaaS queue-worker ("async" invocation mode).

The chances for message delivery failures are/can be reduced by:

- Using TCP/IP as the underlying communication protocol which provides certain ordering (sequencing), back-pressure and retry capabilities at the transmission layer (default in the appliance)
- Using asynchronous function [invocation](#invocation) (defaults to "off", i.e. "synchronous", in the appliance) which internally uses a message queue for event processing
- Following [best practices](contribute-functions.md) for writing functions

## Invocation

Functions in OpenFaaS can be invoked synchronously or asynchronously:

`synchronous:` The function is called and the caller, e.g. OpenFaaS vcenter-connector, waits until the function returns (successful/error) or the timeout threshold is hit.

`asynchronous:` The function is not directly called. Instead, HTTP status code 202 ("accepted") is returned and the request, including the event payload, is stored in a [NATS Streaming](https://docs.nats.io/nats-streaming-concepts/intro){:target="_blank"} queue. One or more "queue-workers" process the queue items.

If you directly invoke your functions deployed in the appliance you can decide which invocation mode is used (per function). More details can be found [here](https://github.com/openfaas/workshop/blob/master/lab7.md){:target="_blank"}.

The VMware Event Broker appliance by default uses synchronous invocation mode. If you experience performance issues due to long-running/slow/blocking functions, consider running the VMware Event Router in asynchronous mode by setting the `"async"` option to `"true"` (quotes required) in the configuration file for the VMware Event Router deployment:

```json
{
    "type": "processor",
    "provider": "openfaas",
    "address": "http://127.0.0.1:8080",
    "auth": {
          ...skipped
        }
    },
    "options": {
        "async": "true"
    }
}
```

When the AWS EventBridge [event processor](#components) is used, events are only forwarded for the patterns configured in the AWS event rule ARN. For example, if the rule is configured with this event pattern:

```json
{
  "detail": {
    "subject": [
      "VmPoweredOnEvent",
      "VmPoweredOffEvent",
      "VmReconfiguredEvent"
    ]
  }
}
```

Only these three vCenter event types would be forwarded. Other events are discarded to save network bandwidth and costs.