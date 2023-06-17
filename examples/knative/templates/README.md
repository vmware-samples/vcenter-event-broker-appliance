# CloudEvent PowerShell and PowerCLI Base Container Images

* [server.ps1](server.ps1) - PowerShell HTTP Listener for handling function invocation
* [Dockerfile.ps](Dockerfile.ps) - Dockerfile for Base PowerShell Image
* [Dockerfile.pcli](Dockerfile.pcli) - Dockerfile Base PowerCLI Image which builds on top of `Dockerfile.ps`

# Run

Pre-built base PowerShell Image:

* us.gcr.io/daisy-284300/veba/ce-ps-base:1.5

Pre-built base PowerCLI Image:

* us.gcr.io/daisy-284300/veba/ce-pcli-base:1.5
# Build

Build Base PowerShell Image
```console
docker build -t <docker-username>/ce-ps-base:<version> -f Dockerfile.ps .
```

Build Base PowerCLI Image

```console
docker build -t <docker-username>/ce-pcli-base:<version> -f Dockerfile.pcli .
```