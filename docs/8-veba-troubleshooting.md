# vCenter Event Broker Appliance - Troubleshooting

# Table of Contents
<!-- TOC depthFrom:2 -->

- [Requirements](#requirements)
- [Troubleshooting an initial deployment](#troubleshooting-an-initial-deployment)
- [OpenFaaS Function troubleshooting](#openfaas-function-troubleshooting)

<!-- /TOC -->

## Requirements

You must log on to the vCenter Event Broker appliance as root. You can do this from the console. If you want SSH access, execute the following command:

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
vi /root/config/event-router-config.json

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
  
## OpenFaaS Function troubleshooting

If a function is not behaving as expected, you can look at the logs to troubleshoot. First, SSH or console to the appliance as shown in the Requirements section.

List out the pods.

```bash
kubectl get pods -A
```

This is the function output:

```
NAMESPACE        NAME                                   READY   STATUS      RESTARTS   AGE
kube-system      coredns-584795fc57-4bp2s               1/1     Running     1          6d4h
kube-system      coredns-584795fc57-76pwr               1/1     Running     1          6d4h
kube-system      etcd-veba01                            1/1     Running     2          6d4h
kube-system      kube-apiserver-veba01                  1/1     Running     2          6d4h
kube-system      kube-controller-manager-veba01         1/1     Running     3          6d4h
kube-system      kube-proxy-fvf2n                       1/1     Running     2          6d4h
kube-system      kube-scheduler-veba01                  1/1     Running     2          6d4h
kube-system      weave-net-v9jss                        2/2     Running     6          6d4h
openfaas-fn      powercli-entermaint-d84fd8d85-sjdgl    1/1     Running     1          6d4h
openfaas         alertmanager-58f8d787d9-nqwm8          1/1     Running     1          6d4h
openfaas         basic-auth-plugin-dd49cd66b-rv6n7      1/1     Running     1          6d4h
openfaas         faas-idler-59ff9778fd-84szz            1/1     Running     4          6d4h
openfaas         gateway-74f6f9489b-btgz8               2/2     Running     5          6d4h
openfaas         nats-6dfbf45d77-9swph                  1/1     Running     1          6d4h
openfaas         prometheus-5f5494b54f-srs2d            1/1     Running     1          6d4h
openfaas         queue-worker-59b67bf4-wqhm5            1/1     Running     4          6d4h
projectcontour   contour-5cddfc8f6-hpzn8                1/1     Running     1          6d4h
projectcontour   contour-5cddfc8f6-tdv2r                1/1     Running     2          6d4h
projectcontour   contour-certgen-wrgnb                  0/1     Completed   0          6d4h
projectcontour   envoy-8mdhb                            1/1     Running     2          6d4h
vmware           tinywww-7fcfc6fb94-v7ncj               1/1     Running     1          6d4h
vmware           vmware-event-router-5dd9c8f858-9g44h   1/1     Running     4          6d4h
```

First, we want to see if the event router is capturing events and forwarding them on to a function. 

Use this command to follow the live Event Router log.

```bash
kubectl logs -n vmware vmware-event-router-5dd9c8f858-9g44h --follow
```

For this sample troubleshooting, we have the sample hostmaintenance alarms function running. To see if the appliance is properly handling the event, we put a host into maintenance mode. 

When we look at the log output, we see various entries regarding EnteredMaintenanceModeEvent, ending with the following:

```
[OpenFaaS] 2020/03/11 22:15:09 invoking function(s) on topic: EnteredMaintenanceModeEvent
[OpenFaaS] 2020/03/11 22:15:09 successfully invoked function powercli-entermaint for topic EnteredMaintenanceModeEvent
```

This lets us know that the function was invoked. If we still don't see the expected result, we need to look at the function logs.

Each OpenFaaS function will have its own pod running in the openfaas-fn namespace. We can examine the logs with the following command.

```bash
kubectl logs -n openfaas-fn powercli-entermaint-d84fd8d85-sjdgl
```

We don't need the --follow switch because we are just trying to look at recent logs, but --follow would work too.  
Some other useful switches are `--since` and `--tail`. 

This command will show you the last 5 minutes worth of logs.

```bash
kubectl logs -n openfaas-fn powercli-entermaint-d84fd8d85-sjdgl --since=5m
```

This command will show you the last 20 lines of logs.

```bash
kubectl logs -n openfaas-fn powercli-entermaint-d84fd8d85-sjdgl --tail=20
```

Log output showing a succesful function invocation: 

```
Connecting to vCenter Server ...

Disabling alarm actions on host: esx01.labad.int
Disconnecting from vCenter Server ...

2020/03/11 22:15:15 Duration: 6.085448 seconds
```

An alternative way to troubleshoot OpenFaaS logs is to use `faas-cli`.

This faas-cli command will show all available functions in the appliance. ```--tls-no-verify``` bypasses SSL certificate validation

```bash
faas-cli list --tls-no-verify
```

The command output is:

```
Function                        Invocations     Replicas
powercli-entermaint             3               1
```

We can  look at the logs with this command.

```bash
faas-cli logs powercli-entermaint --tls-no-verify
```

The logs are the same: 

```
2020-03-11T22:15:15Z Connecting to vCenter Server ...
2020-03-11T22:15:15Z
2020-03-11T22:15:15Z Disabling alarm actions on host: esx01.labad.int
2020-03-11T22:15:15Z Disconnecting from vCenter Server ...
2020-03-11T22:15:15Z
2020-03-11T22:15:15Z 2020/03/11 22:15:15 Duration: 6.085448 seconds
```

All of the same switches shown in the kubectl commands such as `--tail` and `--since` work with `faas-cli`.

> **Note:** `faas-cli` will stop tailing the log after a fixed period.