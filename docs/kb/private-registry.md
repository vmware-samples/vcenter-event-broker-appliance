---
layout: docs
toc_id: private-container-registry
title: VMware Event Broker Appliance - Private Registry
description: Private Registry
permalink: /kb/private-registry
cta:
 description: Using private container registry with the VMware Event Broker Appliance.
---

## Using private container registry with VEBA

By default, the VMware Event Broker Appliance can integrate with any Open Container Initiative (OCI) compliant container registry for hosting and deploying container images that uses a TLS certificate from a trusted authority such as [Docker Hub](https://hub.docker.com/) or [Amazon Elastic Container Registry (ECR)](https://aws.amazon.com/ecr/) as an example.

For organizations that require the use of a private container registry and  uses a self-signed TLS certificate, an additional post-deployment configuration is required within the VMware Event Broker Appliance. Please follow the steps outlined below.

### Assumptions

* VMware Event Broker Appliance v0.7.2 or later
* Root CA Certificates from a trusted authority has been pre-downloaded onto your local desktop
* Un-deploy any functions that has been attempted using private registry prior to instructions below

> **Note:** For those using the Harbor registry, the root CA certificate is located in : **/etc/docker/certs.d/[FQDN]/ca.crt**

### Steps

In this example, the root CA certificate key file is named `ca.crt` and is located in `/root`

1. Copy the root CA certificate from your private registry to VMware Event Broker Appliance Appliance. If SSH has not been enabled, go ahead and start it up by logging into the VM Console and running the following command:

```console
systemctl start sshd
```

1. SSH to the VMware Event Broker Appliance and make a backup of the original containerd configuration file

```console
cp /etc/containerd/config.toml /etc/containerd/config.toml.bak
```

1. Edit the '/etc/containerd/config.toml' file using VI and locate the following section `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]` within the configuration file.

Append the following two lines below this section and replace the **REPLACE_ME_FQDN** value with FQDN of the private registry and **REPLACE_ME_PATH_TO_ROOT_CA_CERT** value the full path to the root CA certificate located on the VMware Event Broker Appliance

```yaml
[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
   [plugins."io.containerd.grpc.v1.cri".registry.configs."REPLACE_ME_FQDN".tls]
   ca_file = "REPLACE_ME_PATH_TO_ROOT_CA_CERT"
```

1. Restart the containerd service for the change to go into effect/.

```console
systemctl restart containderd
```

1. Verify containerd is successfully running before proceeding to the next step

```console
systemctl status containerd

● containerd.service - containerd container runtime
     Loaded: loaded (/usr/lib/systemd/system/containerd.service; enabled; vendor preset: disabled)
     Active: active (running) since Mon 2022-03-28 19:35:06 UTC; 1 day 3h ago
       Docs: https://containerd.io
    Process: 30072 ExecStartPre=/sbin/modprobe overlay (code=exited, status=0/SUCCESS)
   Main PID: 30073 (containerd)
      Tasks: 545
     Memory: 1.2G
     CGroup: /system.slice/containerd.service
```

1. Create a kubernetes secret in the `knative-serving` namespace that points to full path of the root CA certificate of private registry which should reside within the VMware Event Broker Appliance

```console
kubectl -n knative-serving create secret generic customca --from-file=ca.crt=/root/ca.crt
```

1. Retrieve the current Knative Serving controller deployment and save it to a file named `knative-serving-controller.yaml`

```console
kubectl -n knative-serving get deploy/controller -o yaml > knative-serving-controller.yaml
```

1. Create the following YTT overlay which will be used to patch the Knative Serving controller to reference the root CA certificate from the private registry

```console
cat > overlay.yaml <<EOF
#@ load("@ytt:overlay", "overlay")

#@overlay/match by=overlay.subset({"kind":"Deployment", "metadata": {"name": "controller", "namespace": "knative-serving"}})
---
spec:
  template:
    spec:
      containers:
      #@overlay/match by=overlay.subset({"name": "controller"})
      -
         env:
             #@overlay/append
             - name: SSL_CERT_DIR
               value: /etc/customca
         #@overlay/match missing_ok=True
         volumeMounts:
             - name: customca
               mountPath: /etc/customca
      #@overlay/match missing_ok=True
      volumes:
        - name: customca
          secret:
            secretName: customca
EOF
```

1. Apply the YTT transformation to create the new Knative Serving controller YAML file named `new-knative-serving-controller.yaml`

```console
ytt -f overlay.yaml -f knative-serving-controller.yaml > new-knative-serving-controller.yaml
```

1. Apply the new Knative Serving controller configuration

```console
kubectl apply -f new-knative-serving-controller.yaml
```

It can take a couple of minutes for the previous Knative Serving controller to terminate and spawn the new configuration. You can monitor the progress using the following commmand and ensure the `READY` state shows 1/1

```console
kubectl -n knative-serving get deployment/controller -w

NAME         READY   UP-TO-DATE   AVAILABLE   AGE
controller   1/1     1            1           29h
```

> **Note:** If for some reason the deployment is not re-deploying, you can run `kubectl -n knative-serving delete deployment/controller` and then perform the `apply` operation with the new Knative Serving YAML.