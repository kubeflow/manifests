#!/usr/bin/env bash
# This script helps to create a PR to update the Spark Operator manifests

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="spark-operator"
SPARK_OPERATOR_VERSION=${SPARK_OPERATOR_VERSION:="2.4.0"}
SPARK_OPERATOR_HELM_CHART_REPO=${SPARK_OPERATOR_HELM_CHART_REPO:="https://kubeflow.github.io/spark-operator"}
DEV_MODE=${DEV_MODE:=false}
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${SPARK_OPERATOR_VERSION?}}

# Path configurations
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
DESTINATION_MANIFESTS_PATH="applications/spark/${COMPONENT_NAME}/base"

create_branch "$BRANCH_NAME"

echo "Generating manifests from Helm chart version ${SPARK_OPERATOR_VERSION}..."

DESTINATION_DIRECTORY=$MANIFESTS_DIRECTORY/$DESTINATION_MANIFESTS_PATH
mkdir -p $DESTINATION_DIRECTORY
cd $DESTINATION_DIRECTORY

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

# Use OS-compatible sed command
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' 's/Spark Operator[^|]*|[^|]*applications\/spark\/spark-operator[^|]*|[^|]*\[[0-9]\.[0-9]\.[0-9]\]([^)]*)/Spark Operator	|	applications\/spark\/spark-operator	|	['"${SPARK_OPERATOR_VERSION}"'](https:\/\/github.com\/kubeflow\/spark-operator\/tree\/v'"${SPARK_OPERATOR_VERSION}"')/g' "${MANIFESTS_DIRECTORY}/README.md"
else
    sed -i 's/Spark Operator.*|.*applications\/spark\/spark-operator[^|]*|.*\[[0-9]\.[0-9]\.[0-9]\]([^)]*)/Spark Operator	|	applications\/spark\/spark-operator	|	['"${SPARK_OPERATOR_VERSION}"'](https:\/\/github.com\/kubeflow\/spark-operator\/tree\/v'"${SPARK_OPERATOR_VERSION}"')/g' "${MANIFESTS_DIRECTORY}/README.md"
fi

commit_changes "$MANIFESTS_DIRECTORY" "Update kubeflow/${COMPONENT_NAME} manifests to ${SPARK_OPERATOR_VERSION}" \
  "applications/spark" \
  "README.md" \
  "scripts"

echo "Synchronization completed successfully."
