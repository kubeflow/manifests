#!/usr/bin/env bash
# This script helps to create a PR to update the Katib manifests

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="katib"
REPOSITORY_NAME="kubeflow/katib"
REPOSITORY_URL="https://github.com/kubeflow/katib.git"
COMMIT="v0.19.0"
REPOSITORY_DIRECTORY="katib"
SOURCE_DIRECTORY=${SOURCE_DIRECTORY:=/tmp/kubeflow-${COMPONENT_NAME}}
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${COMMIT?}}

# Path configurations
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
SOURCE_MANIFESTS_PATH="manifests/v1beta1"
DESTINATION_MANIFESTS_PATH="applications/${COMPONENT_NAME}/upstream"

# README update patterns
SOURCE_TEXT="\[.*\](https://github.com/${REPOSITORY_NAME}/tree/.*/manifests/v1beta1)"
DESTINATION_TEXT="\[${COMMIT}\](https://github.com/${REPOSITORY_NAME}/tree/${COMMIT}/manifests/v1beta1)"

create_branch "$BRANCH_NAME"

clone_and_checkout "$SOURCE_DIRECTORY" "$REPOSITORY_URL" "$REPOSITORY_DIRECTORY" "$COMMIT"

copy_manifests "${SOURCE_DIRECTORY}/${REPOSITORY_DIRECTORY}/${SOURCE_MANIFESTS_PATH}" "${MANIFESTS_DIRECTORY}/${DESTINATION_MANIFESTS_PATH}"

update_readme "$MANIFESTS_DIRECTORY" "$SOURCE_TEXT" "$DESTINATION_TEXT"

commit_changes "$MANIFESTS_DIRECTORY" "Update ${REPOSITORY_NAME} manifests from ${COMMIT}" \
  "applications" \
  "README.md"

echo "Synchronization completed successfully."
