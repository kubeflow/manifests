#!/usr/bin/env bash
# This script helps to create a PR to update the Dex manifests
SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"
setup_error_handling
COMPONENT_NAME="dex"
REPOSITORY_NAME="dexidp/dex"
COMMIT="v2.45.0"
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${COMMIT?}}
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
DESTINATION_DIRECTORY=$MANIFESTS_DIRECTORY/common/${COMPONENT_NAME}
create_branch "$BRANCH_NAME"
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i "" "s|ghcr.io/dexidp/dex:v[0-9.]*|ghcr.io/dexidp/dex:${COMMIT}|g" \
        $DESTINATION_DIRECTORY/base/deployment.yaml
else
    sed -i "s|ghcr.io/dexidp/dex:v[0-9.]*|ghcr.io/dexidp/dex:${COMMIT}|g" \
        $DESTINATION_DIRECTORY/base/deployment.yaml
fi
SOURCE_TEXT="\[.*\](https://github.com/${REPOSITORY_NAME}/releases/tag/v.*)"
DESTINATION_TEXT="\[${COMMIT#v}\](https://github.com/${REPOSITORY_NAME}/releases/tag/${COMMIT})"
update_readme "$MANIFESTS_DIRECTORY" "$SOURCE_TEXT" "$DESTINATION_TEXT"
commit_changes "$MANIFESTS_DIRECTORY" "Update ${REPOSITORY_NAME} manifests from ${COMMIT}" "$MANIFESTS_DIRECTORY"
echo "Synchronization completed successfully."
