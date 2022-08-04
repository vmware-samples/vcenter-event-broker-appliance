---
layout: timeline
title: Evolution of VEBA
description: The Evolution of VMware Event Broker Appliance
permalink: /evolution
limit: 3
entry:

- title: VEBA [v0.7.4](https://github.com/vmware-samples/vcenter-event-broker-appliance/releases/tag/v0.7.4)
  type: feature
  date: Aug 2022
  id: vzerosevenfour
  detail:
    subtitle: Features
    text:
    - Fixed VEBA vSphere UI for subscribing to multiple events for single function

- title: VEBA [v0.7.3](https://github.com/vmware-samples/vcenter-event-broker-appliance/releases/tag/v0.7.3)
  type: feature
  date: Jul 2022
  id: vzeroseventhree
  detail:
    subtitle: Features
    text:
    - New VMware Harbor, Zapier, NSX Tag Sync & Unapproved Portgroup Usage Functions
    - Ported Datastore Usage Email, vSphere HA Restart & Host Maint. Alarm functions to Knative
    - Enhanced pattern matching for EventBridge processor
    - Various Documentation Updates & Bug Fixes

- title: VEBA [v0.7.2](https://williamlam.com/2022/03/vmware-event-broker-appliance-veba-v0-7-2.html)
  type: feature
  date: Mar 2022
  id: vzeroseventwo
  detail:
    subtitle: Features
    text:
    - New Knative PowerCLI template (quickly get started on building new functions)
    - PowerShell Slack Function enhancement to be event agnostic + customizable message
    - New PowerCLI example function to enforce VDS & DVPortgroup configs
    - Syslog now captures all logs via Fluentbit
    - RabbitMQ (triggers) now supports function scaling to scale out when there's a burst of events
    - Added Let's Encrypt documentation
    - Various backend updates (see this [blog post](https://williamlam.com/2022/03/vmware-event-broker-appliance-veba-v0-7-2.html) for more details)

- title: VEBA [v0.7.1](https://williamlam.com/2021/12/vmware-event-broker-appliance-veba-v0-7-1.html)
  type: feature
  date: Dec 2021
  id: vzerosevenone
  detail:
    subtitle: Features
    text:
    - Fix special character handling for VEBA vSphere UI plugin
    - Fix imagePullPolicy for knative-contour in air-gap deployment
    - Improved website documentation
    - More Knative Function Examples

- title: VEBA [v0.7.0](https://williamlam.com/2021/10/whats-new-in-vmware-event-broker-appliance-veba-v0-7.html)
  type: feature
  date: Oct 2021
  id: vzeroseven
  detail:
    subtitle: Features
    text:
    - New VMware Horizon Event Provider
    - New Generic Webhook Event Provider
    - Embedded cAdvisor for Monitoring
    - Support for External Syslog
    - Lots of new Knative PowerShell & PowerCLI Function Examples
    - Deprecation of OpenFaaS & EventBridge Event Processors

- title: Listed in IT Magazine
  type: milestone
  date: Jul 2021
  id: itmagazine
  description: VEBA listed in IT Magazine - [details](https://twitter.com/lamw/status/1417984664015314947)

- title: VEBA [v0.6.1](https://williamlam.com/2021/06/vmware-event-broker-appliance-veba-v0-6-1.html)
  type: feature
  date: Jun 2021
  id: vzerosixone
  detail:
    subtitle: Features
    text:
    - Knative PowerShell / PowerCLI Base Container Image Templates
    - New Knative PowerCLI, Python and Go Function Examples
    - Custom TLS Certificate support
    - Documentation for adding Trusted Root Certificate to VEBA
    - Helm support for Knative
    - Enhanced Release Notes

- title: VEBA [v0.6.0](https://williamlam.com/2021/04/vmware-event-broker-appliance-veba-v0-6-is-now-available.html)
  type: feature
  date: Apr 2021
  id: vzerosix
  detail:
    subtitle: Features
    text:
    - Embedded Knative (new default Event Processor)​
    - vSphere UI integration (H5 Client Plugin)​
    - 1st class PowerShell Support in Knative​
    - Easy vSphere CloudEvents Viewer (Sockeye)

- title: Listed in CloudEvents.io
  type: milestone
  date: Oct 2020
  id: cloudevents
  description: VEBA listed on CloudEventsIO site! - [details](https://twitter.com/lamw/status/1362572308754305024)

- title: VEBA [v0.5.0](https://williamlam.com/2020/12/vcenter-event-broker-appliance-veba-v0-5-0.html)
  type: feature
  date: Dec 2020
  id: vzerofive
  detail:
    subtitle: Features
    text:
    - Helm chart for simplified deployment experience
    - Introduced support for external Knative environment​ as the Event Processor
    - Implement at-least-once delivery semantics and improve resiliency across all processors
    - Contributions back to Knative and govmomi for better vSphere eventing integration

- title: Knative 2020 Annual Report
  type: milestone
  date: Oct 2020
  id: knativeannualreport
  description: Listed on Knative in the Wild - Page 07 on [Knative 2020 Annual Report](https://knative.dev/community/contributing//annual_reports/Knative%202020%20Annual%20Report.pdf)

- title: VEBA [v0.4.0](https://williamlam.com/2020/05/new-veba-release-new-website-and-new-mascot.html)
  type: feature
  date: May 2020
  id: vzerofour
  detail:
    subtitle: Features
    text:
    - Introduced [Otto](https://www.virtuallyghetto.com/2020/05/new-veba-release-new-website-and-new-mascot.html)
    - Launched [vmweventbroker.io](https://vmweventbroker.io)
    - Easier deployment of router in a Kubernetes environment (non-appliance mode) → towards core vSphere integration (e.g. WCP)
    - Introduced DCUI with several [EasterEggs](https://www.virtuallyghetto.com/2020/05/new-veba-release-new-website-and-new-mascot.html)

- title: VEBA [v0.3.0](https://williamlam.com/2020/03/big-updates-to-the-vcenter-event-broker-appliance-veba-fling.html)
  type: feature
  date: Mar 2020
  id: vzerothree
  detail:
    subtitle: Features
    text:
    - Introduced VMware Event Router
    - Added support for AWS EventBridge as an Event Processor
    - Conformance with CloudEvents
    - Support all vCenter events (incl. full payload)
    - Improved contribution guidelines due to high interest in contributing to VEBA from customers/partners
    - Add more enterprise features, e.g. logging and metrics

- title: VEBA v0.2.0
  type: feature
  date: Jan 2020
  id: vzerotwo
  detail:
    subtitle: Features
    text:
    - Added more enterprise details, e.g. proxy support, offline deployments
    - Documentation improvements and more samples from the community

- title: Launching VEBA [v0.1.0](https://williamlam.com/2019/11/vcenter-event-broker-appliance-updates-vmworld-fling-community-open-source.html)
  type: feature
  date: Nov 2019
  id: vzeroone
  description: VEBA Fling v0.1.0 released live at VMworld Europe - 'Do it, do it, do it\' chants when @embano1 and @lamw ask if they should release the vCenter Event Broker Appliance fling (powered by @openfaas) live during their [session!](https://twitter.com/bbrundert/status/1192366254570508288)
  detail:
    subtitle: Features
    text:
    - Introduced VEBA v0.1.0 with OpenFaaS as the default Event Processor​ and vCenter as the Event Provider

- title: Idea
  type: milestone
  date: APR 2019
  id: vzero
  description: Michael and William meet to discuss use cases, mainly around event-driven automation/notification and compliance (changes to VMs, DRS, etc.)

---

{% for entry in page.entry %}
<div class="timeline-item">
    <div class="timeline-img"></div>
    <div id="{{ entry.id }}" class="timeline-content {% if entry.type == 'milestone' %}milestone{% else if entry.type == 'feature' %}feature{% endif %}">
        <h2 class="post-title"> 
          {{entry.title | markdownify | remove: '<p>' | remove: '</p>'}}
        </h2>
        <p>{{entry.description | markdownify}}</p>
        {% if entry.detail %}
            <h4 class="post-subtitle"> 
                {{entry.detail.subtitle}}
            </h4>
            <ul>
            {% for val in entry.detail.text %}
                <li>
                    <p>{{val | markdownify}}</p>
                </li>
            {% endfor %}
            </ul>
        {% endif %}
        <div class="date">{{entry.date}}</div>
    </div>
</div>
{% endfor %}
