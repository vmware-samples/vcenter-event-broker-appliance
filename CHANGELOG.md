
<a name="v0.7.1"></a>
## [Release v0.7.1](https://github.com/vmware-samples/vcenter-event-broker-appliance/compare/v0.7.0...v0.7.1)

> Release Date: 2021-12-15

### 🐞 Fix

- [8c16b92]	Ensure $ character is properly handled for veba-ui secret (#688) 
- [80985e3]	Update imagePullPolicy for knative-contour for air-gap deployments (#689) 

### 💫 Feature

- [be2a10f]	Add Knative vRO function (#728) 
- [39fd850]	Add Knative Go Tagging example (#729) 
- [cc7c245]	Add vRNI webhook function (#723) 
- [1e35d96]	Add VM Preemption Example (#696) 
- [da053d5]	update website with new veba logo version (#709) 
- [ed4fbd8]	replace tanzu symbol on veba logo (#707) 
- [d7c10d3]	Add Knative NSX tag sync example (#684) 
- [178283e]	Support handling hostname when using all caps

### 📃 Documentation

- [8ea0cb7]	Minor update to the VEBA timeline for v0.7.1 release (#745) 
- [01dcc02]	Add v0.7.1 release to VEBA timeline (#742) 
- [f8c0bc8]	Add Windows instructions to OVA deployment scripts
- [cf33f1e]	Add new functions to website
- [a16cf18]	Update style headers
- [f64091b]	Update Event Router installation on Kind (#698) 
- [d8c355c]	Fix variable rendering in PCLI tag sync (#711) 
- [388b1c2]	Add correct VMware Fling URl to VEBA website (#704) 
- [2851b3c]	Update website adv install (#676) 
- [d4affc1]	Fix community meeting time (#701) 
- [08c7ac8]	Default to Knative Function examples on VEBA website (#699) 
- [b635047]	Add a Community Use Cases document (#687) 
- [11cd0eb]	Update website README to include Windows instructions for Jekyll container
- [d9441f0]	Removing unused zcleanup dir/files (#654) 
- [2b58e3c]	Add video tutorial link to kn-ps-slack README

### 🧹 Chore

- [afd35d6]	Updating website styling per branding guidelines
- [e80f0af]	Update VEBA logo in packer build to v2 (#718) 
- [9aefd3c]	Add VEBA OVA deployment scripts for Knative (#714) 
- [e69a7f4]	Update VEBA OVA filename to use VMware instead of vCenter (#685) 
- [40214b8]	Add note on community calls (#649) 

### ⚠️ BREAKING

### 📖 Commits

- [4e5b6ac]	Bump version to v0.7.1 for release
- [8ea0cb7]	docs: Minor update to the VEBA timeline for v0.7.1 release (#745)
- [afd35d6]	chore: Updating website styling per branding guidelines
- [01dcc02]	docs: Add v0.7.1 release to VEBA timeline (#742)
- [f8c0bc8]	docs: Add Windows instructions to OVA deployment scripts
- [f36ec71]	Bump version to release-0.7.1
- [cf33f1e]	docs: Add new functions to website
- [be2a10f]	feat: Add Knative vRO function (#728)
- [39fd850]	feat: Add Knative Go Tagging example (#729)
- [a16cf18]	docs: Update style headers
- [cc7c245]	feat: Add vRNI webhook function (#723)
- [f64091b]	docs: Update Event Router installation on Kind (#698)
- [1e35d96]	feat: Add VM Preemption Example (#696)
- [e80f0af]	chore: Update VEBA logo in packer build to v2 (#718)
- [9aefd3c]	chore: Add VEBA OVA deployment scripts for Knative (#714)
- [da053d5]	feat: update website with new veba logo version (#709)
- [d8c355c]	docs: Fix variable rendering in PCLI tag sync (#711)
- [ed4fbd8]	feat: replace tanzu symbol on veba logo (#707)
- [388b1c2]	docs: Add correct VMware Fling URl to VEBA website (#704)
- [2851b3c]	docs: Update website adv install (#676)
- [d4affc1]	docs: Fix community meeting time (#701)
- [08c7ac8]	docs: Default to Knative Function examples on VEBA website (#699)
- [b635047]	docs: Add a Community Use Cases document (#687)
- [d7c10d3]	feat: Add Knative NSX tag sync example (#684)
- [8c16b92]	fix: Ensure $ character is properly handled for veba-ui secret (#688)
- [80985e3]	fix: Update imagePullPolicy for knative-contour for air-gap deployments (#689)
- [e69a7f4]	chore: Update VEBA OVA filename to use VMware instead of vCenter (#685)
- [178283e]	feat: Support handling hostname when using all caps
- [8974bc6]	bug: Add deployment methods section back to website (#667)
- [11cd0eb]	docs: Update website README to include Windows instructions for Jekyll container
- [d9441f0]	docs: Removing unused zcleanup dir/files (#654)
- [40214b8]	chore: Add note on community calls (#649)
- [2b58e3c]	docs: Add video tutorial link to kn-ps-slack README

<a name="v0.7.0"></a>
## [Release v0.7.0](https://github.com/vmware-samples/vcenter-event-broker-appliance/compare/v0.6.1...v0.7.0)

> Release Date: 2021-10-05

### 🐞 Fix

- [363bfa9]	Update conntrackd config to run properly in Photon OS 4.0 (#640) 
- [ebcf411]	Update veba-dcui to handle pre-release versioning scheme (#634) 
- [813f698]	Prevent duplicate Processor entries in /etc/veba-release (#633) 
- [7fd4b69]	Formatting in server.ps1 (#623) 
- [b918211]	Fix settings process exit code in server.ps1 (#578) 
- [fa5a4d2]	Update Fluent Bit configuration to fix mem buf overlimit (#604) 
- [4ad0d30]	Fix stopping behaviour in functions templates (#573) 
- [038421c]	Prevent $ char from being eval in escaped variables (#591) 
- [f9c27fd]	Set global retry policy on broker (#580) 
- [b3da8d3]	Correct execption message for TAG_SECRET not being defined (#570) 
- [dd2f4ab]	Update Fluentbit config using YTT overlay (#572) 
- [b517128]	Fix advanced docs cert section (#567) 
- [3d12b55]	Apply server.ps1 fixes to all PS/CLI container images (#549) 
- [01124dc]	Add Horizon Ingress (#563) 
- [eaa135c]	Update TERM env variable to properly output PS hashtable in logs (#553) 
- [7c01575]	Return HTTP Bad Request on invalid CloudEvent
- [f8d8491]	Use ConvertTo-Json to output CloudEvent Data in kn-ps-echo function (#544) 
- [70f8e0b]	Prevent unexpected CLI argument quoting when using set -x (#538) 
- [9d2ad87]	Disable cloud-init to preserve FQDN hostname upon reboot (#532) 
- [d9e1001]	Update /etc/issue login banner (#528) 
- [ba2a842]	Fix Docker credentials in action (#521) 
- [6363acc]	Update Antrea to v1.2.0 to resolve deprecated k8s API resources (#502) 
- [a5b8b76]	Improve variable escapes using jq instead of python snippets (#498) 
- [8fa7085]	Replace is not with != as syntax is no longer valid for python 3.8+ (#497) 
- [8603b55]	Updated files with missing characters + updated handler.ps1 (#480) 
- [d42dd07]	Adjusted typos in tag_secret.json (#478) 
- [3b859f4]	Increased disk space for the two vmdks from 12GB to 25GB each (#476) 
- [c8678cb]	Adjusted the kubectl wait timeout values to address issue 468 (#468) 

### 💫 Feature

- [3582ac3]	Add example to monitor vSphere Inventory Resource deletion
- [a8dafea]	Add Scheduled VM Snapshot Retention using PingSource Function Example (#627) 
- [412d93f]	Update PS/PCLI images w/latest CE SDK/PowerCLI versions (#595) 
- [519f2ff]	Adds full message to the kn-ps-slack payload, enabling function to be used for any event
- [4b1828f]	Add VMware Cloud Notification Gateway example using Microsoft Teams (#590) 
- [643826b]	Add webhook function to ingress custom events (#473) 
- [b335164]	Add VMware Cloud Notification Gateway example using Slack (#589) 
- [e3e8e7b]	Add VMware Horizon Slack example function (#588) 
- [86a059f]	Add notification example using Telegram (#583) 
- [da4938d]	Update kn-ps-slack with try/catch (#598) 
- [81e945a]	Update PS/PCLI images with latest server.ps1 fixes (#594) 
- [29c6812]	Update OVF labels with deprecated event processors (#500) 
- [76838de]	Add Horizon event provider (#510) 
- [1589e6d]	Adding Evolution of VEBA to the vmweventbroker.io/evolution
- [2cd6056]	Integrate Horizon Event Provider into VEBA (#526) 
- [2a6eaa0]	Enhanced VEBA DCUI to display endpoints from /etc/veba-endpoints (#490) 
- [0180234]	Improve OVF debuggability (#537) 
- [669a3ca]	Integrate Webhook Event Provider into VEBA (#496) 
- [db815f2]	Update setup scripts to support multi-event providers (#516) 
- [11cc593]	Replace grub splash screen with VEBA logo (#522) 
- [b792c5f]	Templatize YAML downloads using YTT (#505) 
- [70a6b0b]	Upgrade to Photon OS 4.0 GA (#493) 
- [84efdf2]	Add YAML templating using Carvel YTT (#487) 
- [2bed016]	Add scaling and concurrency settings to VEBA UI (#448) 
- [2105c5c]	Add webhook provider (#462) 
- [9c5e0bc]	Add example function triggered based on vSphere Alarm (#469) 
- [f32cbca]	add resource usage and performance monitoring feature (#450) 
- [95d1ad5]	add log forwarding feature (#451) 
- [05caac1]	Add vsphereapiversion to CE context (#439) 

### 📃 Documentation

- [4c348f9]	Simplify docs testing/contribution using Jekyll Docker image (#618) 
- [3cea87a]	Updated VEBA timeline with v0.7 release (#620) 
- [da9611d]	Update event router web docs (#491) 
- [1bcb8a5]	Updated Troubleshooting Function Guide w/Knative (#610) 
- [c65d826]	Updated Getting Started Guide w/Knative (#611) 
- [fa93f70]	Updated VEBA Appliance deployment guides (#536) 
- [47eefc0]	Updated description for VEBA (#596) 
- [cb92f02]	Update Github Photon OS badge to 4.0
- [85b9c52]	highlighted new providers, updated featured functions (#561) 
- [9da6490]	add deployment methods section to welcome page (#517) 
- [def4cbe]	Update Architecture for v0.7 release (#535) 
- [db2868a]	Update kn-pcli-tag README with instructions to change vm.name property before running a test, rename vm.name property to 'REPLACE-ME'
- [b990435]	Add replace TLS cert instructions
- [15efd97]	Fix YAML annotation for scale to 1 (#474) 
- [d0fd42d]	Update PS/PCLI Images to latest v1.1 (#452) 

### 🧹 Chore

- [9ec9021]	Migrate kn examples to us.gcr.io (#644) 
- [5178c85]	Add missing Helm v0.6.6-pre-release (#629) 
- [dded4bf]	Add unit tests for server.ps1
- [aa9b535]	Bump Event Router build versions (#616) 
- [07f0930]	Update Helm chart for Horizon
- [482466c]	Update deployment manifests
- [b09e763]	Update Docker related actions (#518) 
- [bd89626]	Update workflows to Go v1.16 (#514) 
- [4aa89f3]	Update golangci-lint config (#511) 
- [69af9e4]	Update Helm chart workflow (#464)  (#466)  (#467) 
- [9c1de78]	Include commit details in BREAKING section (#445) 

### ⚠️ BREAKING

Add vsphereapiversion to CE context [05caac1]:
This change sets the `timestamp` in the CloudEvent to the
`CreatedTime` as set by vCenter in a vSphere event instead of
`time.Now()`.

### 📖 Commits

- [c74a5e5]	Bump version to v0.7.0 for release
- [9ec9021]	chore: Migrate kn examples to us.gcr.io (#644)
- [3582ac3]	feat: Add example to monitor vSphere Inventory Resource deletion
- [363bfa9]	fix: Update conntrackd config to run properly in Photon OS 4.0 (#640)
- [ebcf411]	fix: Update veba-dcui to handle pre-release versioning scheme (#634)
- [813f698]	fix: Prevent duplicate Processor entries in /etc/veba-release (#633)
- [c9890cf]	Bump version to release-0.7.0
- [a8dafea]	feat: Add Scheduled VM Snapshot Retention using PingSource Function Example (#627)
- [5178c85]	chore: Add missing Helm v0.6.6-pre-release (#629)
- [412d93f]	feat: Update PS/PCLI images w/latest CE SDK/PowerCLI versions (#595)
- [519f2ff]	feat: Adds full message to the kn-ps-slack payload, enabling function to be used for any event
- [4c348f9]	docs: Simplify docs testing/contribution using Jekyll Docker image (#618)
- [7fd4b69]	fix: Formatting in server.ps1 (#623)
- [3cea87a]	docs: Updated VEBA timeline with v0.7 release (#620)
- [dded4bf]	chore: Add unit tests for server.ps1
- [aa9b535]	chore: Bump Event Router build versions (#616)
- [da9611d]	docs: Update event router web docs (#491)
- [b918211]	fix: Fix settings process exit code in server.ps1 (#578)
- [1bcb8a5]	docs: Updated Troubleshooting Function Guide w/Knative (#610)
- [c65d826]	docs: Updated Getting Started Guide w/Knative (#611)
- [4b1828f]	feat: Add VMware Cloud Notification Gateway example using Microsoft Teams (#590)
- [643826b]	feat: Add webhook function to ingress custom events (#473)
- [fa93f70]	docs: Updated VEBA Appliance deployment guides (#536)
- [b335164]	feat: Add VMware Cloud Notification Gateway example using Slack (#589)
- [fa5a4d2]	fix: Update Fluent Bit configuration to fix mem buf overlimit (#604)
- [47eefc0]	docs: Updated description for VEBA (#596)
- [e3e8e7b]	feat: Add VMware Horizon Slack example function (#588)
- [be093dd]	Add PowerShell SMS example using Twillo (#582)
- [86a059f]	feat: Add notification example using Telegram (#583)
- [da4938d]	feat: Update kn-ps-slack with try/catch (#598)
- [81e945a]	feat: Update PS/PCLI images with latest server.ps1 fixes (#594)
- [4ad0d30]	fix: Fix stopping behaviour in functions templates (#573)
- [038421c]	fix: Prevent $ char from being eval in escaped variables (#591)
- [cb92f02]	docs: Update Github Photon OS badge to 4.0
- [85b9c52]	docs: highlighted new providers, updated featured functions (#561)
- [f9c27fd]	fix: Set global retry policy on broker (#580)
- [9da6490]	docs: add deployment methods section to welcome page (#517)
- [b3da8d3]	fix: Correct execption message for TAG_SECRET not being defined (#570)
- [dd2f4ab]	fix: Update Fluentbit config using YTT overlay (#572)
- [29c6812]	feat: Update OVF labels with deprecated event processors (#500)
- [b517128]	fix: Fix advanced docs cert section (#567)
- [3d12b55]	fix: Apply server.ps1 fixes to all PS/CLI container images (#549)
- [def4cbe]	docs: Update Architecture for v0.7 release (#535)
- [01124dc]	fix: Add Horizon Ingress (#563)
- [07f0930]	chore: Update Helm chart for Horizon
- [482466c]	chore: Update deployment manifests
- [76838de]	feat: Add Horizon event provider (#510)
- [db2868a]	docs: Update kn-pcli-tag README with instructions to change vm.name property before running a test, rename vm.name property to 'REPLACE-ME'
- [eaa135c]	fix: Update TERM env variable to properly output PS hashtable in logs (#553)
- [1589e6d]	feat: Adding Evolution of VEBA to the vmweventbroker.io/evolution
- [7c01575]	fix: Return HTTP Bad Request on invalid CloudEvent
- [2cd6056]	feat: Integrate Horizon Event Provider into VEBA (#526)
- [2a6eaa0]	feat: Enhanced VEBA DCUI to display endpoints from /etc/veba-endpoints (#490)
- [f8d8491]	fix: Use ConvertTo-Json to output CloudEvent Data in kn-ps-echo function (#544)
- [70f8e0b]	fix: Prevent unexpected CLI argument quoting when using set -x (#538)
- [0180234]	feat: Improve OVF debuggability (#537)
- [9d2ad87]	fix: Disable cloud-init to preserve FQDN hostname upon reboot (#532)
- [15e21d3]	Fix processing parallel requested events
- [d9e1001]	fix: Update /etc/issue login banner (#528)
- [669a3ca]	feat: Integrate Webhook Event Provider into VEBA (#496)
- [db815f2]	feat: Update setup scripts to support multi-event providers (#516)
- [ba2a842]	fix: Fix Docker credentials in action (#521)
- [6363acc]	fix: Update Antrea to v1.2.0 to resolve deprecated k8s API resources (#502)
- [11cc593]	feat: Replace grub splash screen with VEBA logo (#522)
- [b09e763]	chore: Update Docker related actions (#518)
- [bd89626]	chore: Update workflows to Go v1.16 (#514)
- [4aa89f3]	chore: Update golangci-lint config (#511)
- [b792c5f]	feat: Templatize YAML downloads using YTT (#505)
- [a5b8b76]	fix: Improve variable escapes using jq instead of python snippets (#498)
- [70a6b0b]	feat: Upgrade to Photon OS 4.0 GA (#493)
- [8fa7085]	fix: Replace is not with != as syntax is no longer valid for python 3.8+ (#497)
- [84efdf2]	feat: Add YAML templating using Carvel YTT (#487)
- [4d66620]	Add Helm release v0.6.5 (#485)
- [2bed016]	feat: Add scaling and concurrency settings to VEBA UI (#448)
- [b990435]	docs: Add replace TLS cert instructions
- [4a7bf8a]	Bump Helm chart version
- [8603b55]	fix: Updated files with missing characters + updated handler.ps1 (#480)
- [d42dd07]	fix: Adjusted typos in tag_secret.json (#478)
- [3b859f4]	fix: Increased disk space for the two vmdks from 12GB to 25GB each (#476)
- [15efd97]	docs: Fix YAML annotation for scale to 1 (#474)
- [2105c5c]	feat: Add webhook provider (#462)
- [9c5e0bc]	feat: Add example function triggered based on vSphere Alarm (#469)
- [f32cbca]	feat: add resource usage and performance monitoring feature (#450)
- [c8678cb]	fix: Adjusted the kubectl wait timeout values to address issue 468 (#468)
- [95d1ad5]	feat: add log forwarding feature (#451)
- [69af9e4]	chore: Update Helm chart workflow (#464) (#466) (#467)
- [9267b2f]	Add `[BUG]` to issue template title
- [d0fd42d]	docs: Update PS/PCLI Images to latest v1.1 (#452)
- [9c1de78]	chore: Include commit details in BREAKING section (#445)
- [05caac1]	feat: Add vsphereapiversion to CE context (#439)

<a name="v0.6.1"></a>
## [Release v0.6.1](https://github.com/vmware-samples/vcenter-event-broker-appliance/compare/v0.6.0...v0.6.1)

> Release Date: 2021-06-10

### 🐞 Fix

- [fd738b3]	Resolve CVE-2021{22901,22898,20266} for PS Base Image (#443) 
- [2fcac13]	Update Knative PS examples for consistency
- [1c78537]	Fix command in adding trusted root CA cert documentation
- [a8c1976]	Update documentation to reflect minimum of vCenter Server 7.0 for VEBA UI
- [0356447]	Wrong association in greeting

### 💫 Feature

- [e6371a6]	add kn-py-echo example (#426) 
- [c37b8e2]	Add Knative Python VM Attribute Example (#434) 
- [509e763]	Add Knative Base PowerShell/PowerCLI Container Images
- [12ac7ac]	Add Knative PowerCLI vSphere Tagging Example
- [e3b02a5]	Add kn-py-slack Python example
- [eefe0d2]	Document Trusted Root Certificate support w/VMware Event Router
- [b2507eb]	Add router support for custom certificates (#370) 
- [4d503c9]	Add Helm Option for Knative

### 📃 Documentation

- [4e4c8f7]	Add new Python/Go examples to Docs
- [8bb8949]	Update Knative Function list w/PowerShell Email Example

### 🧹 Chore

- [59590e8]	Update CHANGELOG template (#429) 
- [0666a0a]	Add CHANGELOG workflow
- [a5a91d0]	Daily build and helm verification (#414) 
- [2ef2aa1]	Automate CHANGELOG (#394) 
- [d0d93c8]	Add issue greeting (#390) 

### ⚠️ BREAKING

### 📖 Commits

- [992f8fb]	Bump version to v0.6.1 for release
- [fd738b3]	fix: Resolve CVE-2021{22901,22898,20266} for PS Base Image (#443)
- [0e3cfb8]	Bump version to v0.6.1 for release
- [b894ab6]	Bump requests lib in OpenFaaS Python fns (#440)
- [4dd5870]	chore(deps): Bump urllib3
- [2eb6ae9]	chore(deps): Bump urllib3
- [8d0dcb9]	Use 1.0 tag for image (#437)
- [e6371a6]	feat: add kn-py-echo example (#426)
- [4e4c8f7]	docs: Add new Python/Go examples to Docs
- [c37b8e2]	feat: Add Knative Python VM Attribute Example (#434)
- [cbeb5bd]	chore(deps): Bump urllib3 in /examples/openfaas/python/tagging/handler
- [59590e8]	chore: Update CHANGELOG template (#429)
- [509e763]	feat: Add Knative Base PowerShell/PowerCLI Container Images
- [2fcac13]	fix: Update Knative PS examples for consistency
- [12ac7ac]	feat: Add Knative PowerCLI vSphere Tagging Example
- [6367ee0]	Bump requests in /examples/knative/python/kn-py-slack
- [e3b02a5]	feat: Add kn-py-slack Python example
- [02bd23c]	Example Knative echo service written in Go
- [0666a0a]	chore: Add CHANGELOG workflow
- [ceaa976]	Update stale action (#420)
- [a5a91d0]	chore: Daily build and helm verification (#414)
- [c3b670d]	Clarify vcsim deprecation (#408)
- [8bb8949]	docs: Update Knative Function list w/PowerShell Email Example
- [761d41c]	Add docs section to CHANGELOG (#404)
- [1c78537]	fix: Fix command in adding trusted root CA cert documentation
- [a8c1976]	fix: Update documentation to reflect minimum of vCenter Server 7.0 for VEBA UI
- [e859262]	Updated docs with new URL www.williamlam.com
- [eefe0d2]	feat: Document Trusted Root Certificate support w/VMware Event Router
- [0356447]	fix: Wrong association in greeting
- [2ef2aa1]	chore: Automate CHANGELOG (#394)
- [b2507eb]	feat: Add router support for custom certificates (#370)
- [5e7c2e4]	Document simplified steps for replacing TLS certifcate in VEBA
- [121c688]	Bump Helm chart version to v0.6.2
- [f03672e]	Add v0.6.2 Chart Pre-Release
- [4d503c9]	feat: Add Helm Option for Knative
- [d0d93c8]	chore: Add issue greeting (#390)
- [168abeb]	Example Knative PowerShell Email Function
- [11a83cd]	Update WIP Action (#382)
- [7f8177f]	Fix sed command
- [5169b39]	Configure container log rotation
- [bf21bbd]	Add support for custom VEBA TLS Certificate
- [d288918]	Refactor Ingress Configuration based on Processor Type
- [46d7e5d]	Bump urllib3 in /examples/openfaas/python/tagging/handler
- [8fa4b40]	Bump urllib3 in /examples/openfaas/python/invoke-rest-api/handler
- [db9399e]	Bump urllib3
- [c778a0a]	VEBA UI fix for incorrect TLS miss-match
- [33ed689]	Add dispatcher container image to VEBA BOM for RabbitMQ Broker deployment
- [c7da5cc]	Add correct AWS Event Bridge Type into Event Router Config
- [6f33d65]	Update the correct cert name for TLS replacement (#367)
- [016bd9a]	Document multiple Knative Triggers
- [166f396]	Document minimum vSphere Privileges for VEBA UI
- [7a24aef]	Verify Helm chart (#359)
- [b567d65]	Update helm chart (#355) (#357)
- [72bb151]	Fix [#355](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues/355) by helm-ignoring releases folder (#355)
- [e234437]	Remove * in front of Closes keyword + Fix Typo
- [7f246ef]	Add Windows specific Docker command to kn-ps-slack function

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
