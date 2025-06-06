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
NAMESPACE="kubeflow"

rm -rf "$TEMPLATES_DIR"
mkdir -p "$TEMPLATES_DIR"

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
    mkdir -p "$TEMPLATES_DIR/crds"
    cp -r "$COMPONENT/crds/"* "$TEMPLATES_DIR/crds/"
}

python3 "$SCRIPT_DIR/patch-templates.py" "$TEMPLATES_DIR"

cd "$CHART_DIR"
rm -rf "$TEMP_DIR"

helm template kubeflow . --debug --dry-run > /dev/null 