# About the Example Functions

Example Functions serve as an easy way to use the appliance and as an inspiration for how to write functions in different languages.

> **Note:** These functions are provided and tested to be used with the VMware Event Broker Appliance deployed with [OpenFaaS](https://vmweventbroker.io/kb/architecture) as the event stream processor. 

VMware Event Broker Appliance with OpenFaaS allows you to write functions in any language. These functions are organized by the language that they are written on as shown above

When you are making a contribution, the [master list of functions](https://vmweventbroker.io/examples) should be updated by updating the yaml within [docs/site/examples.md](https://github.com/vmware-samples/vcenter-event-broker-appliance/blob/master/docs/site/examples.md). You must provide all required fields and the information must be in this particular format. 

```yaml
  - title: POST to any REST API     #short catchy title  
    usecases: 
    - item: automation
    - item: #choose b/w analytics, audit, automation, integration, notification, remediation and other
    id: post-rest-api                #id for hyperlink anchoring
    description: #description of the function purpose and what it does
    links: 
    - language: python              #use python for python3 as well
      url: "/tree/master/examples/python/invoke-rest-api" #relative path to the function
```