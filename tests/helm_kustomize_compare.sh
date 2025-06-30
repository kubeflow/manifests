#!/usr/bin/env bash
# Compare Helm vs Kustomize manifests for particular Model Registry components

set -euo pipefail

SCENARIO=${1:-"base"}
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
CHART_DIR="$ROOT_DIR/experimental/helm/charts/model-registry"
MANIFESTS_DIR="$ROOT_DIR/applications/model-registry/upstream"

declare -A KUSTOMIZE_PATHS=(
    ["base"]="$MANIFESTS_DIR/base"
    ["overlay-postgres"]="$MANIFESTS_DIR/overlays/postgres"
    ["overlay-db"]="$MANIFESTS_DIR/overlays/db"
    ["controller-manager"]="$MANIFESTS_DIR/options/controller/manager"
    ["controller-rbac"]="$MANIFESTS_DIR/options/controller/rbac" 
    ["controller-default"]="$MANIFESTS_DIR/options/controller/default"
    ["controller-prometheus"]="$MANIFESTS_DIR/options/controller/prometheus"
    ["controller-network-policy"]="$MANIFESTS_DIR/options/controller/network-policy"
    ["ui-base"]="$MANIFESTS_DIR/options/ui/base"
    ["ui-standalone"]="$MANIFESTS_DIR/options/ui/overlays/standalone"
    ["ui-integrated"]="$MANIFESTS_DIR/options/ui/overlays/integrated"
    ["ui-istio"]="$MANIFESTS_DIR/options/ui/overlays/istio"
    ["istio"]="$MANIFESTS_DIR/options/istio"
    ["csi"]="$MANIFESTS_DIR/options/csi"
)

declare -A HELM_VALUES=(
    ["base"]="$CHART_DIR/ci/ci-values.yaml"
    ["overlay-postgres"]="$CHART_DIR/ci/values-postgres.yaml"
    ["overlay-db"]="$CHART_DIR/ci/values-db.yaml"
    ["controller-manager"]="$CHART_DIR/ci/values-controller-manager.yaml"
    ["controller-rbac"]="$CHART_DIR/ci/values-controller-rbac.yaml"
    ["controller-default"]="$CHART_DIR/ci/values-controller.yaml"
    ["controller-prometheus"]="$CHART_DIR/ci/values-controller-prometheus.yaml"
    ["controller-network-policy"]="$CHART_DIR/ci/values-controller-network-policy.yaml"
    ["ui-base"]="$CHART_DIR/ci/values-ui.yaml"
    ["ui-standalone"]="$CHART_DIR/ci/values-ui-standalone.yaml"
    ["ui-integrated"]="$CHART_DIR/ci/values-ui-integrated.yaml"
    ["ui-istio"]="$CHART_DIR/ci/values-ui-istio.yaml"
    ["istio"]="$CHART_DIR/ci/values-istio.yaml"
    ["csi"]="$CHART_DIR/ci/values-csi.yaml"
)

declare -A NAMESPACES=(
    ["base"]="kubeflow"
    ["overlay-postgres"]="kubeflow"
    ["overlay-db"]="kubeflow"
    ["controller-manager"]="kubeflow"
    ["controller-rbac"]="kubeflow"
    ["controller-default"]="kubeflow"
    ["controller-prometheus"]="kubeflow"
    ["controller-network-policy"]="kubeflow"
    ["ui-base"]="kubeflow"
    ["ui-standalone"]="kubeflow"
    ["ui-integrated"]="kubeflow"
    ["ui-istio"]="kubeflow"
    ["istio"]="kubeflow"
    ["csi"]="kubeflow"
)

if [[ ! "${KUSTOMIZE_PATHS[$SCENARIO]:-}" ]]; then
    echo "ERROR: Unknown scenario: $SCENARIO"
    for scenario in "${!KUSTOMIZE_PATHS[@]}"; do
        echo "  - $scenario"
    done
    exit 1
fi

KUSTOMIZE_PATH="${KUSTOMIZE_PATHS[$SCENARIO]}"
HELM_VALUES_FILE="${HELM_VALUES[$SCENARIO]}"
NAMESPACE="${NAMESPACES[$SCENARIO]}"

echo "Comparing Model Registry manifests for scenario: $SCENARIO"

if [ ! -d "$KUSTOMIZE_PATH" ]; then
    exit 1
fi

if [ ! -f "$HELM_VALUES_FILE" ]; then
    exit 1
fi

if [ ! -d "$CHART_DIR" ]; then
    exit 1
fi

KUSTOMIZE_OUTPUT="/tmp/kustomize-model-registry-${SCENARIO}.yaml"
HELM_OUTPUT="/tmp/helm-model-registry-${SCENARIO}.yaml"

cd "$ROOT_DIR"
kustomize build "$KUSTOMIZE_PATH" > "$KUSTOMIZE_OUTPUT"

cd "$CHART_DIR"
helm template model-registry . \
    --namespace "$NAMESPACE" \
    --include-crds \
    --values "$HELM_VALUES_FILE" > "$HELM_OUTPUT"

cd "$ROOT_DIR"
python3 "$SCRIPT_DIR/helm_compare_manifests.py" \
    "$KUSTOMIZE_OUTPUT" \
    "$HELM_OUTPUT" \
    "$SCENARIO" \
    "$NAMESPACE" \
    ${VERBOSE:+--verbose}

COMPARISON_RESULT=$?

rm -f "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT"

if [ $COMPARISON_RESULT -eq 0 ]; then
    echo "SUCCESS: Manifests are equivalent for scenario '$SCENARIO'"
else
    echo "FAILED: Manifests are NOT equivalent for scenario '$SCENARIO'"
fi

exit $COMPARISON_RESULT 