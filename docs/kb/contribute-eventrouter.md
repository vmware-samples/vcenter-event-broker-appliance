---
layout: docs
toc_id: contribute-eventrouter
title: Building the Event Router
description: Building the Event Router
permalink: /kb/contribute-eventrouter
cta:
 title: Have a question?
 description: Please check our [Frequently Asked Questions](/faq) first.
---

# Contribute

If you would like to make modification/additions to this code base, please
follow our [CONTRIBUTION](https://vmweventbroker.io/community) guidelines first.

In the `Makefile` we provide `make` targets for building a binary and validating
changes via unit/integration tests (`make test`). These tests will run when a pull
request is submitted, but in order to run them locally to verify your changes
you need to have the following bits installed:

`make unit-test`:

- `go` tool chain
- `make`
- `gofmt`

To run the integration tests without the need to create the testbed manually use
the following script:

`./hack/run_integration_tests.sh`:

- `go` tool chain
- `jq`
- `kind`
- `docker`

# Build VMware Event Router from Source

**Note:** This step is only required if you made code changes to the Go code.

This repository uses [`ko`](https://github.com/google/ko) to build and push
container artifacts. [`goreleaser`](https://goreleaser.com/) is used to build
binary artifacts.

Requirements:

- [`go`](https://go.dev/)
- [`ko`](https://github.com/google/ko) to build images (does not require Docker)
- [`goreleaser`](https://goreleaser.com/) to build executable binaries
- `make`
- A container runtime like Docker or `cri` to run images/integration tests locally
- [`kind`](https://kind.sigs.k8s.io/) to run integration tests

For convenience a `Makefile` is provided.

```console
# from within the vmware-event-router folder
make

Usage:
  make [target]

Targets:
  help                  Display usage
  tidy                  Sync and clean up Go dependencies
  build                 Build binary
  gofmt                 Check code is gofmted
  unit-test             Run unit tests
  integration-test      Run integration tests (requires Kubernetes cluster w/ OpenFaaS or use hack/run_integration_tests.sh)
  test                  Run unit and integration tests
```

To build an image with `kind`:

```console
# only when using kind: 
# export KIND_CLUSTER_NAME=kind
# export KO_DOCKER_REPO=kind.local

export KO_DOCKER_REPO=my-docker-username
export KO_COMMIT=$(git rev-parse --short=8 HEAD)
export KO_TAG=$(git describe --abbrev=0 --tags)

# build, push and run the router in the configured Kubernetes context 
# and vmware Kubernetes namespace
ko resolve -BRf deploy/event-router-k8s.yaml | kubectl -n vmware apply -f -
```

To delete the deployment:

```console
ko -n vmware delete -f deploy/event-router-k8s.yaml
```

> **Note:** For `_test.go` files your editor (e.g. vscode) might show errors and
> not be able to resolve symbols. This is due to the use of build tags which
> `gopls` currently does [not
> support](https://github.com/golang/go/issues/29202#issuecomment-515170916). In
> vscode add this to your configuration:
>
> ```json
> "go.toolsEnvVars": {
>        "GOFLAGS": "-tags=integration,unit"
> }
> ```