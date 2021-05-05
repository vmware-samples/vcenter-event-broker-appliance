---
layout: docs
toc_id: advanced-certificates
title: VMware Event Broker Appliance - Certificates
description: Updating Certificates
permalink: /kb/advanced-certificates
cta:
 description: Replacing the default self-signed TLS certificate in VMware Event Broke Appliance.
---

## Updating the TLS Certificate on VEBA

By default, the VMware Event Broker Appliance generates a self-signed TLS certificate that is used to support different web endpoints running on the appliance such as Stats (`/stats`), Status (`/status`), Logs (`/bootstrap`) and Events (`/events`). This will cause browsers to show the certificate as untrusted.

For organizations that require the use of a TLS certificate from a trusted authority, the VMware Event Broker Appliance provides an option for users to provide their certificate information during the OVF property configuration when deploying the virtual appliance.

In order to use a certificates from a trusted authority, please follow the steps outlined below.

### Assumptions

* Certificates from a trusted authority pre-downloaded onto your local desktop
    * The public/private key pair must exist before hand. The public key certificate must be .PEM encoded and match the given private key.

### Steps

In the example, the private key file is named `privateKey.key` and certificate file is named `certificate.crt`

1. Encode both the private key and the certificate file using base64 encoding.

Microsoft Windows (PowerShell)

```console
$privateKeyContent = Get-Content -Raw privateKey.key
$privateKeybase64 = [System.Convert]::ToBase64String([System.Text.Encoding]::ASCII.GetBytes($privateKeyContent))
Write-Host "Encoded Private Key:`n$privateKeybase64`n"

$certContent = Get-Content -Raw certificate.crt
$certbase64 = [System.Convert]::ToBase64String([System.Text.Encoding]::ASCII.GetBytes($certContent))
Write-Host "Encoded Certificate:`n$certbase64`n"

Encoded Private Key:
LS0tLS1CRUd......==


Encoded Certificate Key:
LS0tLS1CRUe......==
```

MacOS/Linux

```
cat privateKey.key | base64

LS0tLS1CRUd......==

cat certificate.crt | base64

LS0tLS1CRUd......==
```

2. Using the output from the previous step, the base64 content can now be provided in `Custom TLS Certificate Configuration` section of the OVF property during the deployment of the VMware Event Broker Appliance.
    * Custom VMware Event Broker Appliance TLS Certificate Private Key (Base64)
    * Custom VMware Event Broker Appliance TLS Certificate Authority Certificate (Base64)

3. Power on the VMware Event Broker Appliance and ensure that the provided TLS certificate is now used instead of the auto-generated self-sign TLS certificate by opening a browser to one of the VMware Event Broker Appliance endpoints such as `/status`.
