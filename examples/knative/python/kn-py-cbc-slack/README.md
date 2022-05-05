# kn-py-cbc

A pair of example Python functions one of which triggers when a VM is created and the other when VM is deleted. When a VM is created the Carbon Black Cloud API is contacted and sensor enablement initiated. When a VM is deleted the Carbon Black Cloud API is contacted and the stale VM resource record is removed.

The use cases for these are slightly different so we have included functions to deply them seperately or together.  The function-deploy.yaml would be used in situations where new virtual machines are being created using powercli for example and Carbon Black is not pre installed, in this scenario we are automating the deployment of a Carbon Black sensor whenever a new VM is deployed.  The function-remove.yaml might be used on its own in a VDI environment where the clones are pre-installed with Carbon Black but they are deleted whenever logged off without informing the Carbon Black Cloud Console that this device no longer exisits.  In this scenario we are house keeping the console and removing sensors as they are removed from vCenter.  The function-both.yaml can be used if you have both use cases, the secrets are shared between the deployments so you only need to create them once.

If you are making changes to the handler.py logic or want to host your own images in a private registry then you will need to recreate the images as below in steps 1 to 4.  If you are happy deploying the latest public image skip straight to Step 4.

## Step 1 - Build image

[Buildpacks](https://buildpacks.io) are used to create the container images.

```bash
cd ~/kn-py-cbc/deploy
pack build -B gcr.io/buildpacks/builder:v1 <container-registry>/kn-py-cbc-deploy:1.0
cd ~/kn-py-cbc/remove
pack build -B gcr.io/buildpacks/builder:v1 <container-registry>/kn-py-cbc-remove:1.0
```

## Step 2 - Test image

Verify the container image works by simulating creation of secret and other environment variables and start image in interactive tty mode.

```bash
docker secret create CBC_CONFIG_INI ./cbc-ini
docker run -e PORT=8080 -it --env-file ./cbc-envs --rm -p 8080:8080 <container-registry>/kn-py-cbc-deploy:1.0
```

As this is running in interactive mode, you should see Stdout displayed of Flask web framework starting in debug mode serving the handler.py application:

```bash
* Serving Flask app "handler.py" (lazy loading)
 * Environment: development
 * Debug mode: on
 * Running on all addresses.
   WARNING: This is a development server. Do not use it in a production deployment.
 * Running on http://172.17.0.2:8080/ (Press CTRL+C to quit)
 * Restarting with stat
 * Debugger is active!
 * Debugger PIN: 994-125-687
 ```

In a separate terminal window, go to the test directory and edit the `deploy.json` file to to include a VM name registered within CBC but without sensor enabled,  or `remove.json` to include a VM name removed from vcenter but still registered within CBC.

```json
"Vm": 
    {"Name": "new-vm",
```

Simulate CloudEvent being posted to the function using cURL to make HTTP POST of the appropriate test json.

```console
cd test
curl -i -d@deploy.json localhost:8080
```

If you check back to the interactive Stdout from container image you should see the POST being recieved and output from the function:

```console
* Serving Flask app "handler.py" (lazy loading)
 * Environment: development
 * Debug mode: on
 * Running on all addresses.
   WARNING: This is a development server. Do not use it in a production deployment.
 * Running on http://172.17.0.2:8080/ (Press CTRL+C to quit)
 * Restarting with stat
 * Debugger is active!
 * Debugger PIN: 994-125-687
2021-05-26 18:56:27,719 INFO handler Thread-3 : "***cloud event*** {"attributes": {"specversion": "1.0", "id": "08179137-b8e0-4973-b05f-8f212bf5003b", "source": "https://10.0.0.1:443/sdk", "type": "com.vmware.event.router/event", "datacontenttype": "application/json", "subject": "VmPoweredOffEvent", "time": "2020-02-11T21:29:54.9052539Z"}, "data": {"Key": 9902, "ChainId": 9895, "CreatedTime": "2020-02-11T21:28:23.677595Z", "UserName": "VSPHERE.LOCAL\\Administrator", "Datacenter": {"Name": "testDC", "Datacenter": {"Type": "Datacenter", "Value": "datacenter-2"}}, "ComputeResource": {"Name": "cls", "ComputeResource": {"Type": "ClusterComputeResource", "Value": "domain-c7"}}, "Host": {"Name": "10.185.22.74", "Host": {"Type": "HostSystem", "Value": "host-21"}}, "Vm": {"Name": "test-01", "Vm": {"Type": "VirtualMachine", "Value": "vm-56"}}, "Ds": null, "Net": null, "Dvs": null, "FullFormattedMessage": "test-01 on  10.0.0.1 in testDC is powered off", "ChangeTag": "", "Template": false}}
172.17.0.1 - - [26/May/2021 18:56:27] "POST / HTTP/1.1" 204 -
2021-05-26 18:56:27,720 INFO werkzeug Thread-3 : 172.17.0.1 - - [26/May/2021 18:56:27] "POST / HTTP/1.1" 204 -
```

Check the Carbon Black Cloud UI and ensure the sensor installation has been started or the stale record removed.

Once you are happy the function behaves as expected push the container image to container registry.

```console
docker push <container-registry>/kn-py-cbc-deploy:1.0
```

Repeat test by starting the remove container image and using the remove.json test event.

## Step 3 - Deploy function(s)

Ensure the kubernetes manifest file `function.yaml` reflects the correct container registry, images and image versions. Adjust values of secret files to reflect values appropriate to your environment. With these complete create the Kubernetes secrets and Knative triggers and services.

```console
# deploy secret
kubectl create secret generic cbc-env --from-file=CBC_CONFIG_INI=./cbc-ini --from-file=CBC_URL=./cbc-url --from-file=CBC_ORG_KEY=./cbc-org-key --from-file=CBC_TOKEN=./cbc-token --from-file=SLACK_URL=./slack-url --from-file=SENSOR_VER=./sensor-ver -n vmware-functions
# To deploy the function to install CB on new VM's
kubectl -n vmware-functions apply -f function-deploy.yaml
# To deploy the function to remove old vdi clones from the CBC console
kubectl -n vmware-functions apply -f function-remove.yaml
# To deploy both functions together
kubectl -n vmware-functions apply -f function-both.yaml
```

## Step 4 - Undeploy

```console
# undeploy function
kubectl -n vmware-functions delete -f function.yaml
kubectl -n vmware-functions delete secret cbc-env
```
