#!/usr/bin/env bash
# Script to sync Cert Manager templates for AIO Helm chart

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
HELM_DIR="$(dirname "$SCRIPT_DIR")"
CHART_DIR="$HELM_DIR/kubeflow"

COMPONENT="cert-manager"
VERSION="v1.16.1"
REPO="https://charts.jetstack.io"
TEMPLATES_DIR="$CHART_DIR/templates/external/${COMPONENT}"
NAMESPACE="cert-manager"

rm -rf "$TEMPLATES_DIR"
mkdir -p "$TEMPLATES_DIR"

TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# Generate templates using same settings as existing Kustomize setup
# Disable startupapicheck to match Kustomize manifests that don't include it
helm template "$COMPONENT" "$COMPONENT" \
    --version "$VERSION" \
    --repo "$REPO" \
    --namespace "$NAMESPACE" \
    --include-crds \
    --set installCRDs=true \
    --set global.leaderElection.namespace="kube-system" \
    --set startupapicheck.enabled=false \
    --output-dir .

cp -r "$COMPONENT/templates/"* "$TEMPLATES_DIR/"

[ -d "$COMPONENT/crds" ] && {
    mkdir -p "$TEMPLATES_DIR/crds"
    cp -r "$COMPONENT/crds/"* "$TEMPLATES_DIR/crds/"
}

python3 "$SCRIPT_DIR/patch-templates.py" "$TEMPLATES_DIR" "$COMPONENT"

# Add namespace template since cert-manager chart doesn't include it
cat > "$TEMPLATES_DIR/namespace.yaml" << 'EOF'
{{- if .Values.certManager.enabled }}
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Values.global.certManagerNamespace }}
  labels:
    pod-security.kubernetes.io/enforce: restricted
{{- end }}
EOF

# Create kubeflow-issuer template
mkdir -p "$TEMPLATES_DIR/kubeflow-issuer"
cat > "$TEMPLATES_DIR/kubeflow-issuer/cluster-issuer.yaml" << 'EOF'
{{- if .Values.certManager.enabled }}
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: kubeflow-self-signing-issuer
  labels:
    {{- include "kubeflow.labels" . | nindent 4 }}
    app.kubernetes.io/component: cert-manager
    app.kubernetes.io/name: cert-manager
    kustomize.component: cert-manager
spec:
  selfSigned: {}
{{- end }}
EOF

cd "$CHART_DIR"
rm -rf "$TEMP_DIR"

helm template kubeflow . --debug --dry-run > /dev/null

echo "Cert Manager templates synchronized successfully" 