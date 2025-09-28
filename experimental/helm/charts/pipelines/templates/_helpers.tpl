{{/*
Expand the name of the chart.
*/}}
{{- define "kubeflow-pipelines.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kubeflow-pipelines.fullname" -}}
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
{{- define "kubeflow-pipelines.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels 
*/}}
{{- define "kubeflow-pipelines.labels" -}}
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- if ne .Values.installMode.type "multi-user" }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kubeflow-pipelines.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubeflow-pipelines.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Cache server labels
*/}}
{{- define "kubeflow-pipelines.cacheLabels" -}}
app: cache-server
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- if ne .Values.installMode.type "multi-user" }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Cache deployer labels
*/}}
{{- define "kubeflow-pipelines.cacheDeployerLabels" -}}
app: cache-deployer
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
application-crd-id: kubeflow-pipelines
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
ML Pipeline specific labels - matching original manifests
*/}}
{{- define "kubeflow-pipelines.mlPipelineLabels" -}}
app: ml-pipeline
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- if ne .Values.installMode.type "multi-user" }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
ML Pipeline UI labels
*/}}
{{- define "kubeflow-pipelines.uiLabels" -}}
app: ml-pipeline-ui
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- if ne .Values.installMode.type "multi-user" }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
ML Pipeline selector labels
*/}}
{{- define "kubeflow-pipelines.mlPipelineSelectorLabels" -}}
app: ml-pipeline
{{- if eq .Values.installMode.type "multi-user" }}
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- else }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- end }}

{{/*
Cache server selector labels
*/}}
{{- define "kubeflow-pipelines.cacheSelectorLabels" -}}
app: cache-server
{{- if eq .Values.installMode.type "multi-user" }}
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- else }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- end }}

{{/*
UI selector labels
*/}}
{{- define "kubeflow-pipelines.uiSelectorLabels" -}}
app: ml-pipeline-ui
{{- if eq .Values.installMode.type "multi-user" }}
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- else }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- end }}

{{/*
Viewer CRD labels
*/}}
{{- define "kubeflow-pipelines.viewerCrdLabels" -}}
{{- if eq .Values.installMode.type "multi-user" }}
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
app: ml-pipeline-viewer-crd
{{- else }}
{{- include "kubeflow-pipelines.labels" . }}
app: ml-pipeline-viewer-crd
{{- end }}
{{- end }}

{{/*
Viewer CRD selector labels
*/}}
{{- define "kubeflow-pipelines.viewerCrdSelectorLabels" -}}
{{- if eq .Values.installMode.type "multi-user" }}
app: ml-pipeline-viewer-crd
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- else }}
app: ml-pipeline-viewer-crd
application-crd-id: kubeflow-pipelines
{{- end }}
{{- end }}

{{/*
Visualization server labels
*/}}
{{- define "kubeflow-pipelines.visualizationLabels" -}}
app: ml-pipeline-visualizationserver
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- if ne .Values.installMode.type "multi-user" }}
application-crd-id: kubeflow-pipelines
{{- end}}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Visualization server selector labels
*/}}
{{- define "kubeflow-pipelines.visualizationSelectorLabels" -}}
{{- if eq .Values.installMode.type "multi-user" }}
app: ml-pipeline-visualizationserver
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- else }}
app: ml-pipeline-visualizationserver
application-crd-id: kubeflow-pipelines
{{- end }}
{{- end }}

{{/*
MySQL selector labels
*/}}
{{- define "kubeflow-pipelines.mysqlSelectorLabels" -}}
app: mysql
application-crd-id: kubeflow-pipelines
{{- end }}

{{/*
MinIO selector labels
*/}}
{{- define "kubeflow-pipelines.minioSelectorLabels" -}}
app: minio
application-crd-id: kubeflow-pipelines
{{- end }}

{{/*
Metadata GRPC selector labels
*/}}
{{- define "kubeflow-pipelines.metadataGrpcSelectorLabels" -}}
component: metadata-grpc-server
application-crd-id: kubeflow-pipelines
{{- end }}

{{/*
Metadata Envoy selector labels
*/}}
{{- define "kubeflow-pipelines.metadataEnvoySelectorLabels" -}}
component: metadata-envoy
application-crd-id: kubeflow-pipelines
{{- end }}

{{/*
Cache deployer selector labels
*/}}
{{- define "kubeflow-pipelines.cacheDeployerSelectorLabels" -}}
app: cache-deployer
application-crd-id: kubeflow-pipelines
{{- end }}

{{/*
Metadata writer selector labels
*/}}
{{- define "kubeflow-pipelines.metadataWriterSelectorLabels" -}}
app: metadata-writer
{{- if eq .Values.installMode.type "multi-user" }}
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- else }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- end }}

{{/*
Persistence agent selector labels
*/}}
{{- define "kubeflow-pipelines.persistenceAgentSelectorLabels" -}}
app: ml-pipeline-persistenceagent
{{- if eq .Values.installMode.type "multi-user" }}
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- else }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- end }}

{{/*
Scheduled workflow selector labels
*/}}
{{- define "kubeflow-pipelines.scheduledWorkflowSelectorLabels" -}}
app: ml-pipeline-scheduledworkflow
{{- if eq .Values.installMode.type "multi-user" }}
app.kubernetes.io/component: ml-pipeline
app.kubernetes.io/name: kubeflow-pipelines
{{- else }}
application-crd-id: kubeflow-pipelines
{{- end }}
{{- end }}

{{/*
Workflow controller selector labels
*/}}
{{- define "kubeflow-pipelines.workflowControllerSelectorLabels" -}}
app: workflow-controller
application-crd-id: kubeflow-pipelines
{{- end }}

{{/*
Create the name of the service account to use for API Server
*/}}
{{- define "kubeflow-pipelines.apiServer.serviceAccountName" -}}
{{- if .Values.apiServer.serviceAccount.create }}
{{- default "ml-pipeline" .Values.apiServer.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.apiServer.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for Persistence Agent
*/}}
{{- define "kubeflow-pipelines.persistenceAgent.serviceAccountName" -}}
{{- if .Values.persistenceAgent.serviceAccount.create }}
{{- default "ml-pipeline-persistenceagent" .Values.persistenceAgent.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.persistenceAgent.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for Scheduled Workflow
*/}}
{{- define "kubeflow-pipelines.scheduledWorkflow.serviceAccountName" -}}
{{- if .Values.scheduledWorkflow.serviceAccount.create }}
{{- default "ml-pipeline-scheduledworkflow" .Values.scheduledWorkflow.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.scheduledWorkflow.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for UI
*/}}
{{- define "kubeflow-pipelines.ui.serviceAccountName" -}}
{{- if .Values.ui.serviceAccount.create }}
{{- default "ml-pipeline-ui" .Values.ui.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.ui.serviceAccount.name }}
{{- end }}
{{- end }}



{{/*
Create the name of the service account to use for Cache Server
*/}}
{{- define "kubeflow-pipelines.cache.serviceAccountName" -}}
{{- if .Values.cache.serviceAccount.create }}
{{- default "kubeflow-pipelines-cache" .Values.cache.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.cache.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for Cache Deployer
*/}}
{{- define "kubeflow-pipelines.cacheDeployer.serviceAccountName" -}}
{{- if .Values.cacheDeployer.serviceAccount.create }}
{{- default "kubeflow-pipelines-cache-deployer-sa" .Values.cacheDeployer.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.cacheDeployer.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for Metadata
*/}}
{{- define "kubeflow-pipelines.metadata.serviceAccountName" -}}
{{- if .Values.metadata.grpc.serviceAccount.create }}
{{- default "metadata-grpc-server" .Values.metadata.grpc.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.metadata.grpc.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for Viewer CRD Controller
*/}}
{{- define "kubeflow-pipelines.viewerCrd.serviceAccountName" -}}
{{- if .Values.viewerCrd.serviceAccount.create }}
{{- default "ml-pipeline-viewer-crd-service-account" .Values.viewerCrd.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.viewerCrd.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for Visualization Server
*/}}
{{- define "kubeflow-pipelines.visualization.serviceAccountName" -}}
{{- if .Values.visualization.serviceAccount.create }}
{{- default "ml-pipeline-visualizationserver" .Values.visualization.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.visualization.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use for Metadata Writer
*/}}
{{- define "kubeflow-pipelines.metadataWriter.serviceAccountName" -}}
{{- if .Values.metadataWriter.serviceAccount.create }}
{{- default "kubeflow-pipelines-metadata-writer" .Values.metadataWriter.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.metadataWriter.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database configuration helpers
*/}}
{{- define "kubeflow-pipelines.database.host" -}}
{{- if .Values.mysql.enabled -}}
mysql
{{- else if .Values.postgresql.enabled -}}
postgresql
{{- else -}}
{{- .Values.externalDatabase.host | default "mysql" }}
{{- end -}}
{{- end }}

{{- define "kubeflow-pipelines.database.port" -}}
{{- if .Values.mysql.enabled -}}
3306
{{- else if .Values.postgresql.enabled -}}
5432
{{- else -}}
{{- .Values.externalDatabase.port }}
{{- end -}}
{{- end }}

{{- define "kubeflow-pipelines.database.name" -}}
{{- if .Values.mysql.enabled -}}
{{- .Values.mysql.auth.database | default "mlpipeline" }}
{{- else if .Values.postgresql.enabled -}}
{{- .Values.postgresql.auth.database | default "mlpipeline" }}
{{- else -}}
{{- .Values.externalDatabase.database }}
{{- end -}}
{{- end }}

{{- define "kubeflow-pipelines.database.username" -}}
{{- if .Values.mysql.enabled -}}
{{- .Values.mysql.auth.username | default "mlpipeline" }}
{{- else if .Values.postgresql.enabled -}}
{{- .Values.postgresql.auth.username | default "mlpipeline" }}
{{- else -}}
{{- .Values.externalDatabase.username }}
{{- end -}}
{{- end }}

{{- define "kubeflow-pipelines.database.secretName" -}}
{{- if .Values.externalDatabase.existingSecret -}}
{{- .Values.externalDatabase.existingSecret }}
{{- else if .Values.mysql.enabled -}}
mysql-secret
{{- else if .Values.postgresql.enabled -}}
postgresql-secret
{{- else -}}
mysql-secret
{{- end -}}
{{- end }}

{{- define "kubeflow-pipelines.database.secretKey" -}}
{{- if .Values.externalDatabase.existingSecret -}}
{{- .Values.externalDatabase.existingSecretPasswordKey | default "password" }}
{{- else if .Values.mysql.enabled -}}
mysql-password
{{- else if .Values.postgresql.enabled -}}
password
{{- else -}}
password
{{- end -}}
{{- end }}

{{/*
Object storage configuration helpers
*/}}
{{- define "kubeflow-pipelines.objectStore.endpoint" -}}
{{- if eq .Values.objectStore.provider "minio" -}}
{{- if .Values.minio.enabled -}}
{{- include "minio.fullname" .Subcharts.minio }}:9000
{{- else -}}
{{- .Values.objectStore.minio.endpoint }}
{{- end -}}
{{- else if eq .Values.objectStore.provider "s3" -}}
s3.amazonaws.com
{{- else if eq .Values.objectStore.provider "gcs" -}}
storage.googleapis.com
{{- else if eq .Values.objectStore.provider "azure" -}}
{{ .Values.objectStore.azure.storageAccount }}.blob.core.windows.net
{{- else -}}
{{ .Values.objectStore.custom.endpoint }}
{{- end -}}
{{- end }}

{{- define "kubeflow-pipelines.objectStore.bucket" -}}
{{- if eq .Values.objectStore.provider "minio" -}}
{{- .Values.objectStore.minio.bucket | default "mlpipeline" }}
{{- else if eq .Values.objectStore.provider "s3" -}}
{{- .Values.objectStore.bucketName | default .Values.objectStore.s3.bucket }}
{{- else if eq .Values.objectStore.provider "gcs" -}}
{{- .Values.objectStore.gcs.bucket }}
{{- else if eq .Values.objectStore.provider "azure" -}}
{{- .Values.objectStore.azure.container }}
{{- else -}}
{{- .Values.objectStore.custom.bucket }}
{{- end -}}
{{- end }}

{{- define "kubeflow-pipelines.objectStore.secure" -}}
{{- if eq .Values.objectStore.provider "minio" -}}
{{- .Values.objectStore.minio.secure | default "false" }}
{{- else if eq .Values.objectStore.provider "s3" -}}
"true"
{{- else if eq .Values.objectStore.provider "gcs" -}}
"true"
{{- else if eq .Values.objectStore.provider "azure" -}}
"true"
{{- else -}}
{{- .Values.objectStore.custom.secure | default "true" }}
{{- end -}}
{{- end }}

{{- define "kubeflow-pipelines.objectStore.secretName" -}}
{{- if eq .Values.objectStore.provider "minio" -}}
{{- if .Values.objectStore.minio.existingSecret -}}
{{- .Values.objectStore.minio.existingSecret }}
{{- else -}}
mlpipeline-minio-artifact
{{- end -}}
{{- else if eq .Values.objectStore.provider "s3" -}}
{{- if .Values.objectStore.s3.existingSecret -}}
{{- .Values.objectStore.s3.existingSecret }}
{{- else -}}
mlpipeline-minio-artifact
{{- end -}}
{{- else if eq .Values.objectStore.provider "gcs" -}}
{{- if .Values.objectStore.gcs.existingSecret -}}
{{- .Values.objectStore.gcs.existingSecret }}
{{- else -}}
mlpipeline-minio-artifact
{{- end -}}
{{- else if eq .Values.objectStore.provider "azure" -}}
{{- if .Values.objectStore.azure.existingSecret -}}
{{- .Values.objectStore.azure.existingSecret }}
{{- else -}}
mlpipeline-minio-artifact
{{- end -}}
{{- else -}}
mlpipeline-minio-artifact
{{- end -}}
{{- end }}

{{/*
Image helpers
*/}}
{{- define "kubeflow-pipelines.image" -}}
{{- $registry := .context.Values.global.imageRegistry -}}
{{- $repository := .repository -}}
{{- $tag := .tag | default .context.Values.global.imageTag | default .context.Chart.AppVersion -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end }}

{{/*
Image pull policy
*/}}
{{- define "kubeflow-pipelines.imagePullPolicy" -}}
{{- .pullPolicy | default .context.Values.global.imagePullPolicy -}}
{{- end }}

{{/*
Image pull secrets
*/}}
{{- define "kubeflow-pipelines.imagePullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.global.imagePullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Environment detection helpers
*/}}
{{- define "kubeflow-pipelines.isMultiUser" -}}
{{- eq .Values.installMode.type "multi-user" -}}
{{- end }}

{{- define "kubeflow-pipelines.isGeneric" -}}
{{- eq .Values.installMode.type "generic" -}}
{{- end }}

{{/*
Namespace helper
*/}}
{{- define "kubeflow-pipelines.namespace" -}}
{{- .Values.global.namespace | default .Release.Namespace -}}
{{- end }}

{{/*
Common annotations
*/}}
{{- define "kubeflow-pipelines.annotations" -}}
{{- with .Values.commonAnnotations }}
{{ toYaml . }}
{{- end }}
{{- with .Values.global.annotations }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Common environment variables for all components
*/}}
{{- define "kubeflow-pipelines.commonEnv" -}}
- name: POD_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
{{- if .Values.global.logLevel }}
- name: LOG_LEVEL
  value: {{ .Values.global.logLevel | quote }}
{{- end }}
{{- end }}

{{/*
Database environment variables
*/}}
{{- define "kubeflow-pipelines.databaseEnv" -}}
- name: DBCONFIG_HOST
  value: {{ include "kubeflow-pipelines.database.host" . | quote }}
- name: DBCONFIG_PORT
  value: {{ include "kubeflow-pipelines.database.port" . | quote }}
- name: DBCONFIG_DBNAME
  value: {{ include "kubeflow-pipelines.database.name" . | quote }}
- name: DBCONFIG_USER
  valueFrom:
    secretKeyRef:
      name: {{ include "kubeflow-pipelines.database.secretName" . }}
      key: username
- name: DBCONFIG_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "kubeflow-pipelines.database.secretName" . }}
      key: {{ include "kubeflow-pipelines.database.secretKey" . }}
- name: DB_DRIVER_NAME
  value: {{ .Values.externalDatabase.type | default "mysql" | quote }}
{{- if .Values.database.connectionMaxLifetime }}
- name: DBCONFIG_CONMAXLIFETIME
  value: {{ .Values.database.connectionMaxLifetime | quote }}
{{- end }}
{{- end }}

{{/*
Object store environment variables
*/}}
{{- define "kubeflow-pipelines.objectStoreEnv" -}}
- name: OBJECTSTORECONFIG_SECURE
  value: {{ include "kubeflow-pipelines.objectStore.secure" . | quote }}
- name: OBJECTSTORECONFIG_BUCKETNAME
  value: {{ include "kubeflow-pipelines.objectStore.bucket" . | quote }}
- name: OBJECTSTORECONFIG_ACCESSKEY
  valueFrom:
    secretKeyRef:
      name: {{ include "kubeflow-pipelines.objectStore.secretName" . }}
      key: accesskey
- name: OBJECTSTORECONFIG_SECRETACCESSKEY
  valueFrom:
    secretKeyRef:
      name: {{ include "kubeflow-pipelines.objectStore.secretName" . }}
      key: secretkey
{{- end }}
