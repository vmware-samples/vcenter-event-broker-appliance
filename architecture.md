# Architecture

The vCenter Event Broker Appliance follows a highly modular approach, using Kubernetes and containers as an abstraction layer between the base operating system ([Photon OS](https://github.com/vmware/photon)) and the required application services. Currently the following components are used in the appliance:

- VMware Event Router ([Github](https://github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router))
  - Supported Event Stream Sources:
    - VMware vCenter ([Website](https://www.vmware.com/products/vcenter-server.html))
  - Supported Event Stream Processors: 
    - OpenFaaS ([Website](https://www.openfaas.com/))
    - AWS EventBridge ([Website](https://aws.amazon.com/eventbridge/))
- Contour ([Github](https://github.com/projectcontour/contour))
- Kubernetes ([Github](https://github.com/kubernetes/kubernetes))
- Photon OS ([Github](https://github.com/vmware/photon))

<div style="height:250px;width:250px"><img src="veba-appliance-diagram.png" /></div>

In the following sections we describe the individual components.

> **Note:** Encompassing details are also provided in the [FAQ](FAQ.md).

### VMware Event Router

The `vmware-event-router` implements the core functionality of VEBA, that is connecting to event streams ("sources") and processing the events with a configurable event processor such as OpenFaaS or AWS EventBridge.

### OpenFaaS

OpenFaaS&reg; makes it easy for developers to deploy event-driven functions and microservices to Kubernetes without repetitive, boiler-plate coding. Package your code or an existing binary in a Docker image to get a highly scalable endpoint with auto-scaling and metrics.

In the vCenter Event Broker Appliance OpenFaaS powers the appliance-integrated Function-as-a-Service framework to **trigger (custom) functions based on vSphere events**. The OpenFaaS user interface provides an easy to use dashboard to deploy and monitor functions. Functions can be authored and also deployed via an easy to use [CLI](https://github.com/openfaas/faas-cli).

### AWS EventBridge

Amazon EventBridge is a serverless event bus that makes it easy to connect applications together using data from your own applications, integrated Software-as-a-Service (SaaS) applications, and AWS services.

<!-- TODO: decided whether this is configurable via OVF natively or post-deployment option -->
The vCenter Event Broker Appliance offers native integration for **event forwarding to AWS EventBridge**. The only requirement is creating a dedicated IAM user (access_key) and associated EventBridge rule on the default (or custom) event bus in the AWS management console to be used by this appliance. Only events matching the specified event pattern (EventBridge rule) will be forwarded to limit outgoing network traffic and costs.

### Contour

Contour is an Ingress controller for Kubernetes that works by deploying the Envoy proxy as a reverse proxy and load balancer. Contour supports dynamic configuration updates out of the box while maintaining a lightweight profile.

In the vCenter Event Broker Appliance Contour provides **TLS termination for the various HTTP(S) endpoints** served.

### Kubernetes

Kubernetes is an open source system for managing containerized applications across multiple hosts. It provides basic mechanisms for deployment, maintenance, and scaling of applications.

For application and appliance developers Kubernetes provides **powerful platform capabilities**, such as application (container) self-healing, secrets and configuration management, resource management, extensibility, etc. Kubernetes lays the foundation for future improvements of the vCenter Event Broker Appliance with regards to **high availability (n+1) and scalability (horizontal scale out)**.

### Photon OS

Photon OS&trade; is an open source Linux container host optimized for cloud-native applications, cloud platforms, and VMware infrastructure. Photon OS provides a **secure run-time environment for efficiently running containers** and out of the box support for Kubernetes.

Photon OS is the foundation for many appliances built for the vSphere platform and its ecosystem and thus the first choice for building the vCenter Event Broker Appliance.
