# About the Example Functions

Example Functions serve as an easy way to use the appliance and as an
inspiration for how to write functions in different languages.

> **Note:** These functions are provided and tested to be used with the VMware
> Event Broker Appliance deployed with
> [Knative](https://vmweventbroker.io/kb/architecture) as the event stream
> processor. 

VMware Event Broker Appliance with Knative allows you to write functions in any
language. These functions are organized by the language that they are written on
as shown above

When you are making a contribution, the [master list of
functions](https://vmweventbroker.io/examples) should be updated by changing the
YAML within
[docs/site/examples-knative.md](./../docs/site/examples-knative.md).

You must provide all required fields and the information must be in this
particular format. 

```yaml
  - title: POST to any REST API     #short catchy title  
    usecases: 
    - item: automation
    - item: #choose b/w analytics, audit, automation, integration, notification, remediation and other
    id: post-rest-api                #id for hyperlink anchoring
    description: #description of the function purpose and what it does
    links: 
    - language: python              #use python for python3 as well
      url: "/tree/master/examples/knative/python/invoke-rest-api" #relative path to the function
```