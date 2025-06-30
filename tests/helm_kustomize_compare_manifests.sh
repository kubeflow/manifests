#!/usr/bin/env bash
# Compare Helm vs Kustomize manifests for Kubeflow components

set -euo pipefail

COMPONENT=${1:-"cert-manager"}
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
HELM_DIR="$ROOT_DIR/experimental/helm"

case $COMPONENT in
    "spark-operator")
        KUSTOMIZE_PATH="applications/spark/spark-operator/base"
        NAMESPACE="kubeflow"
        ;;
    "cert-manager")
        KUSTOMIZE_PATH="common/cert-manager/base"
        NAMESPACE="cert-manager"
        ;;
    *)
        echo "ERROR: Unsupported component: $COMPONENT"
        echo "Only 'spark-operator' and 'cert-manager' are supported."
        echo "Centraldashboard support has been removed in favor of dynamic generation."
        exit 1
        ;;
esac

cd "$ROOT_DIR"

# Check if we're running in an environment where Helm templates should be generated
# (like GitHub Actions) vs using existing templates
HELM_CHART_DIR="$HELM_DIR/kubeflow"
if [ ! -d "$HELM_CHART_DIR" ]; then
    echo "ERROR: Helm chart directory not found: $HELM_CHART_DIR"
    echo "This script expects Helm templates to be generated first."
    echo "Run the appropriate synchronization script from experimental/helm/scripts/ first."
    exit 1
fi

HELM_TEMPLATE_PATH="$HELM_CHART_DIR/templates/external/$COMPONENT"
if [ ! -d "$HELM_TEMPLATE_PATH" ]; then
    echo "ERROR: Helm template directory does not exist: $HELM_TEMPLATE_PATH"
    echo "Run: experimental/helm/scripts/synchronize-${COMPONENT}.sh"
    exit 1
fi

TEMPLATE_FILES=$(find "$HELM_TEMPLATE_PATH" -name "*.yaml" -o -name "*.yml" -o -name "*.tpl" | wc -l)
if [ "$TEMPLATE_FILES" -eq 0 ]; then
    echo "ERROR: No Helm template files found in $HELM_TEMPLATE_PATH"
    echo "Please run the synchronization script: experimental/helm/scripts/synchronize-${COMPONENT}.sh"
    exit 1
fi

echo "Generating Kustomize manifests for $COMPONENT..."
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

echo "Generating Helm manifests for $COMPONENT..."
cd "$HELM_CHART_DIR"
HELM_OUTPUT="/tmp/helm-aio-${COMPONENT}.yaml"
TEMP_VALUES_FILE="/tmp/test-values-${COMPONENT}.yaml"

  # Create values file with only the target component enabled
create_values_for_component() {
    local component=$1
    
    cat > "$TEMP_VALUES_FILE" << EOF
# Global settings
global:
  kubeflowNamespace: kubeflow
  certManagerNamespace: cert-manager

# Disable all components by default
sparkOperator:
  enabled: false
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
    
    case $component in
        "cert-manager")
            cat >> "$TEMP_VALUES_FILE" << EOF

# Enable cert-manager
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
EOF
            ;;
        "spark-operator")
            cat >> "$TEMP_VALUES_FILE" << EOF

# Enable spark-operator
sparkOperator:
  enabled: true
  spark:
    jobNamespaces: []
  webhook:
    enable: true
    port: 9443
  kubeflowRBAC:
    enabled: true
EOF
            ;;
    esac
}

create_values_for_component "$COMPONENT"

helm template kubeflow . --namespace "$NAMESPACE" --include-crds --values "$TEMP_VALUES_FILE" > "$HELM_OUTPUT"
rm -f "$TEMP_VALUES_FILE"

KUBEFLOW_RBAC_ENABLED="false"
if [ "$COMPONENT" = "spark-operator" ]; then
    INTEGRATIONS_FILE="$HELM_CHART_DIR/templates/integrations/spark-operator-rbac.yaml"
    if [ -f "$INTEGRATIONS_FILE" ]; then
        KUBEFLOW_RBAC_ENABLED="true"
    fi
fi

echo "Comparing manifests for $COMPONENT..."
cd "$ROOT_DIR"
python3 "$ROOT_DIR/tests/helm_kustomize_compare_manifests.py" "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT" "$COMPONENT" "$NAMESPACE" "$KUBEFLOW_RBAC_ENABLED"

rm -f "$KUSTOMIZE_OUTPUT" "$HELM_OUTPUT"

echo "Comparison completed for $COMPONENT" 