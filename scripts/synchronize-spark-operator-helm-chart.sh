#!/usr/bin/env bash
# This script helps to create a PR to update the Spark Operator Helm chart

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="spark-operator"
REPOSITORY_NAME="kubeflow/spark-operator"
REPOSITORY_URL="https://github.com/kubeflow/spark-operator.git"
COMMIT=${COMMIT:="master"}
REPOSITORY_DIRECTORY="spark-operator"
SOURCE_DIRECTORY=${SOURCE_DIRECTORY:=/tmp/kubeflow-${COMPONENT_NAME}-helm}
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-helm-chart-${COMMIT}}

# Path configurations
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
SOURCE_CHART_PATH="charts/spark-operator-chart"
DESTINATION_CHART_PATH="experimental/helm/charts/${COMPONENT_NAME}"

create_branch "$BRANCH_NAME"

echo "Cloning ${REPOSITORY_NAME} repository..."
clone_and_checkout "$SOURCE_DIRECTORY" "$REPOSITORY_URL" "$REPOSITORY_DIRECTORY" "$COMMIT"

# Remove existing chart directory except for our custom values.yaml
if [ -d "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}" ]; then
    # Backup our custom values.yaml if it exists
    if [ -f "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/values.yaml" ]; then
        cp "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/values.yaml" "/tmp/spark-operator-values-backup.yaml"
    fi
    rm -rf "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}"
fi

mkdir -p "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}"
cp -r "${SOURCE_DIRECTORY}/${REPOSITORY_DIRECTORY}/${SOURCE_CHART_PATH}"/* "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/"

# Restore our custom values.yaml if it existed
if [ -f "/tmp/spark-operator-values-backup.yaml" ]; then
    echo "Restoring custom values.yaml..."
    cp "/tmp/spark-operator-values-backup.yaml" "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/values.yaml"
    rm "/tmp/spark-operator-values-backup.yaml"
else
    echo "Creating custom values.yaml..."
    cat > "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/values.yaml" << 'EOF'
# Default values for spark-operator 

# Kubeflow-specific RBAC roles for multi-tenancy
kubeflowRBAC:
  enabled: true

spark:
  # Include empty string to watch all namespaces
  jobNamespaces: 
    - ""

webhook:
  enable: true
  port: 9443
  labels:
    sidecar.istio.io/inject: "false"

controller:
  labels:
    sidecar.istio.io/inject: "false"

# Security context enhancements for Kubeflow
securityContext:
  seccompProfile:
    type: RuntimeDefault

# Namespace configuration
namespace: kubeflow
EOF
fi

# Update Chart.yaml to include our custom configurations
if [ -f "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/Chart.yaml" ]; then
    # Add a note about Kubeflow customizations
    if ! grep -q "# Kubeflow customizations" "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/Chart.yaml"; then
        echo "" >> "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/Chart.yaml"
        echo "# Kubeflow customizations applied" >> "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/Chart.yaml"
        echo "# - Custom values.yaml with Istio sidecar injection disabled" >> "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/Chart.yaml"
        echo "# - Security context enhancements" >> "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/Chart.yaml"
        echo "# - Kubeflow RBAC roles support" >> "${MANIFESTS_DIRECTORY}/${DESTINATION_CHART_PATH}/Chart.yaml"
    fi
fi

rm -rf "$SOURCE_DIRECTORY"

commit_changes "$MANIFESTS_DIRECTORY" "Update ${COMPONENT_NAME} Helm chart from upstream ${COMMIT}" \
  "experimental/helm/charts/${COMPONENT_NAME}" \
  "README.md"

echo "Updated Helm chart at: ${DESTINATION_CHART_PATH}" 