---
layout: function
id: functions
type: knative
title: Prebuilt Functions
description: Community-sourced and validated prebuilt functions for OpenFaaS with VEBA.
permalink: /examples-knative
images:
  powercli: /assets/img/languages/powercli.png
  python: /assets/img/languages/python.png
  go: /assets/img/languages/go.png
  powershell: /assets/img/languages/powershell.png
examples:
  - title: Echo Cloud Event for Knative
    usecases: 
    - item: other
    id: kn-echo-function
    description: Function helps users understand the structure and data of a given vCenter Event using the Knative event processor which will be useful when creating brand new Functions.
    links: 
    - language: python
      url: "/tree/master/examples/knative/python/kn-echo"

---

A complete and updated list of ready to use functions curated by the VMware Event Broker community is listed below. 

# Get started with our prebuilt functions

These functions are prebuilt, available in ready to deploy container and `stack.yml` files for you to deploy as is. Should you need to modify the functions to fit your needs, the `README.md` files provided within each function folder will provide all the information you need to customize, build and deploy the function on your VMware Event Broker appliance. 

> **Note:** These functions are provided and tested to be used with the VMware Event Broker Appliance deployed with [Knative](/kb/install-knative) as the event stream processor. 


 <div class="examples wd-100">
    <h2>Functions</h2>
    {% for ex in page.examples %}
    <div id="{{ ex.id }}" class="row title">
      <h3>{{ex.title}}</h3>
      <div class="language">
      {% for link in ex.links %}
        <a href="{{ link.url | prepend: site.gh_repo}}" target="_blank" class="col-md-2 pr-0">
            <img src="{{ '/assets/img/languages/' | append: link.language | append: '.png' | relative_url}}" alt="{{ link.language }}">
            <span class="m-0">{{ link.language }}</span>
        </a>
      {% endfor %}
      </div>
    </div>
    {{ ex.description | markdownify }}
    <div class="usecases">
      {% for usecase in ex.usecases %}
      <span class="{{usecase.item}}">{{usecase.item}}</span>
      {% endfor %}
    </div>
    
    {% endfor %}
</div>

## Contributions

These functions serve as an easy way to use the appliance and as an inspiration for how to write functions in different languages. If you have an idea for a function and are looking to write your own, start with our documentation [here](/kb/contribute-functions). 

Check our [contributing guidelines](\community#contributing) and join [Team #VEBA](/#team-veba) by submitting a pull request for your function to be showcased on this list. 