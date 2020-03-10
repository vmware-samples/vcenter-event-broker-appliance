# vCenter Event Broker Appliance - Troubleshooting

## Requirements

You must log on to the VEBA appliance as root. You can do this from the console. If you want SSH access, do the following:

**Step 1** - After logging into the console as root, edit the sshd_config file

```
vi /etc/ssh/sshd_config
```

**Step 2** - Change the PermitRootLogin configuration line to `yes`. In vi, you can use the arrow keys to navigate up and down, delete characters with the `x` key, switch to insert mode with the `i` key, and then type your replacement characters. 

```
PermitRootLogin yes
```

**Step 3** - Save the configuration file. In vi, you hit the `Escape` key to get into command mode, type a colon `:` character, then `wq` for "write quit".

```
{Escape} :wq
````

**Step 4** - Restart the SSHD service

```
systemctl restart sshd
```

At this point you should be able to use your favorite SSH client to SSH to the VEBA host.

<BR>

## Troubleshooting an initial deployment

If VEBA is not working immediately after deployment, the first thing to do is check your Kubernetes pods. 

```
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

One of the first things to look for is whether a pod is in a crash state. In this case, the vmware-event-router pod is crashing. We need to look at the logs with this command:

```
kubectl logs vmware-event-router-5dd9c8f858-5c9mh  -n vmware
```

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

```
vi /root/event-router-config.json

```

Here is some of the JSON from the config file - you can see the mistake in the credentials. Fix the credentials and save the file.
```
[{
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
        },

```

We now fix the Kubernetes configuration with 3 commands - delete and recreate the secret file, then delete the broken pod. Kubernetes will automatically spin up a new pod with the new configuration

```console
kubectl --kubeconfig /root/.kube/config -n vmware delete secret event-router-config
kubectl -kubeconfig /root/.kube/config -n vmware create secret generic event-router-config –-from-file=event-router-config.json
kubectl -kubeconfig /root/.kube/config -n vmware delete pod vmware-event-router-5dd9c8f858-5c9mh 
```


We get a pod list again to determine the name of the new pod
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

```
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