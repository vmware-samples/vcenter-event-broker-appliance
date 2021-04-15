# Auto-refresh of a vSphere Client UI plugin

> **Note:** This function is not secure and is only for development and experimental purposes. It should not be used in production systems. 

## Description

Plugin auto-refresh function triggers automatic refresh of the UI of a vSphere Client plugin after data changes. This flow is demonstrated using the vSphere Client SDK remote-plugin-sample - the chassis' list of hosts is automatically refreshed when updated remotely.
This function does not use an OpenFaaS template. Instead, the OpenFaaS handler is implemented right into the plugin code as a VebaController which performs the plugin auto-refresh. Each time a given event occurs an endpoint inside the VebaController handler is called from the upstream URL in the dockerfile of the function image. In turn the handler sends an "update" message from the backend to the frontend of the plugin.

## Get the example function

Clone this repository which contains the example functions. 

```bash
git clone https://github.com/vmware-samples/vcenter-event-broker-appliance
cd vcenter-event-broker-appliance/examples/java/plugin-auto-refresh
git checkout master
```

## Customize the function

Define the function-specific settings. They are performed in the `plugin-auto-refresh.yml` file. Open and edit the `plugin-auto-refresh.yml` provided within the Python plugin auto-refresh example code. Add the environmental variables `VCENTER_FQDN`, `VCENTER_THUMBPRINT`, `VCENTER_PORT` and `VCENTER_GUID`. Change `gateway` and (optionally) `topic` as per your environment/needs.

```yaml
provider:
  name: openfaas
  gateway: https://VEBA_FQDN_OR_IP/ # replace with your vCenter Event Broker Appliance environment
functions:
  plugin-auto-refresh:
    image: vmware/plugin-auto-refresh:latest
    environment:
      write_debug: true
      read_debug: true
      VCENTER_FQDN: YOUR_VCENTER_FQDN
      VCENTER_THUMBPRINT: YOUR_VCENTER_THUMBPRINT
      VCENTER_PORT: YOUR_VCENTER_PORT
      VCENTER_GUID: YOUR_VCENTER_GUID
    annotations:
      topic: "HostConnectedEvent,HostDisconnectedEvent" # the event can be changed depending on the plugin use cases
```

## Set up the function and the plugin

### Deploy the function

After you've performed the steps and modifications above, you can go ahead and deploy the function:

```bash
faas-cli deploy -f plugin-auto-refresh.yml --tls-no-verify
Deployed. 202 Accepted.
```

### Port-forwarding
Then you should forward the port that the plugin uses on your VEBA, so that you can connect to the plugin from outside. Log in to your VEBA machine and execute the following command:

```bash
kubectl port-forward --address 0.0.0.0 pod/<pod_name> 8089:8089 #The plugin runs on port 8089 by default. If you have changed the port, edit it in the command here.
```

### Allow http connections in your vSphere Client
Ssh to your vCenter:

```bash
ssh root@YOUR_VC_IP
```

Then go to etc/vmware/vsphere-ui and add "allowHttp=true" to the end of the webclient.properties file.

After that restart the client:

```bash
service-control --restart vsphere-ui
```

### Plugin registration
Lastly, register the plugin to your vCenter followingin these steps:
   1. Go to https://YOUR_VCENTER_IP/mob/?moid=ExtensionManager
   2. Click `RegisterExtension`
   3. Paste the extension after replacing PLUGIN_SERVER_IP with your VEBA IP and SERVER_THUMBPRINT with your plugin server thumbprint (more information in the [official SDK documentation](https://code.vmware.com/docs/12634/developing-remote-plug-ins-with-the-vsphere-client-sdk--7-0-update-1-/GUID-1FDC38D2-B10B-4DF7-A2B0-5D70B3A1E287.html)):

   ```xml
   <extension>
   <description>
      <label>Remote Plug-in</label>
      <summary>This is a sample plug-in</summary>
   </description>
   <key>com.vmware.sample.remote</key>
   <company>VMware</company>
   <version>1.0.0</version>
   <server>
      <url>http://PLUGIN_SERVER_IP:8089/sample-ui/plugin.json</url>
      <description>
      <label>Remote Plug-in</label>
      <summary>This is a sample plug-in</summary>
      </description>
      <company>VMware</company>
      <type>com.vmware.sample.remote</type>
      <adminEmail>noreply@vmware.com</adminEmail>
      <serverThumbprint>SERVER_THUMBPRINT</serverThumbprint>
   </server>
   <client>
      <version>1.0.0</version>
      <description>
         <label>Remote Plug-in</label>
         <summary>This is a sample plug-in</summary>
      </description>
      <company>VMware</company>
      <type>vsphere-client-remote</type>
      <url>http://PLUGIN_SERVER_IP:8089/sample-ui/plugin.json</url>
   </client>
   <lastHeartbeatTime>1970-01-01T00:00:00Z</lastHeartbeatTime>
   </extension>
   ```
   
   4. Click Invoke


## Trigger the function

Open the vSphere Client on the chassis' related hosts list page of the plugin and add a host. Whenever the host is disconnected or connected again (e.g. from another browser page) the plugin host list UI will automatically refresh with the latest list of related hosts.

## Set up your own plugin auto-refresh

You can embed event-driven capabilities within a UI plugin of your own. To get an idea and replicate the process for your plugin follow the steps below explaining how it is done for the remote-plugin-sample of the vSphere Client SDK.

### Download the SDK plugin sample
Firstly, you need to [download](https://code.vmware.com/web/sdk/7.0/client) the html-client sdk, version 7.0.1.0. 

### Add Dockerfile and plugin-auto-refresh.yml file
Copy the Dockerfile and the plugin-auto-refresh file from the function repo and paste them to  `html-client-sdk/samples/remote-plugin-sample/`.

### Set the plugin port
This step is optional. 
If you want to change the port that the plugin uses, you must edit it in the server.port variable. It is located in the application.properties file, which can be found at `html-client-sdk/samples/remote-plugin-sample/target/classes`. Also, the endpoint in the upstream_url environmental variable must be updated with the new port in the Dockerfile.

### Allow http connection
> **Note:** This function is for testing purposes. Allowing http connection is not recommended for production as it might cause security concerns.

Go to the application.properties file again (`html-client-sdk/samples/remote-plugin-sample/target/classes`) and delete all properties with the pattern `server.ssl.*`

### Get the needed environmental variables
Go to `html-client-sdk/samples/remote-plugin-sample/src/main/java/com/vmware/sample/remote/configuration/Configuration.java` and replace the relevant functions with the code shown below:

```java
public String getVcenterServerFqdn() {
      return System.getenv().get("VCENTER_FQDN");
   }
```
```java
public String getVcenterGuid() {
      return System.getenv().get("VCENTER_GUID");
   }
```
```java
public String getVcenterSslThumbprint() {
      return replaceThumbprintDelimiters(System.getenv().get("VCENTER_THUMBPRINT"));
   }
```
```java
public String getVcenterServerPort() {
      return System.getenv().get("VCENTER_PORT");
   }
```

### Add the VebaController class
Copy the VebaController.java file from the function repo and paste it to 
`html-client-sdk/samples/remote-plugin-sample/src/main/java/com/vmware/sample/remote/controllers`.
Also, add the following code snippet in the spring-context.xml at `html-client-sdk/samples/remote-plugin-sample/target/classes`.

```xml
<bean name="vebaController"
         class="com.vmware.sample.remote.controllers.VebaController">
      <constructor-arg ref="messagingService"/>
   </bean>
```

### Add new message type "HOSTS_LIST_UPDATED"
As the VebaController uses a new type of message, called "HOSTS_LIST_UPDATED", this type should be added to the MessageType.java file at `html-client-sdk/samples/remote-plugin-sample/src/main/java/com/vmware/sample/remote/model`. After the edit your MessageType class should look like that:

```java
public enum MessageType {
   CHASSIS_UPDATED,
   HOSTS_LIST_UPDATED
}
```

### Edit the retrieving of the hosts, which is invoked when auto-refresh occurs.
As we do not want to lose a host from the list after it has been disconneted and then connected back, we must change the `retrieveConnectedHosts` method in HostController.java file, located at `html-client-sdk/samples/remote-plugin-sample/src/main/java/com/vmware/sample/remote/controllers`. The needed code for the method can be found below:

```java
@RequestMapping(value = "/chassis/{chassisId}/hosts", method = RequestMethod.GET)
   public List<Host> retrieveConnectedHosts(
         @PathVariable("chassisId") final String chassisId) {


      final List<String> relatedHostsList = chassisService.getChassisById(chassisId).relatedHostsIds;

      final List<Host> hostsList = vcenterInfoService.retrieveConnectedHosts();

      Map<String, Host> hostsById= hostsList.stream().collect(
              Collectors.toMap(host -> host.id, host -> host));

      List<Host> result = new ArrayList<>();
      for(String id : relatedHostsList){
          if(hostsById.containsKey(id)){
             result.add(hostsById.get(id));
          }
      }

      return result;
   }
```

### Build and push the image
After all these changes in the code of the plugin, we are ready to proceed to install, build and push of our auto-refresh function.
To build the image, open terminal and go to `html-client-sdk/samples/remote-plugin-sample/` where the Dockerfile is located. Then use the following commands:

```bash
mvn clean install # To build the function
docker build -t <repo>/<name>:<version> . # To build an image of the function
docker push <repo>/name>:<version> # To push the image to your docker repository
```

Then the image is ready and you can [deploy](#deploy-the-function) your function (don't forget to edit the repository/image_name in the plugin-auto-refresh.yml file before doing this).

When this is done the plugin is ready to handle event-driven updates of the UI.
