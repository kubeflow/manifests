#!/usr/bin/env bash
# This script helps to create a PR to update the Spark Operator manifests

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

setup_error_handling

COMPONENT_NAME="spark-operator"
SPARK_OPERATOR_VERSION=${SPARK_OPERATOR_VERSION:="2.1.1"}
SPARK_OPERATOR_HELM_CHART_REPO=${SPARK_OPERATOR_HELM_CHART_REPO:="https://kubeflow.github.io/spark-operator"}
DEV_MODE=${DEV_MODE:=false}
BRANCH=${BRANCH:=synchronize-${COMPONENT_NAME}-manifests-${SPARK_OPERATOR_VERSION?}}

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)
DST_MANIFESTS_PATH="apps/spark/${COMPONENT_NAME}/base"

create_branch "$BRANCH"

echo "Generating manifests from Helm chart version ${SPARK_OPERATOR_VERSION}..."

# Generate the manifests using Helm
DST_DIR=$MANIFESTS_DIR/$DST_MANIFESTS_PATH
mkdir -p $DST_DIR
cd $DST_DIR

# Create a kustomization.yaml file if it doesn't exist
if [ ! -f kustomization.yaml ]; then
    cat > kustomization.yaml << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- resources.yaml
EOF
fi

helm template -n kubeflow --include-crds spark-operator spark-operator \
--set "spark.jobNamespaces={}" \
--set webhook.enable=true \
--set webhook.port=9443 \
--version ${SPARK_OPERATOR_VERSION} \
--repo ${SPARK_OPERATOR_HELM_CHART_REPO} > resources.yaml

echo "Successfully generated manifests."

echo "Updating README..."
# Use OS-compatible sed command
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS version
    sed -i '' 's/Spark Operator[^|]*|[^|]*apps\/spark\/spark-operator[^|]*|[^|]*[0-9]\.[0-9]\.[0-9]/Spark Operator	|	apps\/spark\/spark-operator	|	'"${SPARK_OPERATOR_VERSION}"'/g' "${MANIFESTS_DIR}/README.md"
else
    # Linux version
    sed -i 's/Spark Operator.*|.*apps\/spark\/spark-operator[^|]*|.*[0-9]\.[0-9]\.[0-9]/Spark Operator	|	apps\/spark\/spark-operator	|	'"${SPARK_OPERATOR_VERSION}"'/g' "${MANIFESTS_DIR}/README.md"
fi

commit_changes "$MANIFESTS_DIR" "Update kubeflow/${COMPONENT_NAME} manifests to ${SPARK_OPERATOR_VERSION}" \
  "apps/spark" \
  "README.md" \
  "scripts"

echo "Synchronization completed successfully."
