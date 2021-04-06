---
layout: docs
toc_id: advanced-certificates
title: VMware Event Broker Appliance - Certificates
description: Updating Certificates
permalink: /kb/advanced-certificates
cta:
 description: With your certificates updated, you can now skip using the `--tls-no-verify` flag while working with faas-cli.
---

## Updating the TLS Certificate on VEBA
 
The default certificate for OpenFaaS (/ui), Stats (/stats), Status (/status), Logs (/bootstrap) and the other web endpoints running on VEBA are self signed. This might cause browsers to show the certificate as untrusted and would require providing the `--no-tls-verify` flag when working with faas-cli. 
 
In order to update the certificates with a certificate from a trusted authority, please follow the steps outlined below
 
### Assumptions

* Access to VMware Event Broker Appliance terminal 
* Certificates from a trusted authority pre-downloaded onto the Appliance
    * The public/private key pair must exist before hand. The public key certificate must be .PEM encoded and match the given private key.

### Steps

Run the below commands to update the certificate on VEBA

```bash
cd /folder/certs/location
CERT_NAME=eventrouter-tls #DO NOT CHANGE THIS
KEY_FILE=<cert-key-file>.pem
CERT_FILE=<public-cert>.cer

#recreate the tls secret
kubectl -n vmware-system delete secret ${CERT_NAME}
kubectl -n vmware-system create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}

#reapply the config to take the new certificate
kubectl apply -f /root/config/ingressroute-gateway.yaml
```

If you are using the Embedded Knative Broker, you will also need to reference the newly generated certificate as it is also used as part of the Knative Contour integration.

```bash
cd /folder/certs/location
KNATIVE_CERT_NAME=eventrouter-tls #DO NOT CHANGE THIS
KEY_FILE=<cert-key-file>.pem
CERT_FILE=<public-cert>.cer

#recreate the tls secret
kubectl -n contour-external delete secret ${KNATIVE_CERT_NAME}
kubectl -n contour-external create secret tls default-cert --key ${KEY_FILE} --cert ${CERT_FILE}

#reapply the config to take the new certificate
kubectl apply -f /root/config/ingressroute-gateway.yaml
```

Watch this short video to see the steps being performed to successfully update the certs for VEBA configured for OpenFaaS - [here](https://youtu.be/7oMCvxvL2ns){:target="_blank"}
