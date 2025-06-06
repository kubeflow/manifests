{{/*
Expand the name of the chart.
*/}}
{{- define "kubeflow.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "kubeflow.fullname" -}}
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
{{- define "kubeflow.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kubeflow.labels" -}}
helm.sh/chart: {{ include "kubeflow.chart" . }}
{{ include "kubeflow.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kubeflow.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubeflow.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Get the namespace for Kubeflow components
*/}}
{{- define "kubeflow.namespace" -}}
{{- .Values.global.kubeflowNamespace | default "kubeflow" }}
{{- end }}

{{/*
Get the namespace for Istio components
*/}}
{{- define "kubeflow.istioNamespace" -}}
{{- .Values.global.istioNamespace | default "istio-system" }}
{{- end }}

{{/*
Get the namespace for cert-manager components
*/}}
{{- define "kubeflow.certManagerNamespace" -}}
{{- .Values.global.certManagerNamespace | default "cert-manager" }}
{{- end }}

{{/*
Get the domain for Kubeflow
*/}}
{{- define "kubeflow.domain" -}}
{{- .Values.global.domain | default "example.com" }}
{{- end }} 