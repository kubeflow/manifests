#!/usr/bin/env bash
set -euo pipefail

COMPONENT=${1:-spark-operator}
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"

case $COMPONENT in
    "spark-operator")
        KUSTOMIZE_PATH="apps/spark/spark-operator/base"
        HELM_PATH="experimental/helm/charts/spark-operator"
        NAMESPACE="kubeflow"
        ;;
    *)
        exit 1
        ;;
esac

cd "$ROOT_DIR"
KUSTOMIZE_OUTPUT="/tmp/kustomize-${COMPONENT}.yaml"
kustomize build "$KUSTOMIZE_PATH" > "$KUSTOMIZE_OUTPUT"

cd "$ROOT_DIR/$HELM_PATH"
HELM_OUTPUT="/tmp/helm-${COMPONENT}.yaml"

helm dependency update
helm template "$COMPONENT" . --namespace "$NAMESPACE" --include-crds > "$HELM_OUTPUT"

KUBEFLOW_RBAC_ENABLED=$(grep -A4 "kubeflowRBAC:" values.yaml | grep "enabled:" | awk '{print $2}' || echo "false")

cd "$ROOT_DIR"
python3 "$SCRIPT_DIR/helm_kustomize_compare_manifests.py" "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT" "$COMPONENT" "$NAMESPACE" "$KUBEFLOW_RBAC_ENABLED"

rm -f "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT"