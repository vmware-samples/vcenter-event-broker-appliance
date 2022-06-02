---
layout: docs
toc_id: deploy-event-router-kind
title: Deploy Event Router to Kind
description: Deploy Event Router to Kind
permalink: /kb/deploy-event-router-kind
---

# Installation of VMware Event Router with `kind`

The following steps describe the installation of the [VMware Event
Router](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/vmware-event-router)
in a local [kind](https://kind.sigs.k8s.io/) cluster and [Knative](https://knative.dev/) environment.

The steps assume a Mac OSX environment but the links provide resources to
install the components for other platforms.

## Install Knative with [KonK](https://github.com/csantanapr/knative-kind).

Requires:

- [Docker](https://www.docker.com/products/docker-desktop)
- [kind](https://github.com/kubernetes-sigs/kind)
- kubectl

```console
export KIND_CLUSTER_NAME=kind
curl -sL get.konk.dev | bash
```

## Install Knative CLI `kn` 

[kn](https://knative.dev/docs/client/install-kn/) lets you work with Knative
resources, e.g. services, brokers, etc. instead of using kubectl.

```console
brew install knative/client/kn

kn version
Version:      v0.25.1
Build Date:   2021-09-01T17:15:33Z
Git Revision: 99e2c927
Supported APIs:
* Serving
  - serving.knative.dev/v1 (knative-serving v0.25.0)
* Eventing
  - sources.knative.dev/v1 (knative-eventing v0.25.0)
  - eventing.knative.dev/v1 (knative-eventing v0.25.0)
```

## Install Sockeye (recommended)

[`Sockeye`](https://github.com/n3wscott/sockeye/) lets you view incoming events in the browser, which can be helpful with troubleshooting.

```console
kubectl apply -f https://github.com/n3wscott/sockeye/releases/download/v0.7.0/release.yaml
service.serving.knative.dev/sockeye created
```

Open the Sockeye UI:

```console
kn service list
NAME      URL                                       LATEST          AGE    CONDITIONS   READY   REASON
hello     http://hello.default.127.0.0.1.nip.io     hello-00001     14m    3 OK / 3     True
sockeye   http://sockeye.default.127.0.0.1.nip.io   sockeye-00001   6m4s   3 OK / 3     True

open http://sockeye.default.127.0.0.1.nip.io
```

💡 The KonK stack uses `nip.io` so you can access Knative services from your
local machine via the shown URLs.

To avoid missing an event, disable scale-to-zero for the `Sockeye` Knative `service`. Note that this
might only happen because we're using an in-memory `broker` using goroutines as
channels which is not intended for production use.

```console
kn service update --scale 1 sockeye
Updating Service 'sockeye' in namespace 'default':

  0.032s The Configuration is still working to reflect the latest desired specification.
  4.340s Traffic is not yet migrated to the latest revision.
  4.411s Ingress has not yet been reconciled.
  4.470s Waiting for load balancer to be ready
  4.611s Ready to serve.

Service 'sockeye' updated to latest revision 'sockeye-00002' is available at URL:
http://sockeye.default.127.0.0.1.nip.io
```

## Install vCenter Simulator (Event Provider)

To trigger events, e.g. from vSphere, we can use [`vcsim`](https://github.com/vmware/govmomi/tree/master/vcsim).

```console
cat << EOF | kubectl create -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vcsim
spec:
  selector:
    matchLabels:
      app: vcsim
  template:
    metadata:
      labels:
        app: vcsim
    spec:
      containers:
        - name: vcsim
          image: vmware/vcsim:latest
          args: ["/vcsim", "-l", ":8989"]
          ports:
            - name: https
              containerPort: 8989

---
apiVersion: v1
kind: Service
metadata:
  name: vcsim
spec:
  selector:
    app: vcsim
  ports:
    - name: https
      port: 443
      targetPort: 8989
EOF

deployment.apps/vcsim created
service/vcsim created
```

## Install `govc`

[`govc`](https://github.com/vmware/govmomi/tree/master/govc) is used to perform
operations against the (simulated) vCenter, e.g. powering off a virtual machine
which will trigger a corresponding event.

```console
brew install govc

govc about
govc: specify an ESX or vCenter URL
```

## Install `Helm`

The easiest way to deploy the VMware Event Router is via
[Helm](https://helm.sh/).

```console
brew install helm

helm version
version.BuildInfo{Version:"v3.7.1", GitCommit:"1d11fcb5d3f3bf00dbe6fe31b8412839a96b3dc4", GitTreeState:"clean", GoVersion:"go1.17.2"}
```

## Add the VMware Event Router Helm Repo

```console
helm repo add vmware-veba https://projects.registry.vmware.com/chartrepo/veba

# update index in case the repo was already installed
helm repo update

# ignore the index error message which is due to the Harbor environment
helm search repo event-router
index.go:346: skipping loading invalid entry for chart "event-router" "release-0.5.0" from /Users/mgasch/Library/Caches/helm/repository/vmware-veba-index.yaml: validation: chart.metadata.version "release-0.5.0" is invalid
NAME                            CHART VERSION   APP VERSION     DESCRIPTION
vmware-veba/event-router        v0.7.0          v0.7.0          The VMware Event Router is used to connect to v...
```

## Deploy VMware Event Router

Before installing the VMware Event Router create a configuration (Helm
"override") file (passed via `-f -` and read from stdin).

The following commands assume a Knative `broker` in the `default` Kubernetes
namespace with the name `example-broker`. As of writing this document these are
the values used by the KonK stack. You can verify this via:

```console
kubectl get brokers -A
NAMESPACE   NAME             URL                                                                               AGE   READY   REASON
default     example-broker   http://broker-ingress.knative-eventing.svc.cluster.local/default/example-broker   28m   True
```

It is also assumed that you created the `vcsim` deployment as shown above,
otherwise change the address and settings if you have done a custom deployment.

```console
# create the deployment "router" in the "vmware-system" namespace
cat << EOF | helm install -n vmware-system --create-namespace -f - router vmware-veba/event-router
eventrouter:
  config:
    logLevel: debug
  vcenter:
    address: https://vcsim.default.svc.cluster.local
    username: user
    password: pass
    insecure: true # ignore TLS certs
  eventProcessor: knative
  knative:
    destination:
      ref:
        apiVersion: eventing.knative.dev/v1
        kind: Broker
        name: example-broker
        namespace: default
EOF
```

Verify the `router` is correctly running, otherwise make sure you followed the steps as described above:

```console
kubectl -n vmware logs deploy/router

 _    ____  ___                            ______                 __     ____              __
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/

2021-11-19T13:37:34.243Z        WARN    [VCENTER]       vcenter/vcenter.go:126  using potentially insecure connection to vCenter        {"address": "https://vcsim.default.svc.cluster.local", "insecure": true}
2021-11-19T13:37:34.243Z        INFO    [MAIN]  router/main.go:114      connecting to vCenter   {"commit": "82df3ff", "version": "v0.7.0", "address": "https://vcsim.default.svc.cluster.local"}
2021-11-19T13:37:34.262Z        INFO    [KNATIVE]       injection/injection.go:61       Starting informers...
2021-11-19T13:37:34.263Z        INFO    [MAIN]  router/main.go:169      created Knative processor       {"commit": "82df3ff", "version": "v0.7.0", "sink": "http://broker-ingress.knative-eventing.svc.cluster.local/default/example-broker"}
2021-11-19T13:37:34.374Z        WARN    [METRICS]       metrics/server.go:59    disabling basic auth: no authentication data provided
2021-11-19T13:37:34.375Z        INFO    [METRICS]       metrics/server.go:98    starting metrics server {"address": "http://0.0.0.0:8082/stats"}
2021-11-19T13:37:34.375Z        INFO    [VCENTER]       vcenter/vcenter.go:213  checkpointing disabled, setting begin of event stream   {"beginTimestamp": "2021-11-19 13:37:34.3756582 +0000 UTC"}
2021-11-19T13:37:35.000Z        DEBUG   [VCENTER]       vcenter/vcenter.go:313  no new events, backing off      {"delaySeconds": 1}
2021-11-19T13:37:36.001Z        DEBUG   [VCENTER]       vcenter/vcenter.go:313  no new events, backing off      {"delaySeconds": 2}
```

💡 You can easily remove the Helm installation with `helm -n vmware-system uninstall
router`.

## Trigger an Event

Trigger an event and observe the output in the `Sockeye` UI (browser).

**Notes:**

- If you reload the `Sockeye` UI it will discard previous events (stateless)
- Since we're using an in-memory Knative Broker (non-production), in **rare**
  cases, some events are lost - if you feel the event should have gone through,
  please retry by triggering the event again

Create a Knative [`Trigger`](https://knative.dev/docs/eventing/broker/triggers/)
to subscribe to the `Broker` for (all) incoming events:

```console
# assumes example-broker in default namespace
kn trigger create sockeye --broker example-broker --sink ksvc:sockeye
Trigger 'sockeye' successfully created in namespace 'default'.

# conditions must show 5/5 indicating READY
kn trigger list
NAME            BROKER           SINK                 AGE   CONDITIONS   READY   REASON
hello-display   example-broker   ksvc:hello-display   53m   5 OK / 5     True
sockeye         example-broker   ksvc:sockeye         10s   5 OK / 5     True
```

💡 Optionally a filter to only subscribe to specific events can be specified.

Open the `Sockeye` UI:

```console
open http://sockeye.default.127.0.0.1.nip.io
```

In a separate terminal create a Kubernetes port-forwarding so we can use `govc` to connect to `vcsim` running inside Kubernetes.

```console
kubectl port-forward deploy/vcsim 8989:8989
Forwarding from 127.0.0.1:8989 -> 8989
Forwarding from [::1]:8989 -> 8989
```

Open another terminal to trigger an event. First, set `govc` environment
variables (connection):

```console
export GOVC_INSECURE=true
export GOVC_URL=https://user:pass@127.0.0.1:8989/sdk

# verify govc connects correctly by printing the vcsim inventory
govc ls
/DC0/vm
/DC0/host
/DC0/datastore
/DC0/network
```

Trigger an event and observe the output in `Sockeye`:

```console
govc vm.power -off /DC0/vm/DC0_H0_VM0
Powering off VirtualMachine:vm-54... OK
```

`Sockeye` should show a `VmStoppingEvent` followed by a `VmPoweredOffEvent`.

<img src="./img/sockeye.png" width="100%" align="center" class="border m-1 p-1"/>

If you don't see any output, make sure you followed all steps above, with
correct naming and that all resources (`broker`, `trigger`, `service`, `router`,
etc.) are in a `READY` state.