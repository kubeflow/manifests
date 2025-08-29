{{/*
Expand the name of the chart.
*/}}
{{- define "notebook-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "notebook-controller.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "notebook-controller.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "notebook-controller.labels" -}}
helm.sh/chart: {{ include "notebook-controller.chart" . }}
{{ include "notebook-controller.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "notebook-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "notebook-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "notebook-controller.serviceAccountName" -}}
{{- if .Values.rbac.serviceAccount.create }}
{{- default (printf "%s-service-account" (include "notebook-controller.fullname" .)) .Values.rbac.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.rbac.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Image name helper
*/}}
{{- define "notebook-controller.image" -}}
{{- $registry := .Values.global.imageRegistry -}}
{{- $repository := .Values.controller.image.repository -}}
{{- $tag := .Values.controller.image.tag | default .Values.global.imageTag -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end }}

{{/*
Deployment mode helpers
*/}}
{{- define "notebook-controller.isKubeflowMode" -}}
{{- eq .Values.deploymentMode "kubeflow" -}}
{{- end }}

{{- define "notebook-controller.isStandaloneMode" -}}
{{- eq .Values.deploymentMode "standalone" -}}
{{- end }}

{{/*
Namespace helper
*/}}
{{- define "notebook-controller.namespace" -}}
{{- if eq .Values.deploymentMode "kubeflow" -}}
{{- .Values.global.namespace -}}
{{- else -}}
{{- printf "%s-system" (include "notebook-controller.name" .) -}}
{{- end -}}
{{- end }}

{{/*
Resource name prefix helper 
*/}}
{{- define "notebook-controller.namePrefix" -}}
{{- printf "%s-" (include "notebook-controller.name" .) -}}
{{- end }}

{{/*
Full resource name helper
*/}}
{{- define "notebook-controller.resourceName" -}}
{{- if eq .Values.deploymentMode "kubeflow" -}}
{{- include "notebook-controller.fullname" . -}}
{{- else -}}
{{- printf "notebook-controller" -}}
{{- end -}}
{{- end }}

{{/*
Image pull policy helper
*/}}
{{- define "notebook-controller.imagePullPolicy" -}}
{{- .Values.controller.image.pullPolicy | default .Values.global.imagePullPolicy -}}
{{- end }}

{{/*
Config map name helper
*/}}
{{- define "notebook-controller.configMapName" -}}
{{- if eq .Values.deploymentMode "kubeflow" -}}
{{- printf "%s-config" (include "notebook-controller.fullname" .) -}}
{{- else -}}
{{- printf "%s-config" (include "notebook-controller.resourceName" .) -}}
{{- end -}}
{{- end }}

{{/*
Auth proxy image helper
*/}}
{{- define "notebook-controller.authProxyImage" -}}
{{- printf "%s/%s:%s" .Values.controller.authProxy.image.registry .Values.controller.authProxy.image.repository .Values.controller.authProxy.image.tag -}}
{{- end }}

{{/*
Metrics service name helper
*/}}
{{- define "notebook-controller.metricsServiceName" -}}
{{- if .Values.metricsService.name -}}
{{- .Values.metricsService.name -}}
{{- else -}}
{{- printf "%s-metrics-service" (include "notebook-controller.resourceName" .) -}}
{{- end -}}
{{- end }}

{{/*
Common annotations helper
*/}}
{{- define "notebook-controller.annotations" -}}
{{- with .Values.commonAnnotations }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{/*
Pod annotations helper (combines common and monitoring annotations)
*/}}
{{- define "notebook-controller.podAnnotations" -}}
{{- $annotations := dict -}}
{{- if .Values.commonAnnotations -}}
{{- $annotations = mergeOverwrite $annotations .Values.commonAnnotations -}}
{{- end -}}
{{- if .Values.controller.annotations -}}
{{- $annotations = mergeOverwrite $annotations .Values.controller.annotations -}}
{{- end -}}
{{- if and .Values.monitoring.enabled .Values.monitoring.prometheus.enabled (not .Values.controller.authProxy.enabled) -}}
{{- $annotations = mergeOverwrite $annotations .Values.monitoring.prometheus.podAnnotations -}}
{{- end -}}
{{- if $annotations -}}
{{- toYaml $annotations -}}
{{- end -}}
{{- end }}

{{/*
Deployment selector labels (consistent across all resources)
*/}}
{{- define "notebook-controller.deploymentSelectorLabels" -}}
{{- if and .Values.kustomizeMode.enabled .Values.kustomizeMode.useOriginalLabels -}}
app: notebook-controller
kustomize.component: notebook-controller
{{- else -}}
{{- include "notebook-controller.selectorLabels" . -}}
{{- end -}}
{{- end }}

{{/*
Manager container name helper
*/}}
{{- define "notebook-controller.managerContainerName" -}}
{{- if .Values.kustomizeMode.enabled -}}
manager
{{- else -}}
controller-manager
{{- end -}}
{{- end }}

{{/*
Service port name helper
*/}}
{{- define "notebook-controller.servicePortName" -}}
{{- if .Values.controller.authProxy.enabled -}}
https
{{- else -}}
metrics
{{- end -}}
{{- end }}

{{/*
Conditional environment variables helper
*/}}
{{- define "notebook-controller.envVars" -}}
- name: USE_ISTIO
  valueFrom:
    configMapKeyRef:
      name: {{ include "notebook-controller.configMapName" . }}
      key: USE_ISTIO
- name: ISTIO_GATEWAY
  valueFrom:
    configMapKeyRef:
      name: {{ include "notebook-controller.configMapName" . }}
      key: ISTIO_GATEWAY
- name: ISTIO_HOST
  valueFrom:
    configMapKeyRef:
      name: {{ include "notebook-controller.configMapName" . }}
      key: ISTIO_HOST
- name: CLUSTER_DOMAIN
  valueFrom:
    configMapKeyRef:
      name: {{ include "notebook-controller.configMapName" . }}
      key: CLUSTER_DOMAIN
- name: ENABLE_CULLING
  valueFrom:
    configMapKeyRef:
      name: {{ include "notebook-controller.configMapName" . }}
      key: ENABLE_CULLING
- name: CULL_IDLE_TIME
  valueFrom:
    configMapKeyRef:
      name: {{ include "notebook-controller.configMapName" . }}
      key: CULL_IDLE_TIME
- name: IDLENESS_CHECK_PERIOD
  valueFrom:
    configMapKeyRef:
      name: {{ include "notebook-controller.configMapName" . }}
      key: IDLENESS_CHECK_PERIOD
{{- with .Values.env.additional }}
{{- toYaml . }}
{{- end }}
{{- end }}