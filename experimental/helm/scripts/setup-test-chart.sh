#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
HELM_DIR="$(dirname "$SCRIPT_DIR")"
CHART_DIR="$HELM_DIR/kubeflow"


mkdir -p "$CHART_DIR/templates/external"
mkdir -p "$CHART_DIR/templates/integrations"
mkdir -p "$CHART_DIR/templates/_helpers"
mkdir -p "$CHART_DIR/crds"

cat > "$CHART_DIR/Chart.yaml" << 'EOF'
apiVersion: v2
name: kubeflow
description: Kubeflow All-in-One Helm Chart (Test Version)
type: application
version: 0.1.0
appVersion: "v1.10.0"
EOF

cat > "$CHART_DIR/values.yaml" << 'EOF'
# Global configuration
global:
  kubeflowNamespace: kubeflow
  certManagerNamespace: cert-manager

# Component configurations
sparkOperator:
  enabled: false
  kubeflowRBAC:
    enabled: false
  spark:
    jobNamespaces: []
  webhook:
    enable: true
    port: 9443

certManager:
  enabled: false
  installCRDs: true
  global:
    leaderElection:
      namespace: kube-system
  startupapicheck:
    enabled: false
  kubeflowIssuer:
    enabled: true
    name: kubeflow-self-signing-issuer
EOF

cat > "$CHART_DIR/templates/_helpers.tpl" << 'EOF'
{{/*
Common labels
*/}}
{{- define "kubeflow.labels" -}}
app.kubernetes.io/name: {{ include "kubeflow.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Chart name
*/}}
{{- define "kubeflow.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Kubeflow namespace
*/}}
{{- define "kubeflow.namespace" -}}
{{- .Values.global.kubeflowNamespace | default "kubeflow" }}
{{- end }}
EOF

cat > "$CHART_DIR/templates/integrations/spark-operator-rbac.yaml" << 'EOF'
{{- if and .Values.sparkOperator.enabled .Values.sparkOperator.kubeflowRBAC.enabled }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-spark-admin
  labels:
    app: spark-operator
    app.kubernetes.io/name: spark-operator
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
rules: []
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-spark-edit
  labels:
    app: spark-operator
    app.kubernetes.io/name: spark-operator
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
rules:
  - apiGroups:
      - sparkoperator.k8s.io
    resources:
      - sparkapplications
      - scheduledsparkapplications
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - sparkoperator.k8s.io
    resources:
      - sparkapplications/status
      - scheduledsparkapplications/status
    verbs:
      - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-spark-view
  labels:
    app: spark-operator
    app.kubernetes.io/name: spark-operator
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
  - apiGroups:
      - sparkoperator.k8s.io
    resources:
      - sparkapplications
      - scheduledsparkapplications
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - sparkoperator.k8s.io
    resources:
      - sparkapplications/status
      - scheduledsparkapplications/status
    verbs:
      - get
{{- end }}
EOF

