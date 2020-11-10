{{/*
Expand the name of the chart.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "router.name" -}}
{{- default "router" .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "router.fullname" -}}
{{- $name := default "event-router" .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Helm required labels */}}
{{- define "router.labels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: "{{ template "router.fullname" . }}"
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
helm.sh/chart: {{ .Chart.Name }}
{{- end -}}

{{/* matchLabels */}}
{{- define "router.matchLabels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: "{{ template "router.fullname" . }}"
app.kubernetes.io/component: {{ .Chart.Name }}
{{- end -}}
