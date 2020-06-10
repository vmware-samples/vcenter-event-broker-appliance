---
layout: docs
toc_id: contribute-functions
title: VMware Event Broker Appliance - Building Functions
description: Building Functions
permalink: /kb/contribute-functions
cta:
 title: Have a question? 
 description: Please check our [Frequently Asked Questions](/faq) first.
---

# Writing your own functions

The VMware Event Broker Appliance uses OpenFaaS as a Function-as-a-Service (FaaS) platform. If you are looking to understand the basics of functions, start [here](functions).

You can also get started quickly with these quickstart [templates](https://github.com/pksrc/vebafn){:target="_blank"}.

## Instructions

> **ASSUMPTION:** The following steps assume VMware Event Broker Appliance has been [installed (configured with OpenFaaS)](install-openfaas) and is running.


* Create a directory for your function and set up the secret config file

  ```toml
  # vcconfig.toml contents
  # replace with your own values and use a dedicated user/service account with permissions to tag VMs if possible
  [vcenter]
  server = "VCENTER_FQDN/IP"
  user = "tagging-admin@vsphere.local"
  password = "DontUseThisPassword"
  
  [tag]
  urn = "urn:vmomi:InventoryServiceTag:019c0a9e-0672-48f5-ac2a-e394669e2916:GLOBAL" # replace the actual urn of the tag
  action = "attach" # tagging action to perform, i.e. attach or detach tag
  ```

* Login to OpenFaaS and save the secret with `faas-cli secret create vcconfig --from-file=vcconfig.toml`
* Grab the desired language template (there are multiple ways)
  * The first way is with `faas template pull` and see it in `faas new --list`.
  * The second way is to look through the OpenFaas-Incubator, for example, [openfaas-incubator](https://github.com/openfaas-incubator/golang-http-template.git). If found there, retrieve it with `faas template pull https://github.com/openfaas-incubator/golang-http-template`.
  * The third way is `faas template store pull <language-template>`
  * An alternative to templates is not to use them, and make your own Dockerfile. Optionally, after doing that, you can make your own template.
* Create scaffold for the function: `faas-cli new --lang <language template> faas-hello-world --prefix="<docker hub user name>"`.
* Make changes inside scaffold
  * A directory called `faas-hello-world` should be created and within it, should be a file called `handler.go`, except the extension should be appropriate for your choice of language. Edit that file to make a new function.
  * Open and edit the `faas-hello-world.yml` provided. Change provider > gateway and functions >annotations > topic as per your environment/needs. Here is an example for Go:

  ```yaml
  provider:
    name: openfaas
    gateway: https://VEBA_FQDN_OR_IP # replace with your vCenter Event Broker Appliance environment
  functions:
    faas-hello-world:
      lang: golang-http
      handler: ./faas-hello-world
      image: fgold/faas-hello-world:latest
      environment:
        write_debug: true
        read_debug: true
      secrets:
        - vcconfig # leave as is unless you changed the name during the creation of the vCenter credentials secrets above
      annotations:
        topic: vm.powered.on # or drs.vm.powered.on in a DRS-enabled cluster
  ```

* Build the faas function with `faas-cli up -f faas-hello-world.yml`.
  * For the Golang-http template, Build the faas function with `faas-cli up -f faas-hello-world.yml --build-arg GO111MODULE=on`

### Run the Function in VEBA

* Run `faas-cli deploy -f hello-world.yml --tls-no-verify` to deploy the function. It doesn't have to be run on the local machine; it can be run on the machine that is hosting the VEBA appliance.
* Try to trigger the function with a vCenter event.

## Coding - Best Practices

Compared to writing repetitive boilerplate logic to handle vCenter events, the VMware Event Broker Appliance powered by OpenFaaS makes it remarkable easy to consume and process events with minimal code required.

However, as outlined in previous sections in this guide, there are still some best practices and pitfalls to be considered when it comes to messaging in a distributed system. The following list tries to provide guidance for function authors. Before applying them thoroughly think about your problem statement and whether all of these recommendations apply to your specific scenario.

<!-- TODO: add more stuff from AWS? -->

<!-- omit in toc -->
### Single Responsibility Principle

Avoid writing huge function handlers. Instead of describing a huge workflow in your function or using long if/else/switch statements to deal with any type of event, consider breaking your problem up into smaller pieces (functions). This makes your code cleaner, easier to understand/contribute to and maintainable. As a result, your function will likely run faster and return early, avoiding undesired blocking behavior.

Single Responsibility Principle (SRP) is the philosophy behind the UNIX command line tools. "Do one job and do it well". Solve complex problems by breaking them down with composition where the output of one program becomes the input of the next program. 

> **Note:** Generally, workflows should not be handled in functions but by workflow engines, such as vRealize Orchestrator (vRO). vRO and the VMware Event Broker Appliance work well together, e.g. by triggering workflows from functions via the vRO REST API. Upon completion, or for intermediary steps, vRO might call back into the appliance and leverage other functions for lightweight execution handling.

<!-- omit in toc -->
### Deterministic Behavior

Simply speaking, given the same input your function should always produce the same output for predictability and consistency. There's always exceptions to the rule, e.g. when dealing with time(stamps) or leveraging random number generators within your function body.

> **Note:** Whenever you lookup data in the event payload received when your function is invoked, make sure to check for missing/"NULL" keys to avoid your code from throwing an unhandled exception - or worse incorrectly interpreting (missing) data. Senders might retry invoking your function with this message, leading to an endless loop if not handled correctly.

<!-- omit in toc -->
### Keep Functions slim and up to date

Not only for security reasons should you keep your function (and dependencies, such as libraries) up to date with patches. Patches might also include performance improvements which your code immediately benefits from.

> **Note:** Since functions in the VMware Event Broker Appliance are deployed as container images, consider using a registry that supports image scanning such as [VMware Harbor](https://goharbor.io/).

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

As discussed in earlier sections of this guide, the VMware Event Broker Appliance currently does not support retrying function invocation on failure/timeout, and also does not persist events for redelivery/re-drive.

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

Although as of today the VMware Event Broker Appliance does not attempt to redeliver a message ("at least once" delivery, see [message delivery guarantees](#message-delivery-guarantees)) depending on the complexity of your function workflow and involved (external) components, message duplication might still be a concern. Your function logic or the receiving downstream system should be able to detect and deal with duplicate messages to prevent data consistency issues or unwanted [side effects](#dealing-with-side-effects).

To support idempotency checks, the VMware Event Broker Appliance [event payload](#the-event-specification) provides fields which can be used to detect duplicates. It is usually sufficient to use a combination of the event "id" and "subject" or "source" fields in the JSON message body to construct and persist a unique message key in a database (or cache) for lookups:

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
