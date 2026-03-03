#!/usr/bin/env bash
# This script helps to create a PR to update the Spark Operator manifests
SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"
setup_error_handling
COMPONENT_NAME="spark-operator"
REPOSITORY_NAME="kubeflow/spark-operator"
REPOSITORY_URL="https://github.com/kubeflow/spark-operator.git"
COMMIT="v2.5.0-rc.0"
REPOSITORY_DIRECTORY="${COMPONENT_NAME}"
SOURCE_DIRECTORY=${SOURCE_DIRECTORY:=/tmp/kubeflow-${COMPONENT_NAME}}
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${COMMIT?}}
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
DESTINATION_MANIFESTS_PATH="applications/spark/${COMPONENT_NAME}/base"
SOURCE_TEXT="\[[^]]*\](https://github.com/${REPOSITORY_NAME}/tree/[^)]*)"
DESTINATION_TEXT="\[${COMMIT#v}\](https://github.com/${REPOSITORY_NAME}/tree/${COMMIT})"
create_branch "$BRANCH_NAME"
clone_and_checkout "$SOURCE_DIRECTORY" "$REPOSITORY_URL" "$REPOSITORY_DIRECTORY" "$COMMIT"
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
helm template -n kubeflow --include-crds spark-operator \
--set "spark.jobNamespaces={}" \
--set webhook.enable=true \
--set webhook.port=9443 \
"${SOURCE_DIRECTORY}/${REPOSITORY_DIRECTORY}/charts/spark-operator-chart" > resources.yaml
update_readme "$MANIFESTS_DIRECTORY" "$SOURCE_TEXT" "$DESTINATION_TEXT"
commit_changes "$MANIFESTS_DIRECTORY" "Update ${REPOSITORY_NAME} manifests from ${COMMIT}" "$MANIFESTS_DIRECTORY"
echo "Synchronization completed successfully."
