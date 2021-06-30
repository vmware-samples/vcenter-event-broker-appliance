#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup FluentBit

set -euo pipefail

# Retrieve the FluentBit image
VEBA_BOM_FILE=/root/config/veba-bom.json
FLUENTBIT_IMAGE=$(jq -r < ${VEBA_BOM_FILE} '.["fluentbit"].containers[0].name')
FLUENTBIT_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["fluentbit"].containers[0].version')

cat > /root/config/fluentbit-preperations.yaml << __FLUENTBIT_PREPERATIONS__
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fluent-bit
  namespace: vmware-system
  labels:
    k8s-app: fluent-bit
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: fluent-bit-read
  labels:
    k8s-app: fluent-bit
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  - pods
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: fluent-bit-read
  labels:
    k8s-app: fluent-bit
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: fluent-bit-read
subjects:
- kind: ServiceAccount
  name: fluent-bit
  namespace: vmware-system
__FLUENTBIT_PREPERATIONS__

cat > /root/config/fluentbit-configmap.yaml << __FLUENTBIT_CONFIGMAP__
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: vmware-system
  labels:
    k8s-app: fluent-bit
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush         1
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
        HTTP_Server   On
        HTTP_Listen   0.0.0.0
        HTTP_Port     2020

    @INCLUDE input-kubernetes.conf
    @INCLUDE filter-kubernetes.conf
    @INCLUDE filter-record.conf
    @INCLUDE output-syslog.conf

  input-kubernetes.conf: |
    [INPUT]
        Name                tail
        Tag                 kube.*
        Path                /var/log/containers/*.log
        Parser              docker
        DB                  /var/log/flb_kube.db
        Mem_Buf_Limit       10MB
        Skip_Long_Lines     On
        Refresh_Interval    10

  filter-kubernetes.conf: |
    [FILTER]
        Name                kubernetes
        Match               kube.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Kube_Tag_Prefix     kube.var.log.containers.
        Merge_Log           On
        Merge_Log_Key       log_processed
        K8S-Logging.Parser  On
        K8S-Logging.Exclude Off

    [FILTER]
        Name                modify
        Match               kube.*
        Copy                kubernetes k8s

    [FILTER]
        Name                nest
        Match               kube.*
        Operation           lift
        Nested_Under        kubernetes

  filter-record.conf: |
    [FILTER]
        Name                 record_modifier
        Match                *
        Record veba_instance ${HOSTNAME}
        Record veba_cluster  ${HOSTNAME}

    [FILTER]
        Name                nest
        Match               kube.*
        Operation           nest
        Wildcard            veba_instance*
        Nest_Under          veba

  output-syslog.conf: |
    [OUTPUT]
        Name                 syslog
        Match                *
        Host                 ${SYSLOG_SERVER_HOSTNAME}
        Port                 ${SYSLOG_SERVER_PORT}
        Mode                 ${SYSLOG_SERVER_PROTOCOL}
        Syslog_Format        ${SYSLOG_SERVER_FORMAT}
        Syslog_Hostname_key  veba_cluster
        Syslog_Appname_key   pod_name
        Syslog_Procid_key    container_name
        Syslog_Message_key   message
        syslog_msgid_key     msgid
        Syslog_SD_key        k8s
        Syslog_SD_key        labels
        Syslog_SD_key        annotations
        Syslog_SD_key        veba

  parsers.conf: |
    [PARSER]
        Name                 json
        Format               json
        Time_Key             time
        Time_Format          %d/%b/%Y:%H:%M:%S %z

    [PARSER]
        Name                 docker
        Format               json
        Time_Key             time
        Time_Format          %Y-%m-%dT%H:%M:%S.%L
        Time_Keep            On

    [PARSER]
        Name                 cri
        Format               regex
        Regex                ^(?<time>[^ ]+) (?<stream>stdout|stderr) (?<logtag>[^ ]*) (?<message>.*)$
        Time_Key             time
        Time_Format          %Y-%m-%dT%H:%M:%S.%L%z

    [PARSER]
        Name                 syslog
        Format               regex
        Regex                ^\<(?<pri>[0-9]{1,5})\>1 (?<time>[^ ]+) (?<host>[^ ]+) (?<ident>[^ ]+) (?<pid>[-0-9]+) (?<msgid>[^ ]+) (?<extradata>(\[(.*)\]|-)) (?<message>.+)$
        Time_Key             time
        Time_Format          %Y-%m-%dT%H:%M:%S.%L
        Time_Keep            On
__FLUENTBIT_CONFIGMAP__

cat > /root/config/fluentbit-ds.yaml << EOF
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluent-bit
  namespace: vmware-system
  labels:
    k8s-app: fluent-bit
    version: v1
    kubernetes.io/cluster-service: "true"
spec:
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        k8s-app: fluent-bit
        version: v1
        kubernetes.io/cluster-service: "true"
    spec:
      containers:
      - name: fluent-bit
        image: ${FLUENTBIT_IMAGE}:${FLUENTBIT_VERSION}
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 2020
        resources:
          requests:
            cpu: 5m
            memory: 10Mi
          limits:
            cpu: 50m
            memory: 60Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlogcontainers
          mountPath: /var/log/containers
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
        - name: fluent-bit-config
          mountPath: /fluent-bit/etc/
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlogcontainers
        hostPath:
          path: /var/log/containers
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: fluent-bit-config
        configMap:
          name: fluent-bit-config
      serviceAccountName: fluent-bit
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      - operator: Exists
        effect: NoExecute
      - operator: Exists
        effect: NoSchedule
  selector:
    matchLabels:
      k8s-app: fluent-bit
      version: v1
      kubernetes.io/cluster-service: "true"
EOF

kubectl apply -f /root/config/fluentbit-preperations.yaml
kubectl apply -f /root/config/fluentbit-configmap.yaml
kubectl apply -f /root/config/fluentbit-ds.yaml
kubectl wait --for=condition=ready pod -l k8s-app=fluent-bit --timeout=3m -n vmware-system