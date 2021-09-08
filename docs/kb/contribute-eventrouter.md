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

In the `Makefile` we provide `make` targets for building a binary, Docker image
and validating changes via unit tests (`make unit-test`). These tests will run
when a pull request is submitted, but in order to run them locally to verify
your changes you need to have the following bits installed:

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

Requirements: This project uses [Golang](https://golang.org/dl/) and Go
[modules](https://blog.golang.org/using-go-modules){:target="_blank"}. For
convenience a Makefile and Dockerfile are provided requiring `make` and
[Docker](https://www.docker.com/){:target="_blank"} to be installed as well.

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/vmware-event-router

# for Go versions before v1.13
export GO111MODULE=on

# defaults to build with Docker (use make binary for local executable instead)
make
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