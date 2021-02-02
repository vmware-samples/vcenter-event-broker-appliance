# vSphere to NSX Tag Synchronization

### Author: Craig Straka (craig.straka@it-partners.com)
### Version: .1 (Beta)
### Master stored at: https://github.com/IT-Partners/nsx-t_tag-sync


## Description
Synchronize vSphere tags with NSX-T tags unidiretionally (vSphere is the master)
	
Script is deployed as an OpenFaas VMware Event Broker powercli function to intercept and parse vSphere events from event types:
- com.vmware.cis.tagging.attach
- com.vmware.cis.tagging.detach

## Implementation details:
Their are two functional environments stored in this repository, they are stand-alone in that neither requires 
the other to run... pick one, they do the same thing.
	
Both operate within the VMware Event Broker Appliance (VEBA) fling: 
(https://flings.vmware.com/vmware-event-broker-appliance) with the OpenFaaS implementation.

### The environments are:

#### vRealize Orchestrator: https://github.com/IT-Partners/nsx-t_tag-sync/tree/main/vro
Be warned, vRO experience highly recommended if you want to implement this (but it's super cool)
- Self contained VEBA (OpenFaaS) function.  
- Intercepts tagging event messages and engages vRO workflow through a REST API POST to read (vSphere) and write (NSX-T) tags.
- vRO workflow uses VAPI (vSphere) and REST (NSX) calls to find the specified VM's vSPhere tags and apply them to the associated object in NSX-T.
	- VAPI resource must already exist in vRO with an account that has rights to read tags.
	- NSX REST resource must already exist in vRO with an account that has rights to write tags.
- vRO Workflow is included in the git repository and must be imported to vRO.

#### PowerCLI: https://github.com/IT-Partners/nsx-t_tag-sync/tree/main/nsx
PowerCLI based, so a little more friendly.  Script does all the parsing and transformation work and then POSTS to the NSX-T REST API directly.
- Self contained VEBA (OpenFaaS) powercli function.
- Intercepts tagging event messages from vSphere and sync the associated VM's tags to NSX-T.
- Invokes Web request to POST a REST API call to NSX-T to apply the same tags to the associated object in NSX-T.

## Tested with:
- VEBA .5 (OpenFaaS)
- vSphere 7.0.1
- NSX-T 3.0.2
- VMware vRealize Orchestrator 8.2 - Standalone (no VRA) - vRO script only.


## Notes:
- Machine names in a vCenter instance must be unique, as such:
	- Script has no way, based on limited specificity of vSphere event message of the types above, to discern correct machine other than by name.
	- Lots of issues here as the name may not be unique in the cluster.  
		-As such, because NSX-T Tags are central to dynamic firewall rules and alterations to the wrong machine will be a security issue, the script exits without making changes and posts a message to the effect.
- Event data has a number of odd charachters, such as new lines, that must be accomodated to get the machine name.  
	- accomodation efforts to handle naming are ongoing as errata is reported.

## Planned enhancements (script today is minimally viable):
- Updates based on vSphere event changes
	- Hopefully, better vSphere and NSX-T integration will render this script obsolete.
- VMware vRealize Orchestrator 8.2 - Standalone (no VRA)
- Ability to exclude vSphere tag categories if needed
