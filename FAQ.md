<!-- omit in toc -->
# About 

This page provides answers to common questions around the vCenter Event Broker Appliance architecture, event handling and best practices for building functions.

Feel free to raise issues/file pull requests to this Github repository to help us improve the appliance and the documentation. If in doubt you can also reach out to us on Slack [#vcenter-event-broker-appliance](https://vmwarecode.slack.com/archives/CQLT9B5AA), which is part of the [VMware {Code}](https://code.vmware.com/web/code/join) Slack instance.

## Table of Content

- [Architecture](#architecture)
- [Event Handling](#event-handling)
  - [Message Delivery Guarantees](#message-delivery-guarantees)
  - [The Event Specification](#the-event-specification)
- [Functions](#functions)
  - [Getting Started](#getting-started)
  - [Naming and Version Control](#naming-and-version-control)
  - [Invocation](#invocation)
  - [Code Best Practices](#code-best-practices)

# Architecture

Even though the vCenter Event Broker Appliance is instantiated as a single running virtual machine, internally it's components follow a [microservices architecture](architecture.md) running on Kubernetes. The individual services communicate via TCP/IP network sockets. Most of the communication is performed internally in the appliance so the chance of losing network packets is reduced. 

However, in case of a component being unavailable (crash-loop, overloaded and slow to respond) communication might be impacted and so it's important to understand the communication flow as depicted further below (TODO). To avoid the risk of blocking remote calls, which could render the whole system unusable, sensible default timeouts are applied which can be fine-tuned if needed.

Kubernetes is a great platform and foundation for building highly available distributed systems. Even though we currently don't make use of its multi-node clustering capabilities (i.e. scale out), Kubernetes provides a lot of benefits to developers and users. Its self-healing capabilities continuously watch the critical vCenter Event Broker Appliance components and user-deployed functions and trigger restarts when necessary.

Kubernetes and its dependencies, such as the Docker, are deployed as systemd units. This addresses the "who watches the watcher" problem in case the Kubernetes node agent (kubelet) or Docker container runtime crashes.

> **Note:** We are considering to use Kubernetes' cluster capabilities in the future to provide increased resiliency (node crashes), scalability (scale out individual components to handle higher load) and durability (replication and persistency). The downside is the added complexity of deploying and managing a multi-node vCenter Event Broker Appliance environment.

Currently one OpenFaaS vcenter-connector is deployed per appliance (1:1 mapping). That means, only one vCenter event stream can be processed per appliance. We are evaluating options to support multiple vCenter environments per appliance (scale up) or alternatively support multi-node appliance deployments (scale out), which might be required in large deployments (performance, throughput). 

> **Note:** We have not done any extensive performance and scalability testing to understand the limits of the single appliance model.

# Event Handling

As described in the architecture section [above](#architecture) due to the microservices architecture used in the vCenter Event Broker Appliance one always has to consider message delivery problems such as timeouts, delays, reordering, loss. These challenges are fundamental to [distributed systems](https://github.com/papers-we-love/papers-we-love/blob/master/distributed_systems/a-note-on-distributed-computing.pdf) and must be understood and considered by function authors.

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
- The producer immediately after receiving the acknowledgement crashes

> **Note:** For our explanation it doesn't really matter whether the packet (message) actually leaves the machine or the destination (consumer) is on the same host. Of course, having a physical network in between the actors increases the chances of [messaging failures](https://queue.acm.org/detail.cfm?id=2655736). The network protocol in use was intentionally left unspecified. 

One of the following message delivery semantics is typically used to describe the messaging characteristics of a particular distributed system, such as the vCenter Event Broker Appliance:

- At most once semantics: a message will be delivered once or not at all to the consumer
- At least once semantics: a message will be delivered once or multiple times to the consumer
- Exactly once semantics: a message will be delivered exactly once to the consumer

> **Note:** Exactly once semantics is not supported by all messaging systems as it requires significant engineering effort to implement. It is considered the holy grail in messaging while at the same time being a highly [debated](https://medium.com/@jaykreps/exactly-once-support-in-apache-kafka-55e1fdd0a35f) topic.

As of today the vCenter Event Broker Appliance guarantees at most once delivery. While this might sound like a huge limitation in the appliance (and it might be, depending on your use case) in practice the chances for message delivery failures are/can be reduced by:

- Using TCP/IP as the underlying communication protocol which provides certain ordering (sequencing), back-pressure and retry capabilities at the transmission layer (default in the appliance)
- Using asynchronous function [invocation](#invocation) (defaults to "off", i.e. "synchronus", in the appliance) which internally uses a message queue for event processing
- Following [best practices](#code-best-practices) for writing functions

> **Note:** The vCenter Event Broker Appliance currently does not persist (to disk) or retry event delivery in case of failure during function invocation or upstream (external system, such as Slack) communication issues. For introspection and debugging purposes invocations are logged to standard output by the OpenFaaS vcenter-connector ("sync" invocation mode) or OpenFaaS queue-worker ("async" invocation mode).

We are currently investigating options to support at least once delivery semantics. However, this requires significant changes to the OpenFaaS vcenter-connector, such as:

- Tracking and checkpointing (to disk) successfully processed vCenter events (stream history position)
- Buffering events in the connector (incl. queue management to protect from overflows)
- Raising awareness (docs, tutorials) for function authors to deal with duplicated, delayed or out of order arriving event messages
- High-availability deployments (active-active/active-passive) to continue to retrieve the event stream during appliance downtime (maintenance, crash)
- Describe mitigation strategies for data loss in the appliance (snapshots, backups)

## The Event Specification

> **Note:** WIP, this new event spec will be a feature in an upcoming release of the appliance

The event payload structure used by the vCenter Event Broker Appliance has been significantly enriched since the beginning. It mostly follows the [CloudEvents](https://github.com/cloudevents/sdk-go/blob/master/pkg/cloudevents/eventcontext_v1.go) specification (v1), deviating only in some small cases (type definitions). The current data content type which is sent as payload when invoking a function is JSON.

The following example shows the event structure (trimmed for better readability):

```json
{
    "id": "6da664a7-7ad1-4b7a-b97f-8f7c75eae75a",
    "source": "10.0.10.1",
    "specversion": "1.0",
    "type": "com.github.openfaas-incubator.openfaas-vcenter-connector.vm.powered.on",
    "subject": "VmPoweredOnEvent",
    "time": "2019-12-08T10:57:35.596934Z",
    "data": {
        "Key": 9420,
        "CreatedTime": "2019-12-08T10:57:27.915136Z",
        [...]
    },
    "datacontenttype": "application/json"
}
```

> **Note:** This is not the event as emitted by vCenter. The appliance, using the OpenFaaS vcenter-connector, wraps the corresponding vCenter event (as seen in "data") into its own event structure.

`id:` The unique ID ([UUID](https://tools.ietf.org/html/rfc4122)) of the event

`source:` The vCenter emitting the embedded vSphere event (FQDN resolved when available)

`specversion:` The event specification the appliances uses (can be used for schema handling)

`type:` The canonical name of the event in "." dot notation (including the emitter, i.e. OpenFaaS vcenter-connector) 

`subject:` The vCenter event name (CamelCase)

`time:` Timestamp when this event was produced by the appliance

`data:` Original vCenter event

`data.Key:` Monotonically increasing value set by vCenter (the lower the key, the older the message as being created by vCenter)

`data.CreatedTime:` When the embedded event was created by vCenter

`datacontenttype:` Encoding used (JSON)

Please see the section on function [best practices](#code-best-practices) below how you can make use of these fields for advanced requirements.

# Functions

## Getting Started

The vCenter Event Broker Appliance uses OpenFaaS as a Function-as-a-Service (FaaS) platform. Alex Ellis, the creator of OpenFaaS, and the community have put together comprehensive documentation and workshop materials to get you started with writing your first functions:

- [Your first OpenFaaS Function with Python](https://docs.openfaas.com/tutorials/first-python-function/)
- [OpenFaaS Workshop](https://docs.openfaas.com/tutorials/workshop/)

Advanced users who directly want to jump into VMware vSphere-related function code might want to check out the examples we provide in this repository [here](examples/README.md).

## Naming and Version Control

When it comes to authoring functions, it's important to understand how the different fields in the OpenFaaS function's stack definition, e.g. `stack.yml`, are used throughout the appliance. Let's take the following excerpt as an example:

```yaml
# stack.yaml snippet
[...]
functions:
  pytag-fn:
    lang: python3
    handler: ./handler
    image: embano1/pytag-fn:0.2
```

`pytag-fn:` The name of the function used by OpenFaaS as the canonical name and identifier throughout the lifecycle of the function. Internally this will be the name used by Kubernetes to run the function as a Kubernetes deployment.

<!-- TODO: clarify deployment/pod via OpenFaaS -->

The value of this field:

- must not conflict with an existing function
- should not contain special characters, e.g. "$" or "/"
- should represent the intent of the function, e.g. "tag" or "tagging"
- may use a major version suffix, e.g. "pytag-fn-v3" in case of breaking changes/when multiple versions of the function need to run in parallel for backwards compatibility

`image:` The name of the resulting container image following Docker naming conventions `"<repo>/<image>:<tag>"`. OpenFaaS uses this field during the build and deployment phases, i.e. `faas-cli [build|deploy]`. Internally this will be the image pulled by Kubernetes when creating the function.

The value of this field:

- must resolve to a valid Docker container name (see convention above)
- should reflect the name of the function for clarity
- should use a tag other than `"latest"`, e.g. `":0.2"` or `":$GIT_COMMIT"`
- should be updated whenever changes to the function logic are made (before `faas-cli [build|deploy]`)
  - avoids overwriting the existing container image which ensures audibility and eases troubleshooting
  - supports common CI/CD version control flows
  - changing the tag is sufficient


> **Note:** `functions` can contain multiple functions described as a list in YAML (not shown here).

## Invocation

Functions in OpenFaaS can be invoked synchronously or asynchronously:

`synchronous:` The function is called and the caller, e.g. OpenFaaS vcenter-connector, waits until the function returns (successful/error) or the timeout threshold is hit.

`asynchronous:` The function is not directly called. Instead, HTTP status code 202 ("accepted") is returned and the request, including the event payload, is stored in a [NATS Streaming](https://docs.nats.io/nats-streaming-concepts/intro) queue. One or more "queue-workers" process the queue items.

If you directly invoke your functions deployed in the appliance you can decide which invocation mode is used (per function). More details can be found [here](https://github.com/openfaas/workshop/blob/master/lab7.md).

> **Note:** The vCenter Event Broker appliance by default uses synchronous invocation mode. If you experience performance issues due to long-running/slow/blocking functions, consider running the OpenFaaS vcenter-connector in asynchronous mode (`-async` flag in the Kubernetes deployment manifest, TODO).

## Code Best Practices

Compared to writing repetitive boilerplate logic to handle vCenter events, the vCenter Event Broker Appliance powered by OpenFaaS makes it remarkable easy to consume and process events with minimal code required.

However, as outlined in previous sections in this guide, there are still some best practices and pitfalls to be considered when it comes to messaging in a distributed system. The following list tries to provide guidance for function authors. Before applying them thoroughly think about your problem statement and whether all of these recommendations apply to your specific scenario.

<!-- TODO: add more stuff from AWS? -->

<!-- omit in toc -->
### Single Responsibility Principle

Avoid writing huge function handlers. Instead of describing a huge workflow in your function or using long if/else/switch statements to deal with any type of event, consider breaking your problem up into smaller pieces (functions). This makes your code cleaner, easier to understand/contribute to and maintainable. As a result, your function will likely run faster and return early, avoiding undesired blocking behavior.

Single Responsibility Principle (SRP) is the philosophy behind the UNIX command line tools. "Do one job and do it well". Solve complex problems by breaking them down with composition where the output of one program becomes the input of the next program. 

> **Note:** Generally, workflows should not be handled in functions but by workflow engines, such as vRealize Orchestrator (vRO). vRO and the vCenter Event Broker Appliance work well together, e.g. by triggering workflows from functions via the vRO REST API. Upon completion, or for intermediary steps, vRO might call back into the appliance and leverage other functions for lightweight execution handling.

<!-- omit in toc -->
### Deterministic Behavior

Simply speaking, given the same input your function should always produce the same output for predictability and consistency. There's always exceptions to the rule, e.g. when dealing with time(stamps) or leveraging random number generators within your function body.

> **Note:** Whenever you lookup data in the event payload received when your function is invoked, make sure to check for missing/"NULL" keys to avoid your code from throwing an unhandled exception - or worse incorrectly interpreting (missing) data. Senders might retry invoking your function with this message, leading to an endless loop if not handled correctly.

<!-- omit in toc -->
### Keep Functions slim and up to date

Not only for security reasons should you keep your function (and dependencies, such as libraries) up to date with patches. Patches might also include performance improvements which your code immediately benefits from.

> **Note:** Since functions in the vCenter Event Broker Appliance are deployed as container images, consider using a registry that supports image scanning such as [VMware Harbor](https://goharbor.io/).

Try to reduce the container image size by using a container optimized function image (template) and use Docker [multi-stage](https://docs.docker.com/develop/develop-images/multistage-build/) builds in your [custom](https://towardsdatascience.com/going-serverless-with-openfaas-and-golang-building-optimized-templates-730991084443) OpenFaaS templates. Remove unused libraries/files which unnecessarily bloat your image, leading to longer download and startup times.

<!-- omit in toc -->
### Keep Functions "warm" - if possible

Most OpenFaaS function templates support the [`"http"` mode](https://github.com/openfaas-incubator/of-watchdog#1-http-modehttp) for calling your function handler. This prevents the function execution stack `"main()"` from terminating and enables function authors to persist state, such as connections, in memory for faster access and reuse.

This is especially useful when dealing with limited resources such as database or vCenter connections. Another benefit is that connections don't have to be newly established but can be reused. Pseudo-code example:

```python
# db defined outside function handler
db = setup_db(user, password, db_server)
def handle(req):
  event_body = req.get("data")
  db.put(event_body)
```

> **Note:** Your connection/session library should support "keep alive" to periodically send a heartbeat/ping to the remote server and keep the connection open (tokens fresh).

<!-- omit in toc -->
### Return early/defer or externalize Work

Your primary goal should be to avoid long-running functions (minutes) as much as possible. The longer your function runs, the more things can go wrong and you might have to start from scratch (which might not be possible without additional persistency safeties in your logic). 

Usually that's an indicator that your function can be further broken down into smaller steps or could be better handled with a workflow engine, see [Single Responsibility Principle](#single-responsibility-principle) above.

If you can't avoid long-running functions an option is to persist the event payload (if it's important) to a durable (external) queue or database and use dedicated workers to process these items. The [OpenFaaS kafka-connector](https://github.com/openfaas-incubator/kafka-connector) can be a suitable approach.

<!-- omit in toc -->
### Dealing with Side Effects

A side effect is an irreversible action, such as sending an email or printing a log statement to standard output. Since generally you cannot avoid these, it's best to move the related logic for critical side effects to the end of the function handler (if possible). Memoizing state to prevent duplicate execution can be a useful approach to avoid undesired side effects, such as sending an email twice (also see section on idempotency below). Pseudo-code below:

```python
db = setup_db(user, password, db_server)
def handle(req):
  subject = req.get("subject")
  event_id = req.get("id")
  processed = db.get(event_id, "event_table")
  if not processed and subject == "VmPoweredOffEvent":
    send_email("alert", req)
    db.write(event_id, "event_table")
```

> **Note:** Strictly speaking the pseudo-code above is flawed since `send_email` and `db.write` are not part of (the same) atomic operation (transaction). The [outbox pattern](https://debezium.io/blog/2019/02/19/reliable-microservices-data-exchange-with-the-outbox-pattern/), delayed processing and/or compensating transaction such as [Sagas](https://dzone.com/articles/distributed-sagas-for-microservices) are technical solutions for such complex requirements.

<!-- omit in toc -->
### Persistency and Retries

As discussed in earlier sections of this guide, the vCenter Event Broker Appliance currently does not support retrying function invocation on failure/timeout, and also does not persist events for redelivery/re-drive.

A workaround is to persist the event to an external (durable) datastore or queue and consume/process from there. If this fails a log message can be produced with debugging information (critical event payload) or the event sent to a backup system, e.g. dead letter queue (DLQ).

>**Note:** Strictly speaking this does not address the appliance-internal scenario where the OpenFaaS vcenter-connector might not be able to invoke your function (resource busy, unavailable, etc.) but addresses common network communication issues when making outbound calls from the appliance.

If your function executes quickly, retrying within the function might be a viable approach as well (retry three times with an increasing backoff delay). Pseudo-code:

```python
def handle(req):
  success = False
  failures = 0
  while not success:
    success = send_event(req.get("data"))
    if not success:
      failures += 1
      if failures > 3:
        return
      print(f'failure, retrying after {failures * 3} seconds')
      sleep(3 * failures)
```

<!-- omit in toc -->
### Idempotency (Message Deduplication)

Although as of today the vCenter Event Broker Appliance does not attempt to redeliver a message ("at least once" delivery, see [message delivery guarantees](#message-delivery-guarantees)) depending on the complexity of your function workflow and involved (external) components, message duplication might still be a concern. Your function logic or the receiving downstream system should be able to detect and deal with duplicate messages to prevent data consistency issues or unwanted [side effects](#dealing-with-side-effects).

To support idempotency checks, the vCenter Event Broker Appliance [event payload](#the-event-specification) provides fields which can be used to detect duplicates. It is usually sufficient to use a combination of the event "id" and "subject" or "source" fields in the JSON message body to construct and persist a unique message key in a database (or cache) for lookups:

```json
{
  [...]
  "id":"0058c998-cc0f-49ca-8cc3-1b60abf5957c",
  "source":"10.160.94.63",
  "subject":"UserLogoutSessionEvent"
}
```

> **Note:** The "id" field is a UUID which, practically speaking, is guaranteed to be unique per event (even across multiple appliances). "Source" or "subject" can be used for faster indexing/lookups in tables or caches.


<!-- omit in toc -->
### Out of Order Message Arrival

Even though unlikely due to the underlying TCP/IP guarantees, but nevertheless possible in specific environments or deployments - dealing with out of order message arrival in your function/downstream logic might be a requirement. 

Therefore, your function or downstream system can use the vCenter event "Key", a monotonically increasing value set by vCenter, to discard late arriving messages with a lower "Key" value. If your function supports "warm" invocations (see `"http"` mode described above) the value can be cached in memory or alternatively (for increased durability) persisted in an external datastore/cache such as [Redis](https://redis.io/).

```python
last_key = 0
def handle(req):
  key = req.get("data").get("Key", 0)
  if key > last_key:
    # do work
    last_key = key
```

> **Note:** Depending on your logic, it might still be desired to account for late arriving data. This is usually the case for stream processors. You might found this [paper](https://blog.acolyer.org/2015/08/21/millwheel-fault-tolerant-stream-processing-at-internet-scale/) on windowing and watermarks an interesting read.

<!-- omit in toc -->
### Support Debugging

Things will go wrong. Provide useful and correct information via logging to standard output. Example for an incorrect log statement in your code:

```python
def handle(req):
  # do something
  print('stored event in database')
  store_event(event)
```

If `store_event` fails someone troubleshooting your function will have a hard time. Either rephrase the `print` statement to "storing ..." or, better, put it after the function call. Also, consider using a structured logging library that supports consistently formatted and parsable output.

> **Note:** Avoid logging sensitive data, such as usernames, passwords, account information, etc.
