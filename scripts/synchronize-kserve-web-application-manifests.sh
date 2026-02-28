#!/usr/bin/env bash
# This script helps to create a PR to update the KServe Models Web App manifests

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="models-web-app"
REPOSITORY_NAME="kserve/models-web-app"
REPOSITORY_URL="https://github.com/kserve/models-web-app.git"
COMMIT="v0.15.0"
REPOSITORY_DIRECTORY="models-web-app"
SOURCE_DIRECTORY=${SOURCE_DIRECTORY:=/tmp/kserve-${COMPONENT_NAME}}
BRANCH_NAME=${BRANCH_NAME:=synchronize-kserve-${COMPONENT_NAME}-manifests-${COMMIT?}}

# Path configurations
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
SOURCE_MANIFESTS_PATH="config"
DESTINATION_MANIFESTS_PATH="applications/kserve/${COMPONENT_NAME}"

# README update patterns
SOURCE_TEXT="\[.*\](https://github.com/${REPOSITORY_NAME}/tree/.*)"
DESTINATION_TEXT="\[${COMMIT}\](https://github.com/${REPOSITORY_NAME}/tree/${COMMIT}/${SOURCE_MANIFESTS_PATH})"

create_branch "$BRANCH_NAME"

clone_and_checkout "$SOURCE_DIRECTORY" "$REPOSITORY_URL" "$REPOSITORY_DIRECTORY" "$COMMIT"

echo "Copying manifests"
copy_manifests "${SOURCE_DIRECTORY}/${REPOSITORY_DIRECTORY}/${SOURCE_MANIFESTS_PATH}" "${MANIFESTS_DIRECTORY}/${DESTINATION_MANIFESTS_PATH}"

update_readme "$MANIFESTS_DIRECTORY" "$SOURCE_TEXT" "$DESTINATION_TEXT"

commit_changes "$MANIFESTS_DIRECTORY" "Update kserve models web application manifests from ${COMMIT}" \
  "${DESTINATION_MANIFESTS_PATH}" \
  "README.md"

echo "Synchronization completed successfully."
