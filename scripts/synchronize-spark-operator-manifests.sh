#!/usr/bin/env bash
# This script helps to create a PR to update the Spark Operator manifests
SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"
setup_error_handling
COMPONENT_NAME="spark-operator"
REPOSITORY_NAME="kubeflow/spark-operator"
COMMIT=${COMMIT:="2.4.0"}
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${COMMIT?}}
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
DESTINATION_MANIFESTS_PATH="applications/spark/${COMPONENT_NAME}/base"
create_branch "$BRANCH_NAME"
DESTINATION_DIRECTORY=$MANIFESTS_DIRECTORY/$DESTINATION_MANIFESTS_PATH
mkdir -p $DESTINATION_DIRECTORY
cd $DESTINATION_DIRECTORY
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
--version ${COMMIT} \
--repo https://kubeflow.github.io/spark-operator > resources.yaml
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' 's/Spark Operator[^|]*|[^|]*applications\/spark\/spark-operator[^|]*|[^|]*\[[0-9]\.[0-9]\.[0-9]\]([^)]*)/Spark Operator	|	applications\/spark\/spark-operator	|	['"${COMMIT}"'](https:\/\/github.com\/kubeflow\/spark-operator\/tree\/v'"${COMMIT}"')/g' "${MANIFESTS_DIRECTORY}/README.md"
else
    sed -i 's/Spark Operator.*|.*applications\/spark\/spark-operator[^|]*|.*\[[0-9]\.[0-9]\.[0-9]\]([^)]*)/Spark Operator	|	applications\/spark\/spark-operator	|	['"${COMMIT}"'](https:\/\/github.com\/kubeflow\/spark-operator\/tree\/v'"${COMMIT}"')/g' "${MANIFESTS_DIRECTORY}/README.md"
fi
commit_changes "$MANIFESTS_DIRECTORY" "Update ${REPOSITORY_NAME} manifests from ${COMMIT}" "$MANIFESTS_DIRECTORY"
echo "Synchronization completed successfully."
