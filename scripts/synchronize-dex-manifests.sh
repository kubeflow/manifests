#!/usr/bin/env bash
# This script helps to create a PR to update the Dex manifests

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="dex"
DEX_RELEASE="v2.43.1" # Must be a release
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${DEX_RELEASE?}}

MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
DESTINATION_DIRECTORY=$MANIFESTS_DIRECTORY/common/${COMPONENT_NAME}

create_branch "$BRANCH_NAME"

check_uncommitted_changes

echo "Updating Dex image tag to ${DEX_RELEASE}..."

if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i "" "s|ghcr.io/dexidp/dex:v[0-9.]*|ghcr.io/dexidp/dex:${DEX_RELEASE}|g" \
        $DESTINATION_DIRECTORY/base/deployment.yaml
else
    sed -i "s|ghcr.io/dexidp/dex:v[0-9.]*|ghcr.io/dexidp/dex:${DEX_RELEASE}|g" \
        $DESTINATION_DIRECTORY/base/deployment.yaml
fi

if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i "" '/| Dex | common\/dex |/s|\[.*\](https://github.com/dexidp/dex/releases/tag/v.*)|['"${DEX_RELEASE#v}"'](https://github.com/dexidp/dex/releases/tag/'"${DEX_RELEASE}"')|' \
        ${MANIFESTS_DIRECTORY}/README.md
else
    sed -i '/| Dex | common\/dex |/s|\[.*\](https://github.com/dexidp/dex/releases/tag/v.*)|['"${DEX_RELEASE#v}"'](https://github.com/dexidp/dex/releases/tag/'"${DEX_RELEASE}"')|' \
        ${MANIFESTS_DIRECTORY}/README.md
fi

commit_changes "$MANIFESTS_DIRECTORY" "Update common/dex manifests to ${DEX_RELEASE}" \
  "$DESTINATION_DIRECTORY" \
  "README.md"

echo "Synchronization completed successfully." 