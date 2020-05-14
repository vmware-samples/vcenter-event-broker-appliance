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

You must log on to the VMware Event Broker appliance as root. You can do this from the console. If you want SSH access, execute the following command:

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
NAMESPACE        NAME                                   READY   STATUS          RESTARTS        AGE
kube-system      coredns-584795fc57-hcvxh               1/1     Running              1          4d15h
kube-system      coredns-584795fc57-hf72w               1/1     Running              1          4d15h
kube-system      etcd-veba02                            1/1     Running              1          4d15h
kube-system      kube-apiserver-veba02                  1/1     Running              1          4d15h
kube-system      kube-controller-manager-veba02         1/1     Running              1          4d15h
kube-system      kube-proxy-fj47p                       1/1     Running              1          4d15h
kube-system      kube-scheduler-veba02                  1/1     Running              1          4d15h
kube-system      weave-net-vs8ls                        2/2     Running              4          4d15h
projectcontour   contour-5cddfc8f6-8hzd6                1/1     Running              1          4d15h
projectcontour   contour-5cddfc8f6-jq7d8                1/1     Running              1          4d15h
projectcontour   contour-certgen-f92l5                  0/1     Completed            0          4d15h
projectcontour   envoy-gcmqt                            1/1     Running              1          4d15h
vmware           tinywww-7fcfc6fb94-mfltm               1/1     Running              1          4d15h
vmware           vmware-event-router-5dd9c8f858-5c9mh   0/1     CrashLoopBackoff     6          4d13h
```

> **Note:** The status ```Completed``` of the container ```contour-certgen-f92l5``` is expected after successful appliance deployment.

One of the first things to look for is whether a pod is in a crash state. In this case, the vmware-event-router pod is crashing. We need to look at the logs with this command:

```bash
kubectl logs vmware-event-router-5dd9c8f858-5c9mh  -n vmware
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
vi /root/event-router-config.json

```

Here is some of the JSON from the config file - you can see the mistake in the credentials. Fix the credentials and save the file.

```json
[
  {
    "type": "stream",
    "provider": "vmware_vcenter",
    "address": "https://vc01.labad.int/sdk",
    "auth": {
      "method": "user_password",
      "secret": {
        "username": "administrator@vsphere.local",
        "password": "WrongPassword"
      }
    },
    "options": {
      "insecure": "true"
    }
  }
]
```

We now fix the Kubernetes configuration with 3 commands - delete and recreate the secret file, then delete the broken pod. Kubernetes will automatically spin up a new pod with the new configuration. We need to do this because the JSON configuration file is not directly referenced by the event router. The JSON file is mounted into the event router pod as a Kubernetes secret. 

```
kubectl -n vmware delete secret event-router-config
kubectl -n vmware create secret generic event-router-config --from-file=event-router-config.json
kubectl -n vmware delete pod vmware-event-router-5dd9c8f858-5c9mh 
```

We get a pod list again to determine the name of the new pod.

```
kubectl get pods -A
```

Here is the command output:
```
NAMESPACE        NAME                                   READY   STATUS        RESTARTS   AGE
kube-system      coredns-584795fc57-hcvxh               1/1     Running       1          4d19h
kube-system      coredns-584795fc57-hf72w               1/1     Running       1          4d19h
kube-system      etcd-veba02                            1/1     Running       1          4d19h
kube-system      kube-apiserver-veba02                  1/1     Running       1          4d19h
kube-system      kube-controller-manager-veba02         1/1     Running       1          4d19h
kube-system      kube-proxy-fj47p                       1/1     Running       1          4d19h
kube-system      kube-scheduler-veba02                  1/1     Running       1          4d19h
kube-system      weave-net-vs8ls                        2/2     Running       4          4d19h
projectcontour   contour-5cddfc8f6-8hzd6                1/1     Running       1          4d19h
projectcontour   contour-5cddfc8f6-jq7d8                1/1     Running       1          4d19h
projectcontour   contour-certgen-f92l5                  0/1     Completed     0          4d19h
projectcontour   envoy-gcmqt                            1/1     Running       1          4d19h
vmware           tinywww-7fcfc6fb94-mfltm               1/1     Running       1          4d19h
vmware           vmware-event-router-5dd9c8f858-5c9mh   0/1     Terminating   40         3h9m
vmware           vmware-event-router-5dd9c8f858-wt64s   1/1     Running       0          28s
```

Now view the event router logs.

```bash
kubectl logs -n vmware vmware-event-router-5dd9c8f858-wt64s
```

Here is the command output:
```

 _    ____  ___                            ______                 __     ____              __
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/


[VMware Event Router] 2020/03/10 20:37:28 connecting to vCenter https://vc01.labad.int/sdk/sdk
[VMware Event Router] 2020/03/10 20:37:28 connecting to OpenFaaS gateway http://gateway.openfaas:8080 (async mode: false)
[VMware Event Router] 2020/03/10 20:37:28 exposing metrics server on 0.0.0.0:8080 (auth: basic_auth)
[Metrics Server] 2020/03/10 20:37:28 starting metrics server and listening on "http://0.0.0.0:8080/stats"
2020/03/10 20:37:28 Syncing topic map
[OpenFaaS] 2020/03/10 20:37:28 processing event [0] of type *types.UserLoginSessionEvent from source https://vc01.labad.int/sdk: &{SessionEvent:{Event:{DynamicData:{} Key:8755384 ChainId:8755384 CreatedTime:2020-03-10 20:36:19.594 +0000 UTC UserName::<nil> ComputeResource:<nil> Host:<nil> Vm:<nil> Ds:<nil> Net:<nil> Dvs:<nil> FullFormattedMessage:User @10.46.144.4 logged in as VMware vim-java 1.0 ChangeTag:}} IpAddress:192.168.10.24 UserAgent:VMware vim-java 1.0 Locale:en SessionId:5254c9e5-4c2d-0af0-cae3-7fdebdc2eacb}
```

We now see that the Event Router came online, connected to vCenter, and successfully received an event.

## Changing the vCenter service account

If you need to change the account the appliance uses to connect to vCenter, use the following procedure. 

Open a console to the appliance. If you want to do the configuration via SSH, you must first enable the SSH daemon with the following command
```bash 
systemcl start sshd
```
The SSH daemon will run but not automatically start with the next reboot. You can use the same command with `stop` instead of `start` when you are finished. Or you can type everything directly into the console if you do not want to use SSH.

Edit the configuration file with vi
```bash
vi /root/config/event-router-config.json
```

The editor will open with output similar to this (truncated)
```bash
[{
                "type": "stream",
                "provider": "vmware_vcenter",
                "address": "https://vc01.lab.int/sdk",
                "auth": {
                        "method": "user_password",
                        "secret": {
                                "username": "administrator@vsphere.local",
                                "password": "KeepMeSecure123!"
                        }
                },
                "options": {
                        "insecure": "true"
                }
        },
```

Change the username and password, then save the file. Then delete and recreate the event router pod secret with the following commands:
```bash
kubectl -n vmware delete secret event-router-config
kubectl -n vmware create secret generic event-router-config --from-file=/root/event-router-config.json
```

Now, restart the event router pod. Get the current pod name with the following command:
```bash
kubectl get pods -A
```
You will see output similar the following (trucnated):
```bash
projectcontour   contour-certgen-7r9dl                  0/1     Completed   0          22d
projectcontour   envoy-htrwv                            1/1     Running     1          22d
vmware           tinywww-7fcfc6fb94-tv98j               1/1     Running     1          22d
vmware           vmware-event-router-5dd9c8f858-7htv5   1/1     Running     14         19d
```

Find the event router pod. Every environment will have a unique suffix on the pod name - in this example, it is `-5dd9c8f858-7htv5`. Delete the event router pod with the following command (make sure to match the pod name with the one in your environment):
```bash
kubectl -n vmware delete pod vmware-event-router-5dd9c8f858-7htv5
```

The pod will automatically recreate itself. You can repeatedly run the following command:
```bash
kubectl get pods -A
```
Various stages of the pod lifecycle may be shown. Here, we see the original pod terminating while the new pod is spinning up.
```bash
projectcontour   envoy-htrwv                            1/1     Running             1          22d
vmware           tinywww-7fcfc6fb94-tv98j               1/1     Running             1          22d
vmware           vmware-event-router-5dd9c8f858-7htv5   0/1     Terminating         14         19d
vmware           vmware-event-router-5dd9c8f858-l6gdj   0/1     ContainerCreating   0          4s
```

Eventually the old pod will disappear and the new pod will show as running:
```bash
projectcontour   contour-certgen-7r9dl                  0/1     Completed   0          22d
projectcontour   envoy-htrwv                            1/1     Running     1          22d
vmware           tinywww-7fcfc6fb94-tv98j               1/1     Running     1          22d
vmware           vmware-event-router-5dd9c8f858-l6gdj   1/1     Running     0          92s
```

You can check the pod logs with the following command (make sure to add the correct suffix shown in your environment):
```bash
kubectl logs -n vmware kubectl logs -n vmware  vmware-event-router-5dd9c8f858-n9pg6
```
You should see a successful connection to vCenter in the logs
```bash
 _    ____  ___                            ______                 __     ____              __
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/


[VMware Event Router] 2020/04/08 05:07:11 connecting to vCenter https://vc01.lab.int/sdk
[VMware Event Router] 2020/04/08 05:07:11 connecting to OpenFaaS gateway http://gateway.openfaas:8080 (async mode: false)
[VMware Event Router] 2020/04/08 05:07:11 exposing metrics server on 0.0.0.0:8080 (auth: basic_auth)
[Metrics Server] 2020/04/08 05:07:11 starting metrics server and listening on "http://0.0.0.0:8080/stats"
```