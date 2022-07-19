---
layout: page
id: community
title: Join our Community
description: Community Resources
permalink: /community
links:
- title: Twitter
  image: /assets/img/icons/twitter.svg
  items:
  - description: "Follow us at "
    url: "https://twitter.com/VMWEventBroker"
    label: "@VMWEventBroker"
- title: Github
  image: /assets/img/icons/github.svg
  items:
  - description: "Follow us at "
    url: "https://github.com/vmware-samples/vcenter-event-broker-appliance"
    label: vcenter-event-broker-appliance
- title: Slack
  image: /assets/img/icons/slack.svg
  items:
  - description: "Join us at"
    url: "https://vmwarecode.slack.com/archives/CQLT9B5AA"
    label: "&#35;vcenter-event-broker-appliance"
- title: Email
  image: /assets/img/icons/email.svg
  items: 
  - description: "Email us at "
    url: "mailto:dl-veba@vmwarem.com"
    label: dl-veba@vmware.com
---

# Get in touch

<div class="container pb-3 pt-0">
  <div class="row justify-content-md-center">
    {% for link in page.links %}
    <div class="col-md-3 community-item text-center pt-2">
      <div class="icon mt-2">
        <img src="{{ link.image | relative_url }}" style="height: 45px;" alt="{{ link.title}}">
      </div>
      <h2 class="mt-2">{{link.title}}</h2>
      {% for item in link.items %}
      <div class="link-description">
        <p class="mb-0 pb-0">{{ item.description }}</p>
        <span class="mt-0 pt-0"><a href="{{ item.url }}" target="_blank">{{ item.label }}</a></span>
      </div>
      {% endfor %}
    </div>
    {% endfor %}
  </div>
</div>

The VMware Event Broker Appliance team welcomes contributions from the community and this page presents the guidelines for contributing to VMware Event Broker Appliance. 

## Contributing

Following the guidelines helps to make the contribution process easy,
collaborative, and productive.

Before you start working with the VMware Event Broker Appliance, please read our
[Developer Certificate of Origin](https://cla.vmware.com/dco){:target="_blank"}.
All contributions to this repository must be signed as described on that page.
Your signature certifies that you wrote the patch or have the right to pass it
on as an open-source patch.

This is a rough outline of what a contributor's workflow looks like:

- Create an issue describing the feature/fix.
- Create a topic branch from where you want to base your work.
- Make commits of logical units.
- Make sure your commit messages are in the proper format (see below).
- Push your changes to a topic branch in your fork of the repository.
- Submit a pull request to `vmware-samples/vcenter-event-broker-appliance`.

See [below](#format-of-the-commit-message) for details on commit best practices
and **supported prefixes**, e.g. `fix: <message>`.

> **Note:** If you are new to Git(hub) check out [Git rebase, squash...oh
> my!](https://www.mgasch.com/2021/05/git-basics/) for more details on how to
> successfully contribute to an open source project.

### Submitting Bug Reports and Feature Requests

Please submit bug reports and feature requests by using our GitHub
[Issues](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues){:target="_blank"}
page.

Before you submit a bug report about the code in the repository, please check
the Issues page to see whether someone has already reported the problem. In the
bug report, be as specific as possible about the error and the conditions under
which it occurred. On what version and build did it occur? What are the steps to
reproduce the bug?

Feature requests should fall within the scope of the project.

### Pull Requests

Before submitting a pull request, please make sure that your change satisfies
the following requirements:

- The change is signed as described by the [Developer Certificate of
  Origin](https://cla.vmware.com/dco){:target="_blank"} doc.
- The change is clearly documented and follows Git commit best practices (see
  below)

### Format of the Commit Message

We follow the conventions described in [How to Write a Git Commit
Message](http://chris.beams.io/posts/git-commit/).

Be sure to include any related GitHub issue references in the commit message,
e.g. `Closes: #<number>`.

The `CHANGELOG.md` and release page use **commit message prefixes** for grouping
and highlighting. A commit message that starts with `[prefix:] ` will place this
commit under the respective section in the `CHANGELOG`. 

The following example creates a commit referencing the `issue: 1234` and puts
the commit message in the `Feature` `CHANGELOG` section:

```bash
git commit -s -m "feat: Add Slack Example" -m "Closes: #1234"
```

Currently the following prefixes are used:

- `fix:` - Use for bug fixes
- `feat:` - A new feature
- `chore:` - Use for repository related activities
- `docs:` - Use for changes to the documentation

If your contribution falls into multiple categories, e.g. `feat` and `fix` it is
recommended to break up your commits using distinct prefixes.

### Contributions to the Appliance

- See the Build Appliance document [here](/kb/contribute-appliance)
- See the Build Event Router document [here](/kb/contribute-eventrouter)
- Requestor must verify that the VMware Event Broker Appliance can be built and deployed.

### Contributions to the Functions

- See the Build Functions document [here](/kb/contribute-functions)
- PR should contain information on how the function was tested (environment, version etc)
- PR should contain a titled readme and the title is listed in the [Functions](/examples) page

### Contributions to the Website

- See the Build Website document [here](/kb/contribute-functions)
- Requestor must verify that the website change was built and tested locally

Get started quickly with your contributions with our [getting started](/kb/contribute-start) guide

## Join the movement

<div id="contributors-veba" class="section section-background-{{ page.backgrounds.team }} p-3">
    {% include contributors.html %}
</div>
