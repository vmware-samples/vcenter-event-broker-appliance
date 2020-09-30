# Automatic Backup of Virtual Machines via Veeam Backup & Replication

## Description

Veeam-vm-backup function uses 3rd party solution Veeam to provide automatic backup for any virtual machine when when the VM state changes. 

The function is written in Java - Java 11 by default, but Java 8 is supported, too. It gets called on a VM state changed event. Then it [logs on](https://helpcenter.veeam.com/docs/backup/rest/post_sessionmngr.html?ver=100) to the Veeam's server, finds the hierarchyRoot ([per Veeam documentation](https://helpcenter.veeam.com/docs/backup/rest/get_hierarchyroots.html?ver=100)) and constructs the VM reference per Veeam format `urn:VMware:vm:VEEAM_HOST_ID:VM_VALUE`. For example: urn:VMware:vm:3f3fe831-79b8-4123-9ff9-3d5df7e89c9f.vm-28, which is different from the VMware one, so that the actual [backup](https://helpcenter.veeam.com/docs/backup/rest/post_backupservers_id_quickbackup.html?ver=100) succeeds.

## Prerequisites

Install Veaam Backup & Replication and add your VC to the list of Veeam's managed servers.
Details on how to do this can be found here: https://www.veeam.com/veeam_backup_10_0_quick_start_guide_vsphere_pg.pdf

## Get the example function

Clone this repository which contains the example functions. 

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/java/veeam-vm-backup
git checkout master
```

## Customize the function

For security reasons to not expose sensitive data we will create a Kubernetes secret (https://kubernetes.io/docs/concepts/configuration/secret/) which will hold Veeam credentials and information, as well as the vCenter ip. This secret will be mounted into the function during runtime. This is all taken care of for your by the appliance. We only have to create the secret with a simple command through `faas-cli`.

First, change the configuration file `veeamconfig.toml` holding your secret vCenter information located in the veeam-vm-backup example folder:

```toml
# veeamconfig.toml contents
# replace with your own values
[enterprise_manager]
ip = "VEEAM_ENTERPRISE_MANAGER_IP"
port = "9398"	# This is the default port for Veeam's Restful API. Change it, if you changed the port during Veeam configuration. 
user = "USERNAME"
password = "PASSWORD"

[backup_server]
id = "BACKUP_SERVER_ID"  # Can be found through the request to your Veeam server described here: https://helpcenter.veeam.com/docs/backup/rest/get_backupservers.html?ver=100

[vcenter]
ip = "VC_IP"
```

Now go ahead and store this configuration file as secret in the appliance.

```bash
# set up faas-cli for first use
export OPENFAAS_URL=https://VEBA_FQDN_OR_IP
faas-cli login -p VEBA_OPENFAAS_PASSWORD --tls-no-verify # vCenter Event Broker Appliance is configured with authentication, pass in the password used during the vCenter Event Broker Appliance deployment process

# now create the secret
faas-cli secret create veeamconfig --from-file=veeamconfig.toml --tls-no-verify
```

> **Note:**  Delete the local `veeamconfig.toml` after you're done with this exercise to not expose this sensitive information.

Lastly, define the function-specific settings. They are performed in the `veeam-vm-backup.yml` file. Open and edit the `veeam-vm-backup.yml` provided with in the Java veeam-vm-backup example code. Change `gateway` and (optionally) `topic` as per your environment/needs.

```yaml
version: 1.0
provider:
  name: openfaas
  gateway: https://VEBA_FQDN_OR_IP # replace with your vCenter Event Broker Appliance environment
functions:
  veeam-vm-backup:
    lang: java11 # or java8 for the java8 openfaas template
    handler: ./veeam-vm-backup
    image: vmware/veba-java-veeam-vm-backup:latest
    secrets:
      - veeamconfig # leave as is unless you changed the name during the creation of the secrets above
    annotations:
      topic: VmPoweredOffEvent  # the event can be changed, in case the aim is to back up a VM when another type of event occurs
```

## Deploy the function

After you've performed the steps and modifications above, you can go ahead and deploy the function:

```bash
faas-cli template pull # only required during the first build/deployment

# Deploy the function
faas-cli deploy -f veeam-vm-backup.yml --tls-no-verify
Deployed. 202 Accepted.
```

## Trigger the function

Turn off a virtual machine in vCenter to trigger the function via a `VmPoweredOffEvent`. Verify the virtual machine was backed up successfully.

> **Note:** If you don't see a backup for the VM verify that you correctly followed each step above, IPs/FQDNs and credentials are correct and see the [troubleshooting](#troubleshooting) section below.

## Troubleshooting

If your VM did not get the tag attached, verify:

- Veeam IP/username/password
- Backup Server ID
- Permissions of the vCenter user
- Whether the components can talk to each other (connector to vCenter and OpenFaaS, function to Veeam)
- Check the logs (`kubectl` is installed and configured locally on the appliance):

Get the function's pod and check the logs:
```bash
kubectl get pods -n openfaas-fn

# Somewhere in the pods should be our veeam-vm-backup function
veeam-vm-backup-56cdc9485f-9r4h7             1/1     Running   0          4d20h



kubectl logs veeam-vm-backup-56cdc9485f-9r4h7 -n openfaas-fn

# Successful log message
2020/08/26 13:24:57 Started logging stderr from function.
2020/08/26 13:24:57 Started logging stdout from function.
Forking - java [-XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap com.openfaas.entrypoint.App]
2020/08/26 13:24:57 OperationalMode: http
2020/08/26 13:24:57 Timeouts: read: 10s, write: 10s hard: 10s.
2020/08/26 13:24:57 Listening on port: 8080
2020/08/26 13:24:57 Writing lock-file to: /tmp/.lock
2020/08/26 13:24:57 Metrics listening on port: 8081
2020/08/26 13:46:52 stdout: Reading toml configuration data started
2020/08/26 13:46:52 stdout: Reading toml configuration data finished
2020/08/26 13:46:52 stdout: Credentialds encoded!
2020/08/26 13:46:52 stdout: Logon request created!
2020/08/26 13:46:53 stdout: Logon request sent!
2020/08/26 13:46:53 stdout: Session ID saved!
2020/08/26 13:46:53 stdout: Hierarchy roots request created!
2020/08/26 13:46:53 stdout: Hierarchy roots request sent!
2020/08/26 13:46:53 stdout: VM reference received!
2020/08/26 13:46:53 stdout: Backup request body is done!
2020/08/26 13:46:53 stdout: Backup request created!
2020/08/26 13:46:53 stdout: Backup request sent!
2020/08/26 13:46:53 POST / - 200 OK - ContentLength: 416
2020/08/26 13:46:53 stdout: Request / 416 bytes written.
```

Info for the requests used can be found in the Veeam Backup Enterprise Manager RESTful API Reference: https://helpcenter.veeam.com/docs/backup/rest/em_web_api_reference.html?ver=100

## Develop the function

You can make changes to the function's code and configuration. This requires building and publishing the image again.

### Set up the function with Java 8

To set up the function with java 8 openfaas template, you must create new function:

```bash
faas-cli new --lang java8 veeam-vm-backup
```

Then, replace the Handler.java and the veeam-vm-backup.yml in the new function with the relevant files from the function repo. Don't forget to change the function's language and the imаge name and docker repository for the function in the .yml file. Also, delete the "extends com.openfaas.model.AbstractHandler" part from the declaration of the class 'Handler' in Handler.java.
Copy the missing implementation dependencies (omitting the runtime ones) from the build.gradle file in the function repo and paste them to the new function's build.gradle. 

Now, the java8 veeam-vm-backup function is set up and you can build and push it to your repository.

### Build the function

To build your own image instead of using the pre-built one, change the 'image' field in the .yml file, so that it contains your docker repository and image name. Then you can execute the commands below to build and push the image to your repository.
```bash
faas-cli build -f veeam-vm-backup.yml
faas-cli push -f veeam-vm-backup.yml
```

When the image is ready, you must [deploy](#deploy) your function.


