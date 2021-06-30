#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Google cAdvisor

set -euo pipefail

# Retrieve the cAdvisor image
VEBA_BOM_FILE=/root/config/veba-bom.json
CADVISOR_IMAGE=$(jq -r < ${VEBA_BOM_FILE} '.["cadvisor"].containers[0].name')
CADVISOR_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["cadvisor"].containers[0].version')

echo -e "\e[92mDeploying cAdvisor ..." > /dev/console

cat > /root/config/cadvisor-preperations.yaml << __CADVISOR_PREPERATIONS__
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: cadvisor
  name: cadvisor
  namespace: vmware-system
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  labels:
    app: cadvisor
  name: cadvisor
rules:
  - apiGroups: ['policy']
    resources: ['podsecuritypolicies']
    verbs:     ['use']
    resourceNames:
    - cadvisor
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  labels:
    app: cadvisor
  name: cadvisor
  namespace: vmware-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cadvisor
subjects:
- kind: ServiceAccount
  name: cadvisor
  namespace: vmware-system
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  labels:
    app: cadvisor
  name: cadvisor
spec:
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  runAsUser:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  volumes:
  - '*'
  allowedHostPaths:
  - pathPrefix: "/"
  - pathPrefix: "/var/run"
  - pathPrefix: "/sys"
  - pathPrefix: "/var/lib/docker"
  - pathPrefix: "/dev/disk"
__CADVISOR_PREPERATIONS__

cat > /root/config/cadvisor-ds.yaml << __CADVISOR_DS__
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: cadvisor
  name: cadvisor
  namespace: vmware-system
  annotations:
      seccomp.security.alpha.kubernetes.io/pod: 'docker/default'
spec:
  selector:
    matchLabels:
      app: cadvisor
  template:
    metadata:
      labels:
        app: cadvisor
    spec:
      serviceAccountName: cadvisor
      containers:
      - name: cadvisor
        image: ${CADVISOR_IMAGE}:${CADVISOR_VERSION}
        resources:
          requests:
            memory: 400Mi
            cpu: 400m
          limits:
            memory: 2000Mi
            cpu: 800m
        volumeMounts:
        - name: rootfs
          mountPath: /rootfs
          readOnly: true
        - name: var-run
          mountPath: /var/run
          readOnly: true
        - name: sys
          mountPath: /sys
          readOnly: true
        - name: docker
          mountPath: /var/lib/docker
          readOnly: true
        - name: disk
          mountPath: /dev/disk
          readOnly: true
        ports:
          - name: http
            containerPort: 8080
            protocol: TCP
        args:
         - --url_base_prefix=/top
      automountServiceAccountToken: false
      terminationGracePeriodSeconds: 30
      volumes:
      - name: rootfs
        hostPath:
          path: /
      - name: var-run
        hostPath:
          path: /var/run
      - name: sys
        hostPath:
          path: /sys
      - name: docker
        hostPath:
          path: /var/lib/docker
      - name: disk
        hostPath:
          path: /dev/disk
__CADVISOR_DS__

cat > /root/config/cadvisor-svc.yaml << __CADVISOR_SVC__
apiVersion: v1
kind: Service
metadata:
  labels:
    app: cadvisor
  name: cadvisor
  namespace: vmware-system
spec:
  selector:
    app: cadvisor
  ports:
    - name: http
      protocol: TCP
      port: 8080
      targetPort: 8080
  sessionAffinity: None
__CADVISOR_SVC__

kubectl apply -f /root/config/cadvisor-preperations.yaml
kubectl apply -f /root/config/cadvisor-ds.yaml
kubectl apply -f /root/config/cadvisor-svc.yaml
kubectl wait --for=condition=ready pod -l app=cadvisor --timeout=3m -n vmware-system