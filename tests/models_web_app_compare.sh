#!/usr/bin/env bash
# Compare Helm vs Kustomize manifests for KServe Models Web App

set -euo pipefail

SCENARIO=${1:-"base"}
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
CHART_DIR="$ROOT_DIR/experimental/helm/charts/kserve-models-web-app"
MANIFESTS_DIR="$ROOT_DIR/applications/kserve/models-web-app"

declare -A KUSTOMIZE_PATHS=(
    ["base"]="$MANIFESTS_DIR/base"
    ["kubeflow"]="$MANIFESTS_DIR/overlays/kubeflow"
)

declare -A HELM_VALUES=(
    ["base"]=""  
    ["kubeflow"]="--set kubeflow.enabled=true --set global.namespace=kubeflow"
)

declare -A NAMESPACES=(
    ["base"]="kserve"
    ["kubeflow"]="kubeflow"
)

if [[ ! "${KUSTOMIZE_PATHS[$SCENARIO]:-}" ]]; then
    echo "ERROR: Unknown scenario: $SCENARIO"
    echo "Available scenarios:"
    for scenario in "${!KUSTOMIZE_PATHS[@]}"; do
        echo "  - $scenario"
    done
    exit 1
fi

KUSTOMIZE_PATH="${KUSTOMIZE_PATHS[$SCENARIO]}"
HELM_VALUES_ARGS="${HELM_VALUES[$SCENARIO]}"
NAMESPACE="${NAMESPACES[$SCENARIO]}"

echo "Comparing KServe Models Web App manifests for scenario: $SCENARIO"


if [ ! -d "$KUSTOMIZE_PATH" ]; then
    exit 1
fi

if [ ! -d "$CHART_DIR" ]; then
    exit 1
fi



KUSTOMIZE_OUTPUT="/tmp/kustomize-models-web-app-${SCENARIO}.yaml"
HELM_OUTPUT="/tmp/helm-models-web-app-${SCENARIO}.yaml"

cd "$ROOT_DIR"
kustomize build "$KUSTOMIZE_PATH" > "$KUSTOMIZE_OUTPUT"

cd "$ROOT_DIR"
if [ -n "$HELM_VALUES_ARGS" ]; then

    eval "helm template kserve-models-web-app $CHART_DIR --namespace $NAMESPACE $HELM_VALUES_ARGS > $HELM_OUTPUT"
else
    helm template kserve-models-web-app "$CHART_DIR" --namespace "$NAMESPACE" > "$HELM_OUTPUT"
fi

cd "$ROOT_DIR"
python3 "$SCRIPT_DIR/models_web_app_compare_manifests.py" \
    "$KUSTOMIZE_OUTPUT" \
    "$HELM_OUTPUT" \
    "$SCENARIO"

COMPARISON_RESULT=$?

rm -f "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT"

if [ $COMPARISON_RESULT -eq 0 ]; then
    echo "SUCCESS: Manifests are equivalent for scenario '$SCENARIO'"
else
    echo "FAILED: Manifests are NOT equivalent for scenario '$SCENARIO'"
fi

exit $COMPARISON_RESULT 