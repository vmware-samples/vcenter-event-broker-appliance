
<a name="v0.6.0"></a>
## [Release v0.6.0](https://github.com/vmware-samples/vcenter-event-broker-appliance/compare/v0.5.0...v0.6.0)

> Release Date: 2021-04-20

### ⚠️ BREAKING

### 📖 Commits

- [2088553]	Bump version to v0.6.0 for release
- [831193f]	Clarifying Knative OVF description + Docs
- [8246c90]	Adding missing vmware-functions NS to secret example
- [27b6553]	Re-enable Registry Digest Check
- [77f830e]	Fix typo for Docker push command
- [f53f20c]	Fixed VEBA UI image + Documentation Updates (#337)
- [7701edb]	Add `/events` API endpoint (#335)
- [c710557]	Adding code to include external contributors
- [bc15599]	Bump version to v0.6.0 for release
- [df1b12a]	Integrate Sockeye into VEBA
- [1abc331]	Disable digest checking for Harbor Registry
- [fa226c5]	Add Knative Documentation + Reorganize example folders
- [d9def65]	Use mod=vendor (#329)
- [ddba4b6]	Integrate VEBA UI
- [2e0f556]	Add Knative to architecture (#323)
- [5df7dab]	Update stale workflow
- [e6baaca]	Updating VMware container image URLs to VMware Harbor
- [403b586]	Removing symlink to /etc/resolv.conf
- [9d92de5]	Updating the new PhotonOS ISO URL
- [c2f3993]	Renaming namespace vmware to vmware-system
- [891cd96]	Fix namespace bug in prometheus YAML + use apply
- [ca00305]	Deploy local storage provisioner only for embedded Knative
- [625aa14]	Updating files to help users consume knative vs openfaas functions
- [6ef86a3]	vSphere tag synchronization to NSX-T
- [811f65d]	Integrate Embedded Knative Broker
- [b12f536]	Deprecate vcsim provider
- [548208b]	Update to K8s 1.20.2
- [c18f911]	Configure log rotation for contrackd
- [73c772a]	Update to OpenFaaS 0.12.15
- [78a5470]	Fix metadata label
- [2ed89a4]	Added firewall requirement for access to vCenter API
- [9119e38]	Fixed handling of xml predefined entities passed via ovfEnv

<a name="v0.5.0"></a>
## [Release v0.5.0](https://github.com/vmware-samples/vcenter-event-broker-appliance/compare/v0.4.1...v0.5.0)

> Release Date: 2020-12-11

### ⚠️ BREAKING

### 📖 Commits

- [d2876b0]	Remove test from output target
- [7666144]	Update VEBA BOM to final release
- [0a38742]	Update Helm chart to final release
- [0201aea]	Update Resources with the most current articles and blog posts
- [a8afbda]	Create kn-echo example
- [6273def]	Bump OpenFaaS SDK version
- [86d11ce]	Fixing escaping credentials
- [b090158]	Update Ingress deployment to handle Knative processor
- [476b96b]	Bump version to v0.5.0 for release
- [e9f14b7]	Fix metadata in Helm Chart
- [7475285]	Add OVF dropdown for Knative processor selection
- [1773a80]	Integrate Knative Processor
- [a8bb076]	Use Waitgroup to track in-flight requests
- [f1d7bae]	Standardize graceful shutdown handling
- [358e322]	Add Knative Processor
- [5951bf4]	Optimize Router Workflows
- [3278b74]	Add Status Badges for Router Actions
- [da5f3a3]	Increase timeout
- [317570c]	Make linter standalone action (#253)
- [26aaad5]	Update README
- [c8bf2d5]	Remove verbose flag from YAML
- [7547f6c]	Add Helm chart
- [9e1c15e]	Update README
- [e9f6e88]	Remove unused color package
- [d27c7ed]	Update integration tests
- [388c454]	Use structured logging
- [fd164df]	Refactored to enable filtering against all data fields Can specify if all defined filters must match Handles recursing in to dicts/lists in the event data Added functionality to use 'n' in numeric indicies Added faasstack support Unit tests improved Documentation updates and improved logging
- [45e418b]	Add details about vCenter events to README
- [0dd4cdc]	Optimize pattern map locking
- [4a7ffde]	Fix AWS enum in schema
- [61ecab2]	Update docs
- [01b6db0]	Reflect processor package changes in main
- [ec3fb41]	Add retries to OpenFaaS processor
- [925255f]	Move AWS processor to own package
- [ae0d836]	Add invocation details to metrics
- [67e4747]	Update processor interface
- [b8fb290]	Always use latest certificates
- [3f013da]	Fix lock in vcenter
- [7799c9f]	Include all files in gofmt
- [406e29d]	Fix BOM version change in integration tests
- [9fde8e7]	Address review Frankie
- [bafa586]	Add vcsim as event provider (#2134)
- [d95529b]	Function for plugin auto-refresh
- [0dbb217]	Changing development branch to use development container and simplify build script
- [9ba03c6]	Remove set-env references
- [1c41453]	Proposed fix for issue [#211](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues/211) (#211)
- [dde2464]	Update create-docker-dev-image.yml
- [2b8b852]	Moving VMware Event Router section to top for ease of edit
- [0a22792]	Renaming version in BOM to represent Github Repo Tag
- [fa832c1]	Removing VERSION file, no longer being used
- [4bb5315]	Add event UUID to checkpoint
- [36c0d15]	Removing Packer vnc_disable_password
- [96aea76]	Migrate VEBA configuration to YAML
- [c75f0f0]	Fixing Contour envoy.yaml due to changes introduced in v1.4.0 + use latest Contour v1.9.0
- [4b20765]	Implement at-least-once delivery via checkpointing
- [265daf6]	Build development image on push
- [e201fab]	fixed bug in build.sh on CentOS Added support for Packer 1.6.x to support ESXi 7.0+ added min packer version Updates to docs
- [9dcc80f]	Add workflow to build image on dev branch
- [61a2088]	Update echo function to Python3
- [d793958]	VM backup function via VEEAM
- [3870e6a]	Update build and README
- [b27af5c]	Update documentation and deployment files
- [e46b128]	Add JSON schemagen
- [75fefdc]	Reflect changes in build files
- [a719289]	Move router cmd to sub-folder
- [120beec]	Migrate to v1alpha1 config API using YAML
- [88b878c]	Add Pagerduty trigger example in go ([#201](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues/201)) (#201)
- [01fb656]	Add vm-reconfigure-via-tag go handle function example
- [fd2b58a]	Initial commit of pre-filter function
- [3980efb]	Shorten workflow names
- [93def7c]	Add Github Action to close stale issues
- [edceb28]	Bump action/checkout to v2
- [af7ec3b]	Update go mods
- [4d4296b]	Add integration tests
- [62e69ab]	VMware Cloud Notification Slack and Microsoft Teams Functions
- [30dc878]	New example HA Restarted VMs Email Notification
- [b907174]	Github Template to standardized Pull Requests
- [0f3c606]	Global search and replace on flings URL - vcenter->vmware
- [f7fcc4d]	Search functionality added to Documentation
- [7b5be22]	VEBA Issue and Feature Enhancement Templates
- [88bd3f2]	Update K8s, Contour, OpenFaaS to latest stable release
- [8895230]	Refactor VEBA components to reference BOM file
- [5fc30e2]	Add faas-cli version to BOM
- [7fbf2cb]	Update Docker images used in VMware Event Router
- [312d726]	Decouple from types.BaseEvent
- [977d96f]	Update Linter and Unit Test Action

<a name="v0.4.1"></a>
## [Release v0.4.1](https://github.com/vmware-samples/vcenter-event-broker-appliance/compare/v0.4.0...v0.4.1)

> Release Date: 2020-06-10

### ⚠️ BREAKING

### 📖 Commits

- [4bf45d7]	Updating docs to reflect changes with Proxy and SSH
- [a886138]	Fixing VEBA_VERSION reference for VM Notes
- [d126374]	Fix Pre-Release Image Build
- [d1c4496]	Bump version to v0.4.1 for release
- [9025804]	Add VEBA build-of-materials (BOM) file
- [f7d1687]	Change print and add comment to test
- [122e365]	Allow use of "http://" prefix in HTTPS Proxy config Updated OVF properties to suit previous commit
- [1404680]	Add unit test and implement error interface
- [37b2fec]	point at note from bundler docs
- [b4658a9]	add FQDN hint to hostname description
- [5e0e6cf]	Standarding on the relative path for files within docs
- [b9fd0f4]	Updated DCUI loads default values if veba-release unreadable
- [2fa6384]	install gems only for the user, and set a vendor path for bundle use
- [4164f74]	fix broken link
- [7f66cca]	update docs faqs and resources with monitoring and harbor blog posts
- [2140296]	Added enable SSH option to OVA install
- [ae29aed]	Add Action to block PR merge when title is WIP
- [acf040f]	Introduce Error struct in processor
- [f0fd438]	Fixed typo in processing ESCAPED_AWS_EVENTBRIDGE_EVENT_BUS variable
- [5e63b34]	Added recommendation to add appliance IP to NO_PROXY
- [dd29cf5]	Updated docs to reflect new v0.4 path to event-router-config.json
- [967c641]	Updating the container image name for consistency
- [e4c4aec]	Added proxy support for deployment scripts
- [d5e8f7e]	Add Action for Event Router Unit Tests on PRs
- [fc83d68]	Add Action to reject PRs against master
- [783fd0e]	Add Github Action for Docker Pre-Release
- [832642f]	Support datastore custom attribute as To: address

<a name="v0.4.0"></a>
## [Release v0.4.0](https://github.com/vmware-samples/vcenter-event-broker-appliance/compare/v0.3.0...v0.4.0)

> Release Date: 2020-05-11

### ⚠️ BREAKING

### 📖 Commits

- [70d47df]	Fixing minor typo in README
- [2659ced]	Use linux-esx as its optmized on VMware-based Hypervisor
- [e521d4a]	tndf upgrade isn't needed due to newer Photon OS image
- [a3ebd26]	Use VEBA_VERSION defined in build.sh so filename in OVA matches
- [3518ca5]	Update build.sh script to handle release vs master build
- [80cdc3e]	Update to latest PhotonOS image to 3.0 Rev2
- [1294430]	Fix Makefile timeout messed up
- [346323e]	Bump version to v0.4.0 for release
- [3939aa2]	Update READMEs in VMware Event Router
- [3d9d107]	Remove unnecessary vmw:ExtraConfig from OVA
- [0cec03a]	Initializing user focused and friendly website content
- [e2201f1]	Adding a Fx for Rest API Integrations with basic, anonymous or token based auth
- [13bfab4]	Add DCUI binary to files
- [7c12dc7]	Add licenses for libraries used in DCUI
- [4943a23]	Move release workflow to right folder
- [b48d053]	Add Github Action to push Docker Images
- [c2d75cb]	Document deploying VEBA as K8s App
- [1e5fcf3]	Replace Weave with Antrea CNI + required configuration changes
- [459e589]	Delete greetings workflow
- [db43ef7]	Add input entry batching
- [65c71b8]	Add PagerDuty Python Example (tested with VEBA v0.3, vCenter 6.7 with VMPowerOn/Off Events)
- [c0caac1]	Consistent use of Notes in markdown files
- [10e14a8]	Fix VMware Event Router image pull to support air-gap scenario
- [7af6632]	Add Docker image for :VERSION tag (#102)
- [53293b0]	Handling special character which must be escaped in Event Router JSON configuration
- [9bf3702]	Updated to pull Weave YAML from Github rather than from dynamic URL + updated Weave version
- [6bf1e8b]	Add Otto to start page
- [b6421db]	Add official VEBA mascot ([#98](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues/98)) (#98)
- [7addae5]	Updating Troubleshooting docs with correct path to config file
- [4cc3083]	Cleaning up dev template to simplify contributions
- [7a2c0d7]	Reorganize all VEBA config files into /root/config
- [73c9377]	Add example function using Go and govmomi that attaches tag to VM
- [54adab3]	Add initial release of troubleshooting guide
- [c73bf9b]	Adding /etc/veba-release to include VEBA version, commit ID & event processor type
- [7d3b92f]	Remove User Stories ([#84](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues/84)) (#84)
- [902e376]	Add greeting action on pull requests ([#88](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues/88)) (#88)
- [10a536f]	Revert Github action for stale issues/PRs
- [015fa43]	Add Github Action for stale issues/PRs
- [55a0d0d]	Fix rule processing switch statements
- [db01ee1]	Make EventBridge client interface (#69)
- [4d30278]	Support customization of Docker Bridge Network ([#76](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues/76)) (#76)
- [beb7637]	vRO Function
- [11500b7]	Spruce up README with a few badges

<a name="v0.3.0"></a>
## [Release v0.3.0](https://github.com/vmware-samples/vcenter-event-broker-appliance/compare/v0.2.0...v0.3.0)

> Release Date: 2020-03-10

### ⚠️ BREAKING

### 📖 Commits

- [976adc5]	Clarify resync period of AWS EventBridge Processor (#68)
- [d5dc652]	Ensure we pull latest vmware-event-router image
- [61f98e2]	Fixed OpenFaaS admin password
- [d083a45]	Added unauthenticated SMTP and green status emails ([#72](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues/72)) (#72)
- [993eea3]	Fixing FQDN in /etc/issue
- [1c97a8b]	Fixing syntax issue w/creating tools.conf
- [f019970]	Bump version to v0.3.0 for release
- [dd1330e]	Fixed MoRef procesisng for v0.3
- [81587c7]	Ensuring eth0 interface is shown first in vSphere UI
- [8c2327d]	Colorized log output
- [75414da]	Fix linter errors (#58)
- [dd3ca14]	Updated Getting Started User Guide
- [0e193ee]	Stricter linting on VMware Event Router
- [84eb278]	Add branch information to Python examples
- [ac6ddbe]	Test automation script samples to deploy VEBA using either OpenFaaS or EventBridge Processor
- [6e597d7]	Added drop-down menu for Network CIDR selection + clear OVF properties for security
- [3a56753]	Updated functions to support v0.3
- [75e56d1]	Fix disabling SSH
- [c9be69e]	Pull Event Router Image 1683830 + updated Event Router to include stats deployment
- [865c402]	Refactored setup.sh to just process OVF properties and introduce sub-setup scripts for configurations
- [0a7b24f]	Run TinyWWW in VMware namespace
- [a47145b]	Reorganize OVF properties to incoroprate flexible Event Processors + Added support for Event Bridge
- [f37ee3c]	Updated Packer build files to incoroprate refactored setup scripts, new OVF params, invalid Packer option + local test env
- [ec7977e]	Updating VEBA Version in OVA build
- [f871128]	Refactored Installation scripts, Pull Event Router Image + Remove vc-connector
- [3d90bb3]	Update Python examples for v0.3 release
- [55b287c]	Consolidate docs on architecture
- [be766b4]	Use ErrorGroup Context to return on first error
- [7b52f9e]	Update docs for new VMware Event Router
- [7cb1c84]	Update .gitignore for VMware Event Router
- [6c34f73]	Update Python Examples for CloudEvents
- [811f75d]	Implement VMware Event Router
- [9a5be39]	Add commit best practices link to CONTRIBUTING
- [168b2bf]	replace references of [IP] to [hostname]

<a name="v0.2.0"></a>
## [Release v0.2.0](https://github.com/vmware-samples/vcenter-event-broker-appliance/compare/v0.1.0...v0.2.0)

> Release Date: 2020-01-23

### ⚠️ BREAKING

### 📖 Commits

- [0b2c33a]	Bump version to v0.2.0 for release
- [9cbe137]	Proxy settings for docker adjusted
- [94dc2c6]	Added proxy settings to ova
- [fc17e8f]	Add NTP to photon-dev.xml.template photon-dev.json
- [3889768]	Add AWS EventBridge sample to example README
- [8a603d4]	Update default POD CIDR 10.99.0.0/20 & made it configurable
- [2ac99cd]	AWS EventBridge Sample
- [c00a4a1]	Added ntp settings to ova
- [f582274]	Fixed provider name, read_debug, and faas-cli typo
- [d3497b5]	Updated DNS desc + vCenter Event Mapping Details
- [03b38d2]	Add esx mtu fixer function to examples
- [8d894b4]	The default provider name in stack.yml is 'faas', but the function fails to deploy with error "['openfaas'] is the only valid "provider.name" for the OpenFaaS CLI, but you gave: faas". Changing the default provider name to 'openfaas' fixes the problem.
- [3770282]	Update FAQ with current WIP/TODO items
- [74f4480]	Fixed typo 'read_debuge -> read_debug'
- [df02bdd]	WIP: FAQ
- [9ae58a7]	Compliance and Documentation changes
- [8926cac]	Add host maintenance example to README
- [a5a8ce9]	New example for host maintenance
- [a84a772]	Pre-Download k8s Docker Images for non-internet require setup
- [b35a3e2]	new example VM reconfigure to Slack

<a name="v0.1.0"></a>
## v0.1.0

> Release Date: 2019-11-25

### ⚠️ BREAKING

### 📖 Commits

- [25fdc60]	Fix dead link to contributing guide
- [cd125e6]	Add README to Python tagging example
- [c41dabc]	Add DCO info
- [2019596]	Add DCO info
- [6d41d0b]	Initial Commit
- [ac3a2d7]	Adding .gitignore template
- [fd4f8bf]	Add CONTRIBUTING template
- [e950c2a]	Add README template
