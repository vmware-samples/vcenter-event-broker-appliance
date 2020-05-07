---
layout: docs
toc_id: advanced-deploy-k8s
title: VMware Event Broker Appliance - Event Router Standalone
description: Standalone Deployment of Event Router
cta:
 title: What's next?
 description: Extend your vCenter seamlessly with our pre-built functions
 actions:
  - text: See our complete list of prebuilt functions - [here](/examples)
  - text: Deploy a Function - [here](use-functions).
---

# Standalone Deployment of Event Router
VMware Event Router can be deployed and run as standalone binary (see [below](#build-from-source)). However, it is designed to be run in a Kubernetes cluster for increased availability and ease of scaling out. The following steps describe the deployment of the VMware Event Router in **a Kubernetes cluster** for an existing OpenFaaS ("faas-netes") environment, respectively AWS EventBridge.

> **Note:** Docker images are available [here](https://hub.docker.com/r/vmware/veba-event-router){:target="_blank"}.

Create a namespace where the VMware Event Router will be deployed to:

```bash
kubectl create namespace vmware
```

Use one of the configuration files provided [here](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/development/vmware-event-router/deploy){:target="_blank"} to configure the router for **one** VMware vCenter Server event `stream` and **one** OpenFaaS **or** AWS EventBridge event stream `processor`. Change the values to match your environment. The following example will use the OpenFaaS config sample.

> **Note:** Make sure your environment is up and running, i.e. Kubernetes and OpenFaaS (incl. a function for testing) up and running or AWS EventBridge correctly configured (IAM Role, event bus and pattern rule).

After you made your changes to the configuration file, save it as `"event-router-config.json` in your current Git working directory. 

> **Note:** If you have changed the port of the metrics server in the configuration file (default: 8080) make sure to also change that value in the YAML manifest (under the Kubernetes service entry).

Now, from your current Git working directory create a Kubernetes [secret](https://kubernetes.io/docs/concepts/configuration/secret/){:target="_blank"} from the configuration file:

```bash
kubectl -n vmware create secret generic event-router-config --from-file=event-router-config.json
```

> **Note:** You might want to delete the (local) configuration file to not leave behind sensitive information on your local machine.

Now we can deploy the VMware Event Router:

```bash
kubectl -n vmware create -f deploy/event-router-k8s.yaml
```

Check the logs of the VMware Event Router to validate it started correctly:

```bash
kubectl -n vmware logs deploy/vmware-event-router -f
```

If you run into issues, the logs should give you a hint, e.g.:

- configuration file not found -> file naming issue
- connection to vCenter/OpenFaaS cannot be established -> check values in the configuration file
- deployment/pod will not even come up -> check for resource issues, docker pull issues and other potential causes using the standard kubectl troubleshooting ways

To delete the deployment and secret simply delete the namespace we created earlier:

```bash
kubectl delete namespace vmware
```