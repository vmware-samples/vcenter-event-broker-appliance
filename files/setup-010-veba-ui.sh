#!/bin/bash
# Copyright 2019 VMware, Inc. All rights reserved.
# SPDX-License-Identifier: BSD-2

# Setup Knative UI

set -euo pipefail

echo -e "\e[92mSetting up VEBA UI RBAC ..." > /dev/console
cat > /root/config/veba-ui-rbac.yaml << EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: veba-ui
  namespace: vmware-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: veba-ui
  namespace: vmware-functions
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["watch", "get", "list", "create", "update", "delete"]
  - apiGroups: ["serving.knative.dev"]
    resources: ["services"]
    verbs: ["watch", "get", "list", "create", "update", "delete"]
  - apiGroups: ["eventing.knative.dev"]
    resources: ["triggers"]
    verbs: ["watch", get", "list", "create", "update", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: veba-ui
  namespace: vmware-functions
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: veba-ui
subjects:
- kind: ServiceAccount
  name: veba-ui
  namespace: vmware-system
EOF

kubectl apply -f /root/config/veba-ui-rbac.yaml

ESCAPED_VCENTER_SERVER=$(echo -n ${VCENTER_SERVER} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')
ESCAPED_VCENTER_USERNAME_FOR_VEBA_UI=$(echo -n ${VCENTER_USERNAME_FOR_VEBA_UI} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')
ESCAPED_VCENTER_PASSWORD_FOR_VEBA_UI=$(echo -n ${VCENTER_PASSWORD_FOR_VEBA_UI} | python -c 'import sys,json;data=sys.stdin.read(); print json.dumps(data)[1:-1]')

echo -e "\e[92mSetting up VEBA UI Secret ..." > /dev/console
kubectl -n vmware-system create secret generic veba-ui-secret \
    --from-literal=VCENTER_FQDN=${ESCAPED_VCENTER_SERVER} \
    --from-literal=VCENTER_PORT=443 \
    --from-literal=VCENTER_USER=${ESCAPED_VCENTER_USERNAME_FOR_VEBA_UI} \
    --from-literal=VCENTER_PASS=${ESCAPED_VCENTER_PASSWORD_FOR_VEBA_UI} \
    --from-literal=VEBA_FQDN=${HOSTNAME}

# Retrieve the VEBA UI IMage
VEBA_BOM_FILE=/root/config/veba-bom.json
VEBA_UI_IMAGE=$(jq -r < ${VEBA_BOM_FILE} '.["veba-ui"].containers[0].name')
VEBA_UI_VERSION=$(jq -r < ${VEBA_BOM_FILE} '.["veba-ui"].containers[0].version')

echo -e "\e[92mSetting up VEBA UI ..." > /dev/console
cat > /root/config/veba-ui.yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: veba-ui
  name: veba-ui
  namespace: vmware-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: veba-ui
  template:
    metadata:
      labels:
        app: veba-ui
    spec:
      serviceAccountName: veba-ui
      containers:
      - image: ${VEBA_UI_IMAGE}:${VEBA_UI_VERSION}
        imagePullPolicy: IfNotPresent
        name: veba-ui
        ports:
          - containerPort: 8080
        envFrom:
          - secretRef:
              name: veba-ui-secret
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: veba-ui
  name: veba-ui
  namespace: vmware-system
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: veba-ui
  sessionAffinity: None
EOF

kubectl apply -f /root/config/veba-ui.yaml