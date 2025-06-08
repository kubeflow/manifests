#!/usr/bin/env bash
# Only ignores helm-specific labels and annotations

set -euo pipefail

COMPONENT=${1:-"spark-operator"}
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
HELM_DIR="$ROOT_DIR/experimental/helm"
CHART_DIR="$HELM_DIR/kubeflow"

case $COMPONENT in
    "spark-operator")
        KUSTOMIZE_PATH="apps/spark/spark-operator/base"
        NAMESPACE="kubeflow"
        HELM_TEMPLATE_PATH="experimental/helm/kubeflow/templates/external/spark-operator"
        ;;
    *)
        echo "Unknown component: $COMPONENT"
        echo "Supported components: spark-operator"
        exit 1
        ;;
esac

cd "$ROOT_DIR"

if [ ! -d "$HELM_TEMPLATE_PATH" ]; then
    echo "ERROR: Helm template directory does not exist: $HELM_TEMPLATE_PATH"
    exit 1
fi

TEMPLATE_FILES=$(find "$HELM_TEMPLATE_PATH" -name "*.yaml" -o -name "*.yml" -o -name "*.tpl" | wc -l)
if [ "$TEMPLATE_FILES" -eq 0 ]; then
    echo "ERROR: No Helm template files found in $HELM_TEMPLATE_PATH"
    echo "Please implement the Helm templates for $COMPONENT before running this comparison."
    exit 1
fi

KUSTOMIZE_OUTPUT="/tmp/kustomize-${COMPONENT}.yaml"
kustomize build "$KUSTOMIZE_PATH" > "$KUSTOMIZE_OUTPUT" || {
    echo "Failed to generate Kustomize output"
    exit 1
}

cd "$CHART_DIR"
HELM_OUTPUT="/tmp/helm-aio-${COMPONENT}.yaml"
helm template kubeflow . --namespace "$NAMESPACE" --include-crds > "$HELM_OUTPUT" || {
    exit 1
}

KUBEFLOW_RBAC_ENABLED="false"
[ -f "$CHART_DIR/values.yaml" ] && {
    grep -A 10 "sparkOperator:" "$CHART_DIR/values.yaml" | grep -A 5 "kubeflowRBAC:" | grep -q "enabled: true" && KUBEFLOW_RBAC_ENABLED="true"
}

COMPARE_SCRIPT="$ROOT_DIR/tests/helm_kustomize_compare_manifests.py"
if [ -f "$COMPARE_SCRIPT" ]; then
    cd "$ROOT_DIR"
    python3 "$COMPARE_SCRIPT" "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT" "$COMPONENT" "$NAMESPACE" "$KUBEFLOW_RBAC_ENABLED" || {
        exit 1
    }
else
    KUSTOMIZE_LINES=$(wc -l < "$KUSTOMIZE_OUTPUT")
    HELM_LINES=$(wc -l < "$HELM_OUTPUT")
    
    [ -s "$KUSTOMIZE_OUTPUT" ] && [ -s "$HELM_OUTPUT" ] || {
        exit 1
    }
    
    echo "Kustomize output: $KUSTOMIZE_LINES lines"
    echo "Helm output: $HELM_LINES lines"
fi

rm -f "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT" 