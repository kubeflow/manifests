#!/usr/bin/env bash
# Script to sync Spark Operator templates for AIO Helm chart

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
HELM_DIR="$(dirname "$SCRIPT_DIR")"
CHART_DIR="$HELM_DIR/kubeflow"

COMPONENT="spark-operator"
VERSION="2.2.0"
REPO="https://kubeflow.github.io/spark-operator"
TEMPLATES_DIR="$CHART_DIR/templates/external/${COMPONENT}"
CRDS_DIR="$CHART_DIR/crds"
NAMESPACE="kubeflow"

rm -rf "$TEMPLATES_DIR"
mkdir -p "$TEMPLATES_DIR"
mkdir -p "$CRDS_DIR"

TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# Generate templates using same settings as existing Kustomize sync
helm template "$COMPONENT" "$COMPONENT" \
    --version "$VERSION" \
    --repo "$REPO" \
    --namespace "$NAMESPACE" \
    --include-crds \
    --set "spark.jobNamespaces={}" \
    --set webhook.enable=true \
    --set webhook.port=9443 \
    --output-dir .

cp -r "$COMPONENT/templates/"* "$TEMPLATES_DIR/"

[ -d "$COMPONENT/crds" ] && {
    cp -r "$COMPONENT/crds/"* "$CRDS_DIR/"
}

python3 "$SCRIPT_DIR/patch-templates.py" "$TEMPLATES_DIR" "$COMPONENT"

cd "$CHART_DIR"
rm -rf "$TEMP_DIR"

helm template kubeflow . --debug --dry-run > /dev/null 