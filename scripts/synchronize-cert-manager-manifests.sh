#!/usr/bin/env bash
# This script helps to create a PR to update cert-manager manifests.

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="cert-manager"
REPOSITORY_NAME="cert-manager/cert-manager"
COMMIT="v1.19.4" # Must be a release tag in cert-manager/cert-manager
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${COMMIT?}}

MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
DESTINATION_DIRECTORY=$MANIFESTS_DIRECTORY/common/${COMPONENT_NAME}
DESTINATION_FILE="$DESTINATION_DIRECTORY/base/upstream/cert-manager.yaml"

create_branch "$BRANCH_NAME"
check_uncommitted_changes

echo "Downloading ${COMPONENT_NAME} manifest ${COMMIT}..."
wget -O "$DESTINATION_FILE" \
  "https://github.com/${REPOSITORY_NAME}/releases/download/${COMMIT}/cert-manager.yaml"

SOURCE_TEXT="\[.*\](https://github.com/${REPOSITORY_NAME}/releases/tag/v.*)"
DESTINATION_TEXT="\[${COMMIT#v}\](https://github.com/${REPOSITORY_NAME}/releases/tag/${COMMIT})"
update_readme "$MANIFESTS_DIRECTORY" "$SOURCE_TEXT" "$DESTINATION_TEXT"

commit_changes "$MANIFESTS_DIRECTORY" "Update common/${COMPONENT_NAME} manifests to ${COMMIT}" \
  "$DESTINATION_DIRECTORY" \
  "README.md"

echo "Synchronization completed successfully."
