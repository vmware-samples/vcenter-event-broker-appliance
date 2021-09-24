---
layout: docs
toc_id: contribute-start
title: VMware Event Broker Appliance - Getting Started
description: Getting Started
permalink: /kb/contribute-start
cta:
 title: Join our community 
 description: Earn a place amongst our top contributors [here](/community#contributors-veba)
 actions:
    - text: Learn how you can contribute to our Appliance build process [here](contribute-appliance)
    - text: Learn how you can contribute to our VMware Event Router [here](contribute-eventrouter)
    - text: Learn how you can contribute to our Pre-built Functions [here](contribute-functions)
    - text: Learn how you can contribute to our Website [here](contribute-site).
---

# Contributing

The VMware Event Broker Appliance team welcomes contributions from the community.

Before you start working with the VMware Event Broker Appliance, please read our [Developer Certificate of Origin](https://cla.vmware.com/dco){:target="_blank"}. All contributions to this repository must be signed as described on that page. Your signature certifies that you wrote the patch or have the right to pass it on as an open-source patch.

# Step-by-step help

If you're just starting out with containers and source control and are looking for additional guidance, browse to this [blog series](http://www.patrickkremer.com/veba/){:target="_blank"} on VEBA. The blog goes step-by-step with screenshots and assumes zero experience with any of the required tooling.

# Preqrequisites

You must create a [Github](https://github.com/join){:target="_blank"} account. You need to verify your email with Github in order to contribute to the VEBA repository.

The following tools are required:
 * [Git](https://git-scm.com/downloads){:target="_blank"}
 * [Docker](https://docs.docker.com/){:target="_blank"}
 * [Knative CLI](https://knative.dev/docs/install/install-kn/){:target="_blank"}

# Quickstart for Contributing

## Download the VEBA source code
```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
```

## Configure git to sign code with your verified name and email
```bash
git config --global user.name "Your Name"
git config --global user.email "youremail@domain.com"
```

## Contribute documentation changes

Make the necessary changes and save your files. 
```bash
git diff
```

This is sample output from git. It will show you files that have changed as well as all display all changes.
```bash
user@wrkst01 MINGW64 ~/Documents/git/vcenter-event-broker-appliance(master)
$ git diff
diff --git a/docs/kb/contribute-start.md b/docs/kb/contribute-start.md
index 4245046..f86f09f 100644
--- a/docs/kb/contribute-start.md
+++ b/docs/kb/contribute-start.md
@@ -6,6 +6,32 @@ description: Getting Started
 permalink: /kb/contribute-start
 ---

+# Preqrequisites
+
+Three tools are required in order to contribute functions - 
(output truncated)
```

Commit the code and push your commit. -a commits all changed files, -s signs your commit, and -m is a commit message - a short description of your change.


```bash
git commit -a -s -m "Add prereq and git diff output to contribution page."
git push
```

You can then submit a pull request (PR) to the VEBA maintainers - a step-by-step guide with screenshots is available [here](http://www.patrickkremer.com/2019/12/vcenter-event-broker-appliance-part-v-contributing-to-the-veba-project/){:target="_blank"}

# Changing or contributing new functions

The git commands are the same, but in order to change code, you must reference your own Docker image. The example YAML below comes from the Knative [kn-pcli-tag]](https://github.com/vmware-samples/vcenter-event-broker-appliance/tree/master/examples/knative/powercli/kn-pcli-tag){:target="_blank"} sample function.

> Note: The `image:` references the `vmware` harbor registry account. You must change this to your own docker account.

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: kn-pcli-tag
  labels:
    app: veba-ui
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "1"
        autoscaling.knative.dev/minScale: "1"
spec:
  template:
    metadata:
    spec:
      containers:
        - image: projects.registry.vmware.com/veba/kn-pcli-tag:1.1
          envFrom:
            - secretRef:
                name: tag-secret
          env:
            - name: FUNCTION_DEBUG
              value: "false"
---
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: veba-pcli-tag-trigger
  labels:
    app: veba-ui
spec:
  broker: default
  filter:
    attributes:
      type: com.vmware.event.router/event
      subject: DrsVmPoweredOnEvent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: kn-pcli-tag
```

Once you've written or changed function code, you can deploy it to your VEBA appliance for testing.
```bash
export TAG=<version>
docker build -t <docker-username>/kn-pcli-tag:${TAG} .
docker push <docker-username>/kn-pcli-tag:${TAG}
kubectl -n vmware-functions apply -f function.yaml
```
If everything works as expected, you can then commit your code and file a pull request for inclusion into the project.

## Submitting Bug Reports and Feature Requests

Please submit bug reports and feature requests by using our GitHub [Issues](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues){:target="_blank"} page.

Before you submit a bug report about the code in the repository, please check the Issues page to see whether someone has already reported the problem. In the bug report, be as specific as possible about the error and the conditions under which it occurred. On what version and build did it occur? What are the steps to reproduce the bug?

Feature requests should fall within the scope of the project.

## Pull Requests (PRs)

Before submitting a pull request, please make sure that your change satisfies the following requirements:
- VMware Event Broker Appliance can be built and deployed. See the getting started build guide [here](contribute-eventrouter.md).
- The change is signed as described by the [Developer Certificate of Origin](https://cla.vmware.com/dco){:target="_blank"} doc.
- The change is clearly documented and follows Git commit [best practices](https://chris.beams.io/posts/git-commit/){:target="_blank"}
- Contributions to the [examples](/examples) contain a titled readme.

### Commits

Please follow the guidance provided
[here](https://chris.beams.io/posts/git-commit/) how to write good commit
messages.

Please also include the issue number(s) this PR fixes. The
[`CHANGELOG.md`](./../../CHANGELOG.md) generated by this project uses this
information and specific commit title prefixes (see below) to render the
`CHANGELOG` with details and highlights.

Currently, the following commit title prefixes are recognized:

- `fix:` will be highlighted as  `🐞 Fix`
- `feat:` will be highlighted as `💫 Feature`
- `chore:` will be highlighted as `🧹 Chore`
- `docs:` will be highlighted as `📃 Documentation`

For example, to fix a bug in the VMware Event Router:

```bash
# make changes, then
git add <files>
git commit -s -m "fix: Race in VMware Event Router" -m "Closes: #issue-number"
```

### Breaking Changes

If this PR introduces a **breaking** change, please indicate this in the
corresponding commit:

```bash
cat << EOF | git commit -s -F -
Remove vCenter Simulator

This commit removes support for the vCenter Simulator event provider
Closes: #1234

BREAKING: Remove deprecated vCenter Simulator
EOF
```