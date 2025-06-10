#!/usr/bin/env bash
# Compare Helm vs Kustomize manifests for Kubeflow components

set -euo pipefail

COMPONENT=${1:-"cert-manager"}
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
    "cert-manager")
        KUSTOMIZE_PATH="common/cert-manager/base"
        NAMESPACE="cert-manager"
        HELM_TEMPLATE_PATH="experimental/helm/kubeflow/templates/external/cert-manager"
        ;;
    *)
        echo "Unknown component: $COMPONENT"
        echo "Supported components: spark-operator, cert-manager"
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
case $COMPONENT in
    "cert-manager")
        {
            kustomize build "$KUSTOMIZE_PATH"
            kustomize build "common/cert-manager/kubeflow-issuer/base"
        } > "$KUSTOMIZE_OUTPUT"
        ;;
    *)
        kustomize build "$KUSTOMIZE_PATH" > "$KUSTOMIZE_OUTPUT"
        ;;
esac

cd "$CHART_DIR"
HELM_OUTPUT="/tmp/helm-aio-${COMPONENT}.yaml"
TEMP_VALUES_FILE="/tmp/test-values-${COMPONENT}.yaml"

case $COMPONENT in
    "cert-manager")
        cat > "$TEMP_VALUES_FILE" << EOF
sparkOperator:
  enabled: false
certManager:
  enabled: true
  installCRDs: true
  global:
    leaderElection:
      namespace: kube-system
  startupapicheck:
    enabled: false
  kubeflowIssuer:
    enabled: true
    name: kubeflow-self-signing-issuer
trainingOperator:
  enabled: false
istio:
  enabled: false
oauth2Proxy:
  enabled: false
dex:
  enabled: false
centraldashboard:
  enabled: false
profiles:
  enabled: false
jupyter:
  enabled: false
pipelines:
  enabled: false
kserve:
  enabled: false
katib:
  enabled: false
tensorboard:
  enabled: false
volumesWebApp:
  enabled: false
admissionWebhook:
  enabled: false
pvcviewerController:
  enabled: false
modelRegistry:
  enabled: false
EOF
        ;;
    "spark-operator")
        cat > "$TEMP_VALUES_FILE" << EOF
sparkOperator:
  enabled: true
  spark:
    jobNamespaces: []
  webhook:
    enable: true
    port: 9443
  kubeflowRBAC:
    enabled: true
certManager:
  enabled: false
trainingOperator:
  enabled: false
istio:
  enabled: false
oauth2Proxy:
  enabled: false
dex:
  enabled: false
centraldashboard:
  enabled: false
profiles:
  enabled: false
jupyter:
  enabled: false
pipelines:
  enabled: false
kserve:
  enabled: false
katib:
  enabled: false
tensorboard:
  enabled: false
volumesWebApp:
  enabled: false
admissionWebhook:
  enabled: false
pvcviewerController:
  enabled: false
modelRegistry:
  enabled: false
EOF
        ;;
esac

helm template kubeflow . --namespace "$NAMESPACE" --include-crds --values "$TEMP_VALUES_FILE" > "$HELM_OUTPUT"
rm -f "$TEMP_VALUES_FILE"

KUBEFLOW_RBAC_ENABLED="false"
if [ "$COMPONENT" = "spark-operator" ] && [ -f "$CHART_DIR/values.yaml" ]; then
    grep -A 10 "sparkOperator:" "$CHART_DIR/values.yaml" | grep -A 5 "kubeflowRBAC:" | grep -q "enabled: true" && KUBEFLOW_RBAC_ENABLED="true"
fi

cd "$ROOT_DIR"
python3 "$ROOT_DIR/tests/helm_kustomize_compare_manifests.py" "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT" "$COMPONENT" "$NAMESPACE" "$KUBEFLOW_RBAC_ENABLED"

rm -f "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT" 