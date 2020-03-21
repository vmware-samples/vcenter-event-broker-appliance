## Updating the TLS Certificate on VEBA
 
The default certificate for OpenFaaS (/ui) or the EventBridge (/stats) and the other web endpoints running on VEBA are self signed. This might cause browsers to show the certificate as untrusted and would require providing the `--no-tls-verify` flag when working with faas-cli. 
 
In order to update the certificates with a certificate from a trusted authority, please follow the steps outlined below
 
### Assumptions

* Access to VMware Event Broker Appliance terminal 
* Certificates from a trusted authority predownloaded onto the Appliance
    * The public/private key pair must exist before hand. The public key certificate must be .PEM encoded and match the given private key.

### Steps

Run the below commands to update the certificate on VEBA

```bash
cd /folder/certs/location
CERT_NAME=eventrouter-tls #DO NOT CHANGE THIS
KEY_FILE=<cert-key-file>.pem
CERT_FILE=<public-cert>.cer

#recreate the tls secret
kubectl --kubeconfig /root/.kube/config -n vmware delete secret ${CERT_NAME}
kubectl --kubeconfig /root/.kube/config -n vmware create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}

#reapply the config to take the new certificate
kubectl --kubeconfig /root/.kube/config apply -f /root/ingressroute-gateway.yaml
```

Watch this short video to see the steps being performed to successfully update the certs for VEBA configured for OpenFaaS - https://youtu.be/7oMCvxvL2ns
