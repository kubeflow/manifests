#!/usr/bin/env bash
# This script helps to create a PR to update the Model Registry manifests

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="model-registry"
REPOSITORY_NAME="kubeflow/model-registry"
REPOSITORY_URL="https://github.com/kubeflow/model-registry.git"
COMMIT="v0.2.13"
REPOSITORY_DIRECTORY="model-registry"
DEV_MODE=${DEV_MODE:=false}
SOURCE_DIRECTORY=${SOURCE_DIRECTORY:=/tmp/kubeflow-${COMPONENT_NAME}}
BRANCH_NAME=${BRANCH_NAME:=synchronize-kubeflow-${COMPONENT_NAME}-manifests-${COMMIT?}}

# Path configurations
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
SOURCE_MANIFESTS_PATH="manifests/kustomize"
DESTINATION_MANIFESTS_PATH="applications/${COMPONENT_NAME}/upstream"

# README update patterns
SOURCE_TEXT="\[.*\](https://github.com/${REPOSITORY_NAME}/tree/.*/manifests/kustomize)"
DESTINATION_TEXT="\[${COMMIT}\](https://github.com/${REPOSITORY_NAME}/tree/${COMMIT}/manifests/kustomize)"

create_branch "$BRANCH_NAME"

clone_and_checkout "$SOURCE_DIRECTORY" "$REPOSITORY_URL" "$REPOSITORY_DIRECTORY" "$COMMIT"

echo "Copying ${COMPONENT_NAME} manifests..."
DESTINATION_DIRECTORY=$MANIFESTS_DIRECTORY/$DESTINATION_MANIFESTS_PATH
if [ -d "$DESTINATION_DIRECTORY" ]; then
    rm -r "$DESTINATION_DIRECTORY"
fi
mkdir -p $DESTINATION_DIRECTORY
cp -r "$SOURCE_DIRECTORY/$REPOSITORY_DIRECTORY/$SOURCE_MANIFESTS_PATH/"* "$DESTINATION_DIRECTORY"

echo "Successfully copied all manifests."

update_readme "$MANIFESTS_DIRECTORY" "$SOURCE_TEXT" "$DESTINATION_TEXT"

commit_changes "$MANIFESTS_DIRECTORY" "Update ${REPOSITORY_NAME} manifests from ${COMMIT}" \
  "applications" \
  "README.md"

echo "Synchronization completed successfully."
