# vSphere to NSX Tag Synchronization with vRealize Orchestrator

### Author: Craig Straka (craig.straka@it-partners.com)
### Version: .1

## Description
Synchronize vSphere tags with NSX-T tags unidiretionally (vSphere is the master)
	
Script is deployed as an OpenFaas VMware Event Broker powercli function to intercept and parse vSphere events from event types:
- com.vmware.cis.tagging.attach
- com.vmware.cis.tagging.detach

## Prerequisites
- vRealize Orchestrator deplyoment
	- vSphere VASA
	- NSX-T REST Host
- You have deployed the 'com.it-partners.com.tagsync.package' Workflow package from https://github.com/IT-Partners/nsx-t_tag-sync/tree/main/vro
- VMware Event Broker Appliance (VEBA) fling (https://flings.vmware.com/vmware-event-broker-appliance) Installtion

## Notes:
- Machine names in a vCenter instance may not be unique.
- event messages coming from vCenter to VEBA does not contain uniquely identifiable VM details other than name.
- As such, Script has no way, based on limited specificity of vSphere event message of the types above, to discern correct machine other than by name.
		- As such, because NSX-T Tags are central to dynamic firewall rules and alterations to the wrong machine will be a security issue, the script exits without making changes and posts a message to the effect.
- Event data has an number of odd charachters, such as new lines, that must be accomodated to get the machine name.  
	- accomodation efforts to handle naming are ongoing as errata is reported.

### vRealize Orchestrator Notes:
Be warned, vRO experience highly recommended if you want to implement this (but it's super cool)
- Self contained VEBA (OpenFaaS) function.  
- Intercepts tagging event messages and engages vRO workflow through a REST API POST to read (vSphere) and write (NSX-T) tags.
- vRO workflow uses VAPI (vSphere) and REST (NSX) calls to find the specified VM's vSphere tags and apply them to the associated object in NSX-T.
	- VAPI resource must already exist in vRO with an account that has rights to read tags.
	- NSX REST resource must already exist in vRO with an account that has rights to write tags.
- vRO Workflow is included in the git repository and must be imported to vRO.

## Tested with:
- VEBA .5 (OpenFaaS)
- vSphere 7.0.1
- NSX-T 3.0.2
- VMware vRealize Orchestrator 8.2 - Standalone (no VRA) - vRO script only.

## Planned enhancements (script today is minimally viable):
- Updates based on vSphere event changes
	- Hopefully, better vSphere and NSX-T integration will render this script obsolete.
- Ability to exclude vSphere tag categories if needed

## Instruction Consuming Function

Step 1 - Initialize function, only required during the first deployment

```
faas-cli template pull
```

Step 2 - Update `stack.yml` and `vro-secrets.json` with your environment information

> **Note:** If you are building your own function, you will need to update the `image:` property in the stack.yaml to point to your own Dockerhub account and Docker Image (e.g. `<dockerhubid>/<dockerimagename>`)

Step 3 - Deploy function to VMware Event Broker Appliance

```
VEBA_GATEWAY=https://phxlvveba01.itplab.local
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli login --username admin --password-stdin --tls-no-verify # login with your admin password
faas-cli secret create vro-secrets --from-file=vro-secrets.json --tls-no-verify # create secret, only required once
faas-cli deploy -f stack.yml --tls-no-verify
```

Step 4 - To remove the function and secret from VMware Event Broker Appliance

```
VEBA_GATEWAY=https://phxlvveba01.itplab.local
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli remove -f stack.yml --tls-no-verify
faas-cli secret remove vro-secrets --tls-no-verify
```

## Instruction Building Function

Follow Step 1 from above and then any changes made to your function, you will need to run these additional two steps before proceeding to Step 2 from above.

Step 1 - Build the function container

```
faas-cli build -f stack.yml
```

Step 2 - Push the function container to Docker Registry (default but can be changed to internal registry)

```
faas-cli push -f stack.yml
```
