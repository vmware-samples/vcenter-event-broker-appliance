---
layout: docs
toc_id: troubleshoot-appliance
title: VMware Event Broker Appliance Troubleshooting
description: Troubleshooting guide for general appliance issues
permalink: /kb/troubleshoot-appliance
cta:
 title: Still having trouble?
 description: Please submit bug reports and feature requests by using our GitHub [Issues](https://github.com/vmware-samples/vcenter-event-broker-appliance/issues){:target="_blank"} page or Join us on slack [#vcenter-event-broker-appliance](https://vmwarecode.slack.com/archives/CQLT9B5AA){:target="_blank"} on vmwarecode.slack.com.
---
# VMware Event Broker Appliance - Troubleshooting

## Requirements

You must log on to the VMware Event Broker appliance as root. If you did not enable SSH as part of the initial VMware Event Broker Appliance deployment, you can perform this operation at the console. To enable SSH access, execute the following command:

```bash
systemctl start sshd
```

This turns on the SSH daemon but does not enable it to start on appliance boot. You should now be able to SSH into the appliance.

If you wish to disable the SSH daemon when you are done troubleshooting, execute the following command:

```bash
systemctl stop sshd
```
## Troubleshooting an initial deployment

If the appliance is not working immediately after deployment, the first thing to do is check your Kubernetes pods.

```bash
kubectl get pods -A
```

Here is the command output:

```
k get pods -A
NAMESPACE            NAME                                                  READY   STATUS      RESTARTS   AGE
contour-external     contour-5869594b-bmccv                                1/1     Running     0          6h37m
contour-external     contour-5869594b-jr4k8                                1/1     Running     0          6h37m
contour-external     contour-certgen-v1.10.0-btmlp                         0/1     Completed   0          6h37m
contour-external     envoy-hzhnz                                           2/2     Running     0          6h37m
contour-internal     contour-5d47766fd8-5shp2                              1/1     Running     0          6h37m
contour-internal     contour-5d47766fd8-hv9zl                              1/1     Running     0          6h37m
contour-internal     contour-certgen-v1.10.0-szssj                         0/1     Completed   0          6h37m
contour-internal     envoy-hpch5                                           2/2     Running     0          6h37m
knative-eventing     eventing-controller-658f454d9d-pfs5d                  1/1     Running     0          6h37m
knative-eventing     eventing-webhook-69fdcdf8d4-fdtmp                     1/1     Running     0          6h37m
knative-eventing     rabbitmq-broker-controller-88fc96b44-6jb82            1/1     Running     0          6h37m
knative-serving      activator-85cd6f6f9-7l9q5                             1/1     Running     0          6h38m
knative-serving      autoscaler-7959969587-kzdrj                           1/1     Running     0          6h38m
knative-serving      contour-ingress-controller-6d5777577c-sb6vr           1/1     Running     0          6h37m
knative-serving      controller-577558f799-fmfgq                           1/1     Running     0          6h38m
knative-serving      webhook-78f446786-zj9n4                               1/1     Running     0          6h38m
kube-system          antrea-agent-xgn24                                    2/2     Running     0          6h38m
kube-system          antrea-controller-849fff8c5d-h5tjb                    1/1     Running     0          6h38m
kube-system          coredns-74ff55c5b-kpfxz                               1/1     Running     0          6h38m
kube-system          coredns-74ff55c5b-wrjp4                               1/1     Running     0          6h38m
kube-system          etcd-veba.primp-industries.local                      1/1     Running     0          6h38m
kube-system          kube-apiserver-veba.primp-industries.local            1/1     Running     0          6h38m
kube-system          kube-controller-manager-veba.primp-industries.local   1/1     Running     0          6h38m
kube-system          kube-proxy-gs59c                                      1/1     Running     0          6h38m
kube-system          kube-scheduler-veba.primp-industries.local            1/1     Running     0          6h38m
local-path-storage   local-path-provisioner-5696dbb894-sf457               1/1     Running     0          6h38m
rabbitmq-system      rabbitmq-cluster-operator-7bbbb8d559-sw4kt            1/1     Running     0          6h37m
vmware-functions     default-broker-ingress-5c98bf68bc-nl5m7               1/1     Running     0          6h36m
vmware-system        tinywww-dd88dc7db-f74br                               1/1     Running     0          6h36m
vmware-system        veba-rabbit-server-0                                  1/1     Running     0          6h36m
vmware-system        veba-ui-677b77dfcf-q9t84                              1/1     Running     0          6h36m
vmware-system        vmware-event-router-5dd9c8f858-5c9mh                  0/1     CrashLoopBackoff     6  4d13h
```

> **Note:** The status ```Completed``` of the container ```contour-certgen-v1.10.0-btmlp``` is expected after successful appliance deployment.

One of the first things to look for is whether a pod is in a crash state.

### Recovering from a crashing event router pod
In the above case, the vmware-event-router pod is crashing. We need to look at the logs with this command:

```bash
kubectl -n vmware-system logs vmware-event-router-5dd9c8f858-5c9mh
```

> **Note:** The pod suffix ```-5dd9c8f858-5c9mh``` will be different in each environment

Here is the command output:

```
 _    ____  ___                            ______                 __     ____              __
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/


[VMware Event Router] 2020/03/10 18:59:47 connecting to vCenter https://vc01.labad.int/sdk
[VMware Event Router] 2020/03/10 18:59:52 could not connect to vCenter: could not create vCenter client: ServerFaultCode: Cannot complete login due to an incorrect user name or password.
```

The error message shows us that we made a mistake when we configured our username or password. We must now edit the Event Router JSON configuration file to fix the mistake.

```bash
vi /root/config/event-router/vmware-event-router-config-vcenter.yaml
```

Here is some of the YAML from the config file - you can see the mistake in the credentials. Fix the credentials and save the file.

```yaml
eventProvider:
  name: veba-vc-01
  type: vcenter
  vcenter:
    address: https://vc01.labad.int/sdk
    auth:
      basicAuth:
        password: "wrongPassword"
        username: "veba@vsphere.local"
      type: basic_auth
    insecureSSL: true
    checkpoint: false
```

We now fix the Kubernetes configuration with 3 commands - delete and recreate the secret file, then delete the broken pod. Kubernetes will automatically spin up a new pod with the new configuration. We need to do this because the YAML configuration file is not directly referenced by the event router. The YAML file is mounted into the event router pod as a Kubernetes secret.

```
kubectl -n vmware-system delete secret event-router-config
kubectl -n vmware-system create secret generic event-router-config --from-file=event-router-config.yaml
kubectl -n vmware-system delete pod vmware-event-router-5dd9c8f858-5c9mh
```

We get a pod list again to determine the name of the new pod.

```
kubectl -n vmware-system get pods
```

Here is the command output:
```
NAME                                  READY   STATUS        RESTARTS   AGE
tinywww-dd88dc7db-f74br               1/1     Running       0          6h39m
veba-rabbit-server-0                  1/1     Running       0          6h40m
veba-ui-677b77dfcf-q9t84              1/1     Running       0          6h39m
vmware-event-router-5dd9c8f858-5c9mh  0/1     Terminating   3          6h40m
vmware-event-router-5dd9c8f858-wt64s  1/1     Running       0          28s
```

Now view the event router logs.

```bash
kubectl -n vmware-system logs vmware-event-router-5dd9c8f858-wt64s
```

Here is the command output:
```
 _    ____  ___                            ______                 __     ____              __
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/

2021-04-04T14:30:16.589Z        ESC[34mINFOESC[0m       [MAIN]  router/main.go:111      connecting to vCenter   {"address": "https://vc01.labad.int/sdk"}
2021-04-04T14:30:16.589Z        ESC[34mINFOESC[0m       [KNATIVE]       injection/injection.go:61       Starting informers...
2021-04-04T14:30:16.699Z        ESC[34mINFOESC[0m       [MAIN]  router/main.go:149      created Knative processor       {"sink": "http://default-broker-ingress.vmware-functions.svc.cluster.local"}
2021-04-04T14:30:16.699Z        ESC[33mWARNESC[0m       [METRICS]       metrics/server.go:59    no credentials found, disabling authentication for metrics server
2021-04-04T14:30:16.700Z        ESC[34mINFOESC[0m       [METRICS]       metrics/server.go:131   starting metrics server {"address": "http://0.0.0.0:8082/stats"}
2021-04-04T14:30:16.703Z        ESC[34mINFOESC[0m       [VCENTER]       vcenter/vcenter.go:174  checkpointing disabled, setting begin of event stream   {"beginTimestamp": "2021-04-04 14:30:16.70642 +0000 UTC"}
```

We now see that the Event Router came online, connected to vCenter, and successfully received an event.

### Check for completed installation

If the pods appear to be up without a crash status, check to make sure the installation completed. The file `/root/ran_customzation` gets created when installation completes successfully. If this file is missing, you can turn to the installation logs to find out why.
```bash
root@veba [ ~ ]# ls -al /root/ran_customization
-rw-r--r-- 1 root root 0 Oct  1  2021 /root/ran_customization
```

### Examine log files

The appliance installation log file is found in `/var/log/bootstrap.log`. If enabled at install time, a debug log is available in `/var/log/bootstrap-debug.log`. The logs should point you toward the source of the issue. Don't hesitate to [reach out](#bottom) to the team if you need help.
## Changing the vCenter service account

If you need to change the account the appliance uses to connect to vCenter, the procedure above can be used.

## Troubleshooting VMware Event Broker Appliance vSphere UI

Ensure there is proper bi-directional network connectivity between the VMware Event Broker Appliance and vCenter Server for proper UI functionality which runs over port 443.

When the VMware Event Broker Appliance vSphere UI container starts up, it will attempt to register itself as a vSphere remote plugin with the vCenter Server. Upon a successful registration, a vCenter Plugin Extension will be created on the vCenter Server and that will direct the vSphere UI to connect to the VMware Event Broker Appliance, which is where the VMware Event Broker Appliance UI plugin will be running.

There are several areas which may prevent the VMware Event Broker Appliance vSphere UI from properly running.

Ensure that the VMware Event Broker Appliance UI container is running:

```bash
kubectl -n vmware-system get deployments/veba-ui
```

Here is the command output:


```console
kubectl -n vmware-system get deployments/veba-ui
NAME      READY   UP-TO-DATE   AVAILABLE   AGE
veba-ui   1/1     1            1           6h50m
```

Ensure there are no errors in the logs for the VMware Event Broker Appliance UI container. Incorrect credentials or credentials without the correct permissions will prevent the registration with vCenter Server and the logs should give you some additional insights.

```bash
kubectl -n vmware-system logs deployments/veba-ui
```

Here is the command output:

```console
  .   ____          _            __ _ _
 /\\ / ___'_ __ _ _(_)_ __  __ _ \ \ \ \
( ( )\___ | '_ | '_| | '_ \/ _` | \ \ \ \
 \\/  ___)| |_)| | | | | || (_| |  ) ) ) )
  '  |____| .__|_| |_|_| |_\__, | / / / /
 =========|_|==============|___/=/_/_/_/
 :: Spring Boot ::        (v2.0.3.RELEASE)

2021-04-04 14:30:22.690  INFO 1 --- [           main] c.v.sample.remote.SpringBootApplication  : Starting SpringBootApplication on veba-ui-677b77dfcf-q9t84 with PID 1 (/app.jar started by root in /)
2021-04-04 14:30:22.695  INFO 1 --- [           main] c.v.sample.remote.SpringBootApplication  : No active profile set, falling back to default profiles: default
2021-04-04 14:30:22.806  INFO 1 --- [           main] ConfigServletWebServerApplicationContext : Refreshing org.springframework.boot.web.servlet.context.AnnotationConfigServletWebServerApplicationContext@3339ad8e: startup date [Sun Apr 04 14:30:22 GMT 2021]; root of context hierarchy
2021-04-04 14:30:23.569  INFO 1 --- [           main] o.s.b.f.xml.XmlBeanDefinitionReader      : Loading XML bean definitions from class path resource [spring-context.xml]
...
```

Ensure the VMware Event Broker Appliance UI was able to successfully register with vCenter Server and create vCenter Plugin Extension.

Open a web browser and enter the following URL and login. Replace the FQDN or IP Address of your vCenter Server

```console
https://[VCENTER_FQDN_OR_IP]/mob/?moid=ExtensionManager&doPath=extensionList%5b%22com.vmware.veba%22%5d
```

If the registration exists, this most likely means that vCenter Server can not reach the VMware Event Broker Appliance.

To verify, you can SSH to the vCenter Server Appliance and attempt to retrieve the `plugin.json`.

```bash
curl -L -k https://[VEBA_FQDN_OR_IP_ADDRESS]/veba-ui/plugin.json
```

Here is the command output:

```console
{
   "manifestVersion": "1.0.0",
   "requirements": {
      "plugin.api.version": "1.0.0"
   },
   "configuration": {
      "nameKey": "plugin.name",
      "icon": {
         "name": "main"
      }
   },
   "global": {
      "view": {
         "navigationId": "entryPoint",
         "uri": "index.html#/veba-ui",
         "navigationVisible": false
      }
   },
   "definitions": {
      "iconSpriteSheet": {
         "uri": "assets/images/veba_otto_the_orca.png",
         "definitions": {
            "main": {
               "x": 0,
               "y": 0
            }
         },
         "themeOverrides": {
            "dark": {
               "uri": "assets/images/veba_otto_the_orca.png",
               "definitions": {
                  "main": {
                     "x": 0,
                     "y": 0
                  }
               }
            }
         }
      },
      "i18n": {
         "locales": [
            "en-US",
            "de-DE",
            "fr-FR"
         ],
         "definitions": {
            "plugin.name": {
               "en-US": "VMware Event Broker",
               "de-DE": "VMware Event Broker",
               "fr-FR": "VMware Event Broker"
            }
         }
      }
   }
}
```
<a name="bottom"></a>