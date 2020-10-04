#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup VMware Event Router

set -euo pipefail

echo -e "\e[92mDeploying VMware Event Router ..." > /dev/console
kubectl --kubeconfig /root/.kube/config -n vmware create secret generic event-router-config --from-file=${EVENT_ROUTER_CONFIG}

# Retrieve the VMware Event Router image
VEBA_BOM_FILE=/root/config/veba-bom.json
EVENT_ROUTER_IMAGE=$(jq -r < ${VEBA_BOM_FILE} '.["vmware-event-router"].containers[0].name')
EVENT_ROUTER_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["vmware-event-router"].containers[0].version')

cat > /root/config/event-router-k8s.yaml << __EVENT_ROUTER_CONFIG
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: vmware-event-router
  name: vmware-event-router
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vmware-event-router
  template:
    metadata:
      labels:
        app: vmware-event-router
    spec:
      containers:
      - image: ${EVENT_ROUTER_IMAGE}:${EVENT_ROUTER_VERSION}
        imagePullPolicy: IfNotPresent
        args: ["-config", "/etc/vmware-event-router/event-router-config.yaml"]
        name: vmware-event-router
        resources:
          requests:
            cpu: 200m
            memory: 200Mi
        volumeMounts:
        - name: config
          mountPath: /etc/vmware-event-router/
          readOnly: true
      volumes:
      - name: config
        secret:
          secretName: event-router-config
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: vmware-event-router
  name: vmware-event-router
spec:
  ports:
  - port: 8082
    protocol: TCP
    targetPort: 8082
  selector:
    app: vmware-event-router
  sessionAffinity: None
__EVENT_ROUTER_CONFIG

kubectl --kubeconfig /root/.kube/config -n vmware create -f /root/config/event-router-k8s.yaml