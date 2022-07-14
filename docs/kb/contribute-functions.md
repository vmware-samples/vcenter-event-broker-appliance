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

# Writing your own Functions

The VMware Event Broker Appliance (VEBA) uses Knative as a Function-as-a-Service
(FaaS) platform. If you are looking to understand the basics of functions, start
[here](functions). You can also get started quickly with these quickstart
[templates](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/examples/knative){:target="_blank"}.

## Intro

This guide describes how to create a function with PowerCLI (PowerShell) to
apply a vSphere tag when a Virtual Machine is powered on.

> **Note:** The following steps assume VMware Event Broker Appliance has been
> [installed (configured with Knative)](install-knative) and is running
> correctly. Access to the Kubernetes environment in VEBA via `kubectl` is also
> assumed to be working.

A template for Knative PowerCLI functions is available in [kn-pcli-template](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/examples/knative/powercli/kn-pcli-template). If you do not want to build all of the required files from scratch, you can copy all of the files from this template directory. Follow the instructions in the [README](https://github.com/vmware-samples/vcenter-event-broker-appliance/blob/master/examples/knative/powercli/kn-pcli-template/README.md) instead of the instructions on this page.

To create a function from scratch, continue with the instructions on this page.

## Instructions

First, create a directory for your function code, credentials (implemented via
Kubernetes [secrets](https://kubernetes.io/docs/concepts/configuration/secret/))
and test files.

```bash
mkdir tag-fn && cd tag-fn
```

Before we start looking at the actual function business logic (inside
`handler.ps1`), let's discuss how `secrets`, such as credentials, are injected
and used inside a function.

### Credentials (Secrets)

There's multiple ways to inject credentials or other forms of secrets into a
VEBA function. 

The schema and encoding of your credentials is flexible and will be projected
into your function using OS environment variables (can be changed). The
decoding/interpretation of these environment variables needs to be handled in
your function code.

For example, if you need to inject a secure token (e.g. from Slack) into your
function, you can simply write that token into a plain file and create a secret
from it (steps will be shown later). This token will be projected into your
function via a (customizable) environment variable and can be read via your
standard programming/scripting language primitives.

If you have more complex secrets or credentials, e.g. a set of username,
password, server URL, etc. we recommend using a JSON structure (again in a plain
text file). After creating a secret, the information in this file will be
projected as a single environment variable string into your function. You then
need to decode (parse) that string into a proper structure, e.g.
`PSCustomObject` or `hashmap` in PowerShell, `dict` in Python or `struct` in Go.

Also, instead of creating multiple secrets or other forms of objects to hold
your configuration data, one can collapse all information into one file
(secret). This has PROs and CONs, e.g. is easier to reason about because all
information is projected into one environment variable and data structure. But
on the other hand might violate separation of concerns by mixing different
information (and data classifications) into one object.

To make this topic more tangible, the following example creates two files
(holding a simple token and somewhat more complex authentication credentials
including some other configuration data needed for this example). These files
will be used to create one Kubernetes secret (with multiple environment variable
entries) in a subsequent step.

```bash
# create a text file holding a simple token
cat << EOF > simple_token.txt
a1b2c3d4e5f6g7h8i9j0
EOF


# create a JSON file holding credentials and configuration information
cat << EOF > vc_creds.json
{
  "VCENTER_SERVER": "https://vc-prod-01.corp.local",
  "VCENTER_USERNAME" : "service-account-01",
  "VCENTER_PASSWORD" : "imInsecure",
  "VCENTER_TAG_NAME" : "my-demo-tag",
  "VCENTER_CERTIFICATE_ACTION" : "Fail"
}
EOF
```

Next, create the Kubernetes secret in the `vmware-functions` namespace using
both files as input. This requires the `kubectl` to be installed and access
(permissions) to a deployed VMware Event Broker Appliance.

```bash
kubectl -n vmware-functions create secret generic tag-secret \
--from-file=TOKEN=simple_token.txt --from-file=VC_CREDS=vc_creds.json
```

In the above command, `tag-secret` is the name of the Kubernetes secret
(which we'll reference in the function manifest YAML later). `TOKEN` and
`VC_CREDS` are the names of the environment variables holding the contents of
the respective file. This can be verify with:

```bash
# inspect the data field of the secret
kubectl -n vmware-functions get secret tag-secret -o json
{
    "apiVersion": "v1",
    "data": {
        "TOKEN": "YTFiMmMzZDRlNWY2ZzdoOGk5ajAK",
        "VC_CREDS": "ewogICJWQ0VOVEVSX1NFUlZFUiI6ICJodHRwczovL3ZjLXByb2QtMDEuY29ycC5sb2NhbCIsCiAgIlZDRU5URVJfVVNFUk5BTUUiIDogInNlcnZpY2UtYWNjb3VudC0wMSIsCiAgIlZDRU5URVJfUEFTU1dPUkQiIDogImltSW5zZWN1cmUiLAogICJWQ0VOVEVSX1RBR19OQU1FIiA6ICJteS1kZW1vLXRhZyIsCiAgIlZDRU5URVJfQ0VSVElGSUNBVEVfQUNUSU9OIiA6ICJGYWlsIgp9Cg=="
    },
    "kind": "Secret",
    "metadata": {
        "creationTimestamp": "2021-09-06T15:00:38Z",
        "name": "tag-secret",
        "namespace": "default",
        "resourceVersion": "126008",
        "uid": "03002066-96ac-47f1-a078-e058ea437f24"
    },
    "type": "Opaque"
}
```

> **Note:** The content of the individual `data` keys in the secret (`TOKEN` and
> `VC_CREDS`) is `base64` encoded.


### Dockerfile

VEBA functions are build and deployed as OCI-compliant images. In this example
we use Docker (`Dockerfile`) for this purpose.

> **Note:** Depending on your function runtime (language), creating a
> `Dockerfile` might not be required, e.g. if you're using [Cloud Native
> Buildpacks](https://buildpacks.io/), [`ko`](https://github.com/google/ko) or
> similar tools.

The VEBA community provides ready-to-use Powershell/PowerCLI base images. Create
a minimal `Dockerfile` for PowerCLI:

```bash
cat << EOF > Dockerfile
FROM projects.registry.vmware.com/veba/ce-pcli-base:latest
COPY handler.ps1 handler.ps1
CMD ["pwsh","./server.ps1"]
EOF
```

> **Note:** For convenience we use the `:latest` tag in this example which is
> discouraged (see best practices below).

The business logic of the function lives inside `handler.ps1` (and will be
created in the next step). VEBA Docker templates for Powershell/PowerCLI wrap
the handler inside an HTTP server (`server.ps1` of type
[`HttpListener`](https://docs.microsoft.com/en-us/dotnet/api/system.net.httplistener?view=net-5.0))
which accepts and validates incoming
[CloudEvents](https://www.powershellgallery.com/packages/CloudEvents.Sdk/) and
then passes the CloudEvent to the `Process-Handler` function in `handler.ps1`.

### Function (Business Logic) with PowerCLI

In an editor, create a file `handler.ps1` with the following required `Function`
definitions:

```powershell
Function Process-Init {
   [CmdletBinding()]
   param()
}

Function Process-Shutdown {
   [CmdletBinding()]
   param()
}

Function Process-Handler {
   [CmdletBinding()]
   param(
      [Parameter(Position=0,Mandatory=$true)][CloudNative.CloudEvents.CloudEvent]$CloudEvent
   )
```

`Process-Init` and `Process-Shutdown` are hooks called by the HTTP server during
startup/shutdown. These can be used to set up and gracefully shut down (on
`SIGTERM` signal) long-running operations, e.g. a vCenter connection.

`Process-Handler` is where you write your business logic, e.g. to tag a virtual
machine based on the incoming `$CloudEvent`. The full example and code can be
found [here
(Github)](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/examples/knative/powercli/kn-pcli-tag).
Please paste the code from the linked Github example `handler.ps1` into your
local function to follow along with this example.

The secrets created in the earlier steps can be accessed from within the
function. This is typically done during the `Process-Init` phase to fail early
in case of environment variable parsing errors or missing values. 

In an earlier step, we created a Kubernetes secret with two values: `TOKEN` and
`VC_CREDS`. `TOKEN` is a simple test string and thus can be easily retrieved
inside the function's environment with `$token = ${env:TOKEN}`.

`VC_CREDS` is a JSON-encoded string, thus it must be decoded ("converted") into
a
[`PSCustomObject`](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/convertfrom-json?view=powershell-7.1)
as described in the example below:

```powershell
<snip>
  try {
      $jsonSecrets = ${env:VC_CREDS} | ConvertFrom-Json
   } catch {
      throw "Kubernetes secret: $env:VC_CREDS does not look to be defined"
   }

   # Define variables for all tag secret values for ease of use in function
   $VCENTER_SERVER = ${jsonSecrets}.VCENTER_SERVER
   $VCENTER_USERNAME = ${jsonSecrets}.VCENTER_USERNAME
   $VCENTER_PASSWORD = ${jsonSecrets}.VCENTER_PASSWORD
   $VCENTER_TAG_NAME = ${jsonSecrets}.VCENTER_TAG_NAME
   $VCENTER_CERTIFICATE_ACTION = ${jsonSecrets}.VCENTER_CERTIFICATE_ACTION
   
   # not used in this example, just to show secret as plain text to env parsing
   $TOKEN = ${env:TOKEN}
<snip>
```

It is  good practice to wrap PowerShell commands into `try/catch` blocks and
handle errors appropriately, e.g. throwing an annotated or custom exception.

#### A Note on Exception Handling

The provided PowerShell/PowerCLI images in VEBA define a lose contract between
the server.ps1 and handler.ps1 when it comes to exception handling:

Exceptions thrown in `Process-Init` will immediately **terminate** the whole
function (i.e. the server/container) with exit `code 1`. VEBA will attempt to
restart failed functions with backoff logic.

Exceptions thrown in `Process-Handler` will immediately interrupt the handler
logic **for the current** (failed) CloudEvent but will **not terminate** the
server. Currently, handler exceptions are not type-checked and the server will
respond with HTTP `INTERNAL_SERVER_ERROR (500)` to the caller. The caller might
retry so it's important to make the function (handler) logic idempotent (see
notes further below on this matter).

Exceptions thrown in `Process-Shutdown` will immediately interrupt the shutdown
process. The server will terminate with a successful exit code `0` since
shutdown failures are not considered erroneous.

### Build the Container Image

In order to deploy the function to VEBA, a container image (here using Docker)
must be built and pushed to a container registry accessible for VEBA.

```bash
# adjust values to match your environment
export REGISTRY=your-docker-username/kn-pcli-tag
export TAG=1.0

# build and push image
docker build -t ${REGISTRY}:${TAG} .
docker push ${REGISTRY}:${TAG}
```

### The Function Manifest

Create a file `function.yaml` to wire all components together before deploying
it to VEBA. 

> **Note:** The following steps assume a working Knative environment using the
> default Rabbit broker. The Knative service and trigger will be installed in
> the `vmware-functions` Kubernetes namespace, assuming that the broker is also
> available there.

```bash
cat << EOF > function.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: kn-pcli-tag
  labels:
    app: veba-ui
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "1"
        autoscaling.knative.dev/minScale: "1"
    spec:
      containers:
        # as created above
        - image: your-docker-username/kn-pcli-tag:1.0
          envFrom:
            - secretRef:
                # as created above
                name: tag-secret
          env:
            # additional env vars if needed by function
            - name: FUNCTION_DEBUG
              value: "false"
```

Let's discuss some important fields here:

- `name`: a unique name for the function deployment
- annotation `autoscaling.knative.dev/minScale`: minimum instances to run (`0`
  for scale-to zero)
- annotation `autoscaling.knative.dev/maxScale`: maximum instances to run (`0`
  for undefined)
- `image`: name of the container image you created above
- `envFrom`: array with a list of `secretRef` to project keys from a secret into
  environment variables inside the function
- `env`: array with a list of additional key/value strings projected as
  environment variables inside the function

#### Create an Event Trigger

In order to have your function execute on (specific) events, e.g assign a tag
on a `DrsVmPoweredOnEvent`, we also need to create a `Trigger`.

```bash
cat << EOF >> function.yaml
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-pcli-tag-trigger
  labels:
    app: veba-ui
spec:
  broker: default
  filter:
    attributes:
      subject: DrsVmPoweredOnEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-pcli-tag
```

Let's discuss some important fields here:

- `name`: a unique name for the trigger deployment
- `broker`: register this trigger to the defined broker (`default` in VEBA)
- `filter.attributes`: an object of CloudEvent envelop attributes, e.g.
  `source`, `type` or `subject` to filter on (cannot filter on vSphere event
  payload) - no filter means retrieving **all** events
- `subscriber.ref`: if the trigger fires, where to send the CloudEvent, i.e. the
  function defined above

### Deploy the Function in VEBA

Now deploy the function to the VMware Event Broker Appliance (VEBA).

```bash
kubectl -n vmware-functions apply -f function.yaml
```

> **Note:** See the function troubleshooting
> [documentation](troubleshoot-functions) in case
> of issues.

## Coding - Best Practices

Compared to writing repetitive boilerplate logic to handle VMware events, the
VMware Event Broker Appliance powered by Knative makes it remarkable easy to
consume and process events with minimal code required.

However, as outlined in previous sections in this guide, there are still some
best practices and pitfalls to be considered when it comes to messaging in a
distributed system. The following list tries to provide guidance for function
authors. Before blindly applying them, thoroughly think about your problem
statement and whether all of these recommendations apply to your specific
scenario.

### Single Responsibility Principle

Avoid writing huge function handlers. Instead of describing a huge workflow in
your function or using long if/else/switch statements to deal with any type of
event, consider breaking your problem up into smaller pieces (functions). This
makes your code cleaner, easier to understand/contribute to and maintainable. As
a result, your function will likely run faster and return early, avoiding
undesired blocking behavior.

Single Responsibility Principle (SRP) is the philosophy behind the UNIX command
line tools. "Do one job and do it well". Solve complex problems by breaking them
down with composition where the output of one program becomes the input of the
next program. 

⚠️ Generally, workflows should not be handled in functions but
by workflow engines, such as vRealize Orchestrator (vRO). vRO and the VMware
Event Broker Appliance work well together, e.g. by triggering workflows from
functions via the vRO REST API. Upon completion, or for intermediary steps, vRO
might call back into the appliance and leverage other functions for lightweight
execution handling. Another option is coordinating workflows or task tracking
triggered by events and functions through a database with strong consistency,
e.g. a SQL database running in serializable snapshot isolation mode (SSI). 

### Deterministic Behavior

Simply speaking, given the same input to your function, it should always produce
the same output to guarantee predictability and consistency. There's always
exceptions to the rule, e.g. when dealing with time(stamps) or leveraging random
number generators within your function body, but those should be minimized
whenever possible.

Sending an email/Slack notification, persisting data in an external state store,
or printing a log line to name a few more scenarios must also be considered
non-reversible side effects that may require special attention in the code path.

A side effect is an irreversible action. Since generally you cannot avoid these,
it's best to move the related logic for critical side effects to the end of the
function handler (if possible). Memoizing state to prevent duplicate execution
can be a useful approach to avoid undesired side effects, such as sending an
email twice (also see section on idempotency below). Python pseudo-code below:

```python
db = setup_db(user, password, db_server)
def handle(cloudevent):
  subject = cloudevent.headers.get("subject")
  event_id = cloudevent.headers.get("id")
  processed = db.get(event_id, "event_table")
  if not processed and subject == "VmPoweredOffEvent":
    send_email("alert", cloudevent)
    db.write(event_id, "event_table")
```

> **Note:** Strictly speaking the pseudo-code above is flawed since `send_email`
> and `db.write` are not part of (the same) atomic operation (transaction). The
> [outbox
> pattern](https://debezium.io/blog/2019/02/19/reliable-microservices-data-exchange-with-the-outbox-pattern/),
> delayed processing and/or compensating transaction such as
> [Sagas](https://dzone.com/articles/distributed-sagas-for-microservices) are
> technical solutions for such complex requirements, if a workflow engine cannot
> be used.

⚠️ Whenever you lookup data in the event payload received when your function is
invoked, make sure to check for missing or `"NULL"` keys to avoid your code from
throwing an unhandled exception - or worse incorrectly interpreting (missing)
data. Senders might retry invoking your function with this message, leading to a
potentially endless loop if not handled correctly.

### Keep Functions slim and up to date

Not only for security reasons should you keep your function (and dependencies,
such as libraries) up to date with patches. Patches might also include
performance improvements which your code immediately benefits from.

> **Note:** Since functions in the VMware Event Broker Appliance are deployed as
> container images, consider using a registry that supports image scanning such
> as [VMware Harbor](https://goharbor.io/).

Try to reduce the container image size by using a container optimized function
image (e.g. templates provided by the VEBA community) or use
[Buildpacks](https://buildpacks.io/) or Docker
[multi-stage](https://docs.docker.com/develop/develop-images/multistage-build/)
builds for custom images. Remove unused libraries/files which unnecessarily
bloat your image, leading to longer download and startup times.

### Keep Functions "warm"

Functions deployed in VEBA in principle are HTTP servers listening for incoming
HTTP `POST` requests (CloudEvent). The server keeps running and invokes the
CloudEvent handler on every matching CloudEvent. For example the
`Process-Handler` in `handler.ps1` when using the VEBA provided
PowerShell/PowerCLI templates.

Thus, stateful logic, e.g. vCenter or database connections can be initialized
once and reused over the function's lifetime to significantly improve the
latency and throughput of functions.

Example in PowerCLI to initialize a vCenter connection only once during startup
using `Process-Init`:

```powershell
Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   try {
      $jsonSecrets = ${env:TAG_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:TAG_SECRET does not look to be defined"
   }

   # Extract all tag secrets for ease of use in function
   $VCENTER_SERVER = ${jsonSecrets}.VCENTER_SERVER
   $VCENTER_USERNAME = ${jsonSecrets}.VCENTER_USERNAME
   $VCENTER_PASSWORD = ${jsonSecrets}.VCENTER_PASSWORD
   $VCENTER_CERTIFICATE_ACTION = ${jsonSecrets}.VCENTER_CERTIFICATE_ACTION

   # Configure TLS 1.2/1.3 support as this is required for latest vSphere release
   [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor [System.Net.SecurityProtocolType]::Tls12 -bor [System.Net.SecurityProtocolType]::Tls13

   Write-Host "$(Get-Date) - Configuring PowerCLI Configuration Settings`n"
   Set-PowerCLIConfiguration -InvalidCertificateAction:${VCENTER_CERTIFICATE_ACTION} -ParticipateInCeip:$true -Confirm:$false

   Write-Host "$(Get-Date) - Connecting to vCenter Server $VCENTER_SERVER`n"

   try {
      Connect-VIServer -Server $VCENTER_SERVER -User $VCENTER_USERNAME -Password $VCENTER_PASSWORD
   } catch {
      Write-Error "$(Get-Date) - Failed to connect to vCenter Server"
      throw $_
   }

   Write-Host "$(Get-Date) - Successfully connected to $VCENTER_SERVER`n"

   Write-Host "$(Get-Date) - Init Processing Completed`n"
}
```

> **Note:** Your connection/session library should support "keep-alive" to
> periodically send a heartbeat/ping to the remote server and keep the
> connection open (tokens fresh). When the function terminates gracefully
> (SIGTERM), a shutdown handler should be used, e.g. `Process-Shutdown` in the
> provided PowerShell/PowerCLI templates.

### Return early/externalize long-running Tasks

Your primary goal should be to avoid long-running functions (minutes) as much as
possible. The longer your function runs, the more things can go wrong and you
might have to start from scratch (which might not be possible without additional
persistency and safety measures in your logic). 

Usually that's an indicator that your function can be further broken down into
smaller steps or could be better handled with a workflow engine, see [Single
Responsibility Principle](#single-responsibility-principle) above.

If you can't avoid long-running functions, an option is to persist the event
payload (if it's important) to a durable (external) queue or database and use
dedicated workers to process these items. In such cases, an external SQL
database or [Knative Channels](https://knative.dev/docs/eventing/channels/) can
be used (requires manual installation in VEBA).

### Retries and Idempotency

The VMware Event Broker Appliance provides several safety measures for message
delivery and tries to reliably deliver events to the configured event
`processor`, e.g. functions in Knative. Once an event is accepted and persisted
in the `broker`, the `broker` attempts multiple retries if the invoked function
does not return with a successful HTTP `2xx` response.

Thus, functions must be written so that they can be safely invoked
`at-least-once` (i.e. one or more times) with the same input (event). To achieve
deterministic behavior and avoid side effects (see above), message
deduplication, e.g. based on the CloudEvent `id`, should be performed.

One option is to send and persist the event to an external (durable) datastore
or queue and continue to process it from there. If this fails a log message can
be produced with debugging information (critical event payload) or the event
sent to a backup system, e.g. dead letter queue (DLQ) which can be configured in
the `broker` (manual step in VEBA).

### Out of Order Message Arrival

Even though unlikely due to the underlying TCP/IP guarantees, but nevertheless
possible depending on event delivery issues, concurrency, retries, etc. dealing
with out of order message arrival in your function/downstream logic might be a
requirement. 

Depending on the incoming event, e.g. a CloudEvent created by the `vcenter`
event `provider` in VEBA, a function can use a specific ordering key (if
available) to detect out-of-order (and even missing data) in an event stream.
These capabilities naturally require simple (external database lookup) or more
complex stateful stream processing (Kafka Streams, etc.).

> **Note:** Depending on your logic, it might still be desired to account for
> late arriving data. This is usually the case for stream processors. You might
> found this
> [paper](https://blog.acolyer.org/2015/08/21/millwheel-fault-tolerant-stream-processing-at-internet-scale/)
> on windowing and watermarks an interesting read.

The following Python pseudo-code function example uses an external database to
store the last processed `vcenter` event key to detect out-of-order arrivals. To
speed up processing, this value can be cached in memory in addition to
persisting it in an external datastore/cache such as [Redis](https://redis.io/).

```python
db = db.init()
last_key = db.load("last_key", "vc-keys-table")
def handle(cloudevent):
  key = cloudevent.get("Key", 0)
  if key >= last_key:
    # do work
    last_key = key
    db.save("last_key", last_key, "vc-keys-table")
  else:
      log.error("out of order event received")
```

### Support Debugging

There's one guarantee in distributed systems: Things will go wrong. Besides
writing safe, secure and deterministic code (see earlier sections), it's also
important to provide useful and correct debugging information by logging to
standard output (which then can be forwarded to a centrally logging system for
durability). 

Here's another pseudo-code example (Python) which does not meet these
requirements:

```python
def handle(req):
  print('stored event in database')
  store_event(event)
```

If `store_event` fails, the person troubleshooting your function (you?) will
have a hard time. 

Either rephrase the `print` statement to "storing ..." or, better, put it after
the function call.Consider using a structured logging library that supports
consistently formatted and parsable output with different log levels, e.g.
`DEBUG`, `WARN`, etc. 

Logging should be encompassed with robust exception handling, e.g. caused by
unexpected payloads, schema issues, network or database problems.

> **Note:** Avoid logging sensitive data, such as usernames, passwords, account
> information, etc.
