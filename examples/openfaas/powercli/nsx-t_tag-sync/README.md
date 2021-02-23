# vSphere to NSX Tag Synchronization

### Version: .5 (aligned with VEBA)

## Description
Synchronize vSphere tags with NSX-T tags unidiretionally (vSphere is the master)
	
Script is deployed as an OpenFaas VMware Event Broker powercli function to intercept and parse vSphere events from event types:
- com.vmware.cis.tagging.attach
- com.vmware.cis.tagging.detach

## Prerequisites
- VMware Event Broker Appliance (VEBA) fling (https://flings.vmware.com/vmware-event-broker-appliance) Installtion

## Notes:
- Machine names in a vCenter instance may not be unique.
- event messages coming from vCenter to VEBA does not contain uniquely identifiable VM details other than name.
- As such, Script has no way, based on limited specificity of vSphere event message of the types above, to discern correct VM other than by name.
- As NSX-T Tags are central to dynamic firewall rules, and alterations to the wrong machine will be a security issue, if duplicate VM names are detected in vCenter the script exits without making changes and posts a message to the effect.

## Tested with:
- VEBA .5 (OpenFaaS)
- vSphere 7.0.1
- NSX-T 3.0.2

## Planned enhancements (script today is minimally viable):
- Updates based on vSphere event changes
- Hopefully, better vSphere and NSX-T integration will render this script obsolete.
- Ability to exclude vSphere tag categories if needed

## Instruction Consuming Function

Step 1 - Initialize function, only required during the first deployment

```
faas-cli template pull
```

Step 2 - Update `stack.yml` and `nsx-secrets.json` with your environment information

> **Note:** If you are building your own function, you will need to update the `image:` property in the stack.yaml to point to your own Dockerhub account and Docker Image (e.g. `<dockerhubid>/<dockerimagename>`)

Step 3 - Deploy function to VMware Event Broker Appliance

```
VEBA_GATEWAY=https://phxlvveba01.itplab.local
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli login --username admin --password-stdin --tls-no-verify # login with your admin password
faas-cli secret create nsx-secrets --from-file=nsx-secrets.json --tls-no-verify # create secret, only required once
faas-cli deploy -f stack.yml --tls-no-verify
```

Step 4 - To remove the function and secret from VMware Event Broker Appliance

```
VEBA_GATEWAY=https://phxlvveba01.itplab.local
export OPENFAAS_URL=${VEBA_GATEWAY} # this is handy so you don't have to keep specifying OpenFaaS endpoint in command-line

faas-cli remove -f stack.yml --tls-no-verify
faas-cli secret remove nsx-secrets --tls-no-verify
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
