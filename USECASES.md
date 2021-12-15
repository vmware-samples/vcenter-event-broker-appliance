# Community Use Cases

This page contains a collection of use cases from users on how they are consuming the VMware Event Broker Appliance (VEBA) solution.

This is not an exhaustive list and we welcome additional contributions ([Submit Github Pull Request](https://vmweventbroker.io/community)) from the community in sharing their use cases that can help both new and existing users of the VEBA solution.

---

1. Send a Slack notification when a VM is powered off
1. Apply vSphere Tag when a VM is powered on
1. Synchronize between vSphere Tags and NSX-T Tags
1. Send a Slack notification when a VM has been reconfigured
1. Disable vSphere Alarms for an ESXi host when going into maintenance mode and re-enable vSphere Alarms when host exists maintenance mode
1. Send an Email report containing the list of VMs that were restarted by vSphere HA
1. Send an Email notification when a vSphere Datastore reaches a certain usage threshold (warning/error)
1. Send a Slack notification when a VMware Cloud on AWS SDDC has completed provisioning using VMware Cloud Notification Gateway (NGW) service
1. Send a Microsoft Teams notification when a VMware Cloud on AWS SDDC has completed provisioning using VMware Cloud Notification Gateway (NGW) service
1. Execute a vRealize Orchestrator (vRO) workflow using the vRO REST API when a particular vSphere event occurs
1. Create a PagerDuty incident ticket when a host is no longer responsive
1. Create a ServiceNow ticket when a host is no longer responsive
1. Automatically backup VM using Veeam Backup on VM configuration changes
1. Apply vSphere Custom Attributes when a particular vSphere event occurs
1. Run a scheduled job (cron) for managing VM snapshot retention policies (age/size of VM snapshot)
1. Send a Telegram notification when a VM has been successfully migrated
1. Send a Slack notification when a specific vSphere Horizon event occurs
1. Send a Slack notification when a specific vSphere Alarm occurs
1. Send a text message notification using Twillio when a specific vSphere event occurs
1. Ingest a custom incoming webhook to create a new CloudEvent and forward to VMware Event Router (broker)
1. Send an email notification when the password for a vSphere SSO account password has been changed
1. Automatically add instances to vRealize Operations (vROPs) based on vSphere Tags
1. Automatically associate a newly provisioned VM to a specific vSphere Resource Pool
1. Apply VM permissions based on specific vSphere Tags
1. Automatically resize VMs resources (CPU/Memory/Storage) using vSphere Tags to annotate desired state
1. Send all vCenter Create/Update/Delete (CRUD) operations to external system for compliance/security purposes
1. Send a Slack notification for a failed login attempt to vCenter Server including client IP Address
1. Send specific vCenter events to Splunk for archival purposes
1. Add a VM Annotation (notes) on who powered on the VM and from which IP Address
1. Add a VM Annotation (notes) on who powered off the VM and from which IP Address
1. Add a VM Annotation (notes) on who paused the VM and from which IP Address
1. Add a VM Annotation (notes) on who shutdown the VM and from which IP Address
1. Add a VM Annotation (notes) on who forcefully killed a VM and from which IP Address
1. Add a VM Annotation (notes) on who registered a VM and from which IP Address
1. Add a VM Annotation (notes) on who cloned a VM along with the date and the VM Template used
1. Add a VM Annotation (notes) on when an OVF is deployed with actual users versus vpxd
1. Send an email or Slack notification when an ESXi host is no longer responding
1. Send an email when an ESXi host is disconnected and update Change Management Database (CMDB)
1. Apply specific vSphere Host Profile and Tags when an ESXi host is added to vCenter Server
1. Remove unused vSphere Tags when an ESXi host is removed from vCenter Server
1. Update Change Management Database (CMDR) on VM location when it is vMotion
1. Send a Slack notification when a VM is removed containing the VM path to ensure no files are left over
1. Update vRealize Automation (vRA) image mapping when a VM is converted to VM Template
1. Create or update network profile in vRealize Automation (vRA) when vSphere Portgroup or Opaque Network is created
1. Delete network profile in vRealize Automation (vRA) when vSphere Portgroup or Opaque Network is deleted
1. Create or update storage profile in vRealize Automation (vRA) when a new vSphere Datastore is added
1. Delete storage profile in vRealize Automation (vRA) when a vSphere Datastore is delete
1. Apply specific storage policy when a vSphere Datastore is added
1. When a Tanzu Kubernetes Grid (TKG) Cluster is created, automatically apply specific vSphere Tags with K8s name, backup exemption and usage to both Control and Worker Node VMs
1. Update Change Management Database (CMDR) on when a VM is created
1. Update shares in a vSphere Resource Pool (RP) when a VM has been created or deleted within an RP
1. Update shares in a vSphere Resource Pool (RP) when a VM has been added or removed from a RP
1. Update storage profile in vRealize Automation (vRA) when a Datastore Cluster has been added
1. Provision additional storage using vRealize Orchestrator (vRO) when a Datastore Cluster alerts
1. When a First Class Disk (FCD) is provisioned, update a report on number of Persistent Volume Claims (PVC) for historical data analysis
1. When a First Class Disk (FDC) is deleted, update a report to remove the number of Persistent Volume Claims (PVC)
1. When an ESXi host is added to vCenter Server, automatically create a baseline and upload results to external store
1. Run scheduled reporting against vCenter on daily basis to report on basic inventory/info
1. Power off development VMs every night based on specific vSphere Tags
1. Power on development VMs every morning based on specific vSphere Tags
1. Power off development VMs every weekend based on specific vSphere Tags
1. Power on development VMs that were powered off over the weekend based on specific vSphere Tags
1. Power off QA VMs on the weekend based on specific vSphere Tags
1. Power on QA VMs that were powered off over the weekend based on specific vSphere Tags
1. Apply VM right-sizing recommendations from vRealize Operations (vROPs) over the weekend when downtime is possible
1. Rename VM display name to lower-case if they were created with upper-case
1. Create DNS record and DHCP reservation for the network that a VM is connected to and powered on
1. When a VM is created, add it to AWX/Ansible Tower inventory based on OS,vSphere Tags and naming convention
1. When a VM is deleted, remove it from AWX/Ansible Tower
1. Create vSphere DRS Affinity Host Group for VMs that leverage SRIOV
1. Enforce Guest OS licensing to physical host cores
1. Trigger Ansible playbooks based on specific vSphere events
1. Send a notification when a vSphere Permission for an individual user and/or group has been modified
1. Send a notification when a vSphere Permission for an individual user and/or group has been deleted
1. Send a notification from the VMware Cloud Notification Gateway (NGW) service when vCenter Server TLS certificate is replaced to update the new SSL Thumbprint for solutions like VMware Horizon/vRealize Automaton (vRA)/vRealize Orchestrator (vRO)
1. Automatically refresh vSphere Client Plugin(s) based on specific data changes