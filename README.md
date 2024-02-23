# VMware Event Broker Appliance

[![Photon OS
4.0](https://img.shields.io/badge/Photon%20OS-4.0-orange)](https://vmware.github.io/photon/)
[![Published VMware
Fling](https://img.shields.io/badge/VMware-Fling-green)](https://vmwa.re/flings)
![Website](https://img.shields.io/website?label=vmweventbroker.io&url=https%3A%2F%2Fvmweventbroker.io%2F)

[![Twitter
Follow](https://img.shields.io/twitter/follow/lamw?style=social)](https://twitter.com/lamw)
[![Twitter
Follow](https://img.shields.io/twitter/follow/embano1?style=social)](https://twitter.com/embano1)
[![Twitter
Follow](https://img.shields.io/twitter/follow/vmw_rguske?style=social)](https://twitter.com/vmw_rguske)

<img src="logo/veba_icon_only.png" align="right" height="320px"/>

## Table of Contents

- [VMware Event Broker Appliance](#vmware-event-broker-appliance)
  - [Table of Contents](#table-of-contents)
  - [Getting Started](#getting-started)
  - [Overview](#overview)
  - [Architecture](#architecture)
  - [Getting in touch](#getting-in-touch)
  - [Contributing](#contributing)
  - [License](#license)


## Getting Started

Visit our website [vmweventbroker.io](https://vmweventbroker.io/) and explore
our [documentation](https://vmweventbroker.io/kb) to get started quickly.

## Overview

The [VMware Event Broker
Appliance](https://vmwa.re/flings)
Fling enables customers to unlock the hidden potential of events in their SDDC
to easily create [event-driven
automation](https://octo.vmware.com/vsphere-power-event-driven-automation/). The
VMware Event Broker Appliance includes support for vCenter Server and VMware
Horizon events as well as any valid `CloudEvent` through the native webhook
event provider. Easily triggering custom or prebuilt actions to deliver powerful
integrations within your datacenter across public cloud has never been more
easier before. A detailed list of use cases and possibilities with VMware Event
Broker Appliance is available [here](./USECASES.md)

With this solution, end-users, partners and independent software vendors only
have to write minimal business logic without going through a steep learning
curve understanding the vSphere or Horizon APIs. As such, we believe this
solution not only offers a better user experience in solving existing problems
for VI/Cloud Admins, SRE/Operators, Automation Engineers and 3rd Party Vendors.
More importantly, it will enable new integration use cases and workflows to grow
the VMware ecosystem and community, similar to what AWS has achieved with AWS
Lambda.

Learn more about the VMware Event Broker Appliance
[here](https://vmweventbroker.io).

Additional resources can be found [here](https://vmweventbroker.io/resources) and some
quick references are highlighted below.
 - Watch [Michael Gasch](https://github.com/embano1) and [William
   Lam](https://github.com/lamw/) present a session at VMware {Code} called [VEBA Revolutions - Unleashing the Power of Event-Driven Automation](https://youtu.be/jwgJpZM68mA?si=Vyafppqgebg1vhqd).
 - Listen to [William Lam](https://github.com/lamw/) talking about [Event-Driven Automation with Project VEBA](https://open.spotify.com/episode/3xLuJFOB4BSY749gsGn88p?si=ox8jT4mWSrS5qp5V154sJQ) in Episode #006 of the Unexplored Territory podcast.
 - Watch [Robert Guske](https://rguske.github.io/) present a session at ContainerDays 2023 called [Embark on a Transformative Odyssey - Event-Driven Automation Unveiled through Knative](https://youtu.be/J_3-ILnPbQQ?si=rutOJV_5xxl7vSmA).
 - Watch [Michael Gasch](https://github.com/embano1) and [Steven Wong](https://twitter.com/cantbewong) present a session at KubeCon EU 2022 called [Optimize Kubernetes on vSphere with Event-Driven Automation](https://youtu.be/NJYBwJemdoY?si=ploMJ2tnWgZLRqbE).

<!-- ## Users and Use Cases

Hear from the community on how they are taking advantage of the vCenter Server Appliance [here](https://vmweventbroker.io/casestudy-wip.md) -->

## Architecture

VMware Event Broker Appliance is provided as a Virtual Appliance that can be
deployed to any vSphere-based infrastructure, including an on-premises and/or
any public cloud environment running on vSphere.

The VMware Event Broker Appliance follows a highly modular approach, using
Kubernetes and containers as an abstraction layer between the base operating
system ([Photon OS](https://github.com/vmware/photon)) and the required
application services. Currently the following components are used in the
appliance:

- Tanzu Sources for Knative ([Github](https://github.com/vmware-tanzu/sources-for-knative))
  - Supported Event Stream Sources:
    - VMware vCenter ([Website](https://www.vmware.com/products/vcenter-server.html))
    - VMware Horizon ([Website](https://www.vmware.com/products/horizon.html))
    - Incoming Webhooks
  - Supported Event Stream Processors:
    - Knative ([Website](https://www.knative.dev/))
- Contour ([Github](https://github.com/projectcontour/contour))
- Kubernetes ([Github](https://github.com/kubernetes/kubernetes))
- Photon OS ([Github](https://github.com/vmware/photon))

<center><div style="height:250px;"><img src="docs/kb/img/veba-architecture.png"/></div></center>

For more details about the individual components and how they are used in the
VMware Event Broker Appliance, please see the [Architecture
page](https://vmweventbroker.io/kb/architecture).

## Getting in touch

Feel free to reach out to [Team #VEBA](https://vmweventbroker.io/#team-veba)
  - Email us at [dl-veba@vmware.com](mailto:dl-veba@vmware.com)
  - Follow for updates [@VMWEventBroker](https://twitter.com/VMWEventBroker)

## Contributing

The VMware Event Broker Appliance team welcomes contributions from the
community.

To help you get started making contributions to VMware Event Broker Appliance,
we have collected some helpful best practices in the [Contributing
guidelines](https://vmweventbroker.io/community#guidelines).

Before submitting a pull request, please make sure that your change satisfies
the requirements specified
[here](https://vmweventbroker.io/community#pull-requests)

## License

VMware Event Broker Appliance is available under the BSD-2 license. Please see
[LICENSE.txt](LICENSE.txt).
