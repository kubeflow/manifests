#!/usr/bin/env bash
# This script helps to create a PR to update the OAuth2-Proxy manifests

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="oauth2-proxy"
OAUTH2_PROXY_RELEASE="v7.10.0" # Must be a release
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${OAUTH2_PROXY_RELEASE?}}

MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
DESTINATION_DIRECTORY=$MANIFESTS_DIRECTORY/common/${COMPONENT_NAME}

create_branch "$BRANCH_NAME"

check_uncommitted_changes

echo "Updating OAuth2-Proxy image tag to ${OAUTH2_PROXY_RELEASE}..."

if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i "" "s|quay.io/oauth2-proxy/oauth2-proxy:.*|quay.io/oauth2-proxy/oauth2-proxy:${OAUTH2_PROXY_RELEASE}|g" \
        $DESTINATION_DIRECTORY/base/deployment.yaml
else
    sed -i "s|quay.io/oauth2-proxy/oauth2-proxy:.*|quay.io/oauth2-proxy/oauth2-proxy:${OAUTH2_PROXY_RELEASE}|g" \
        $DESTINATION_DIRECTORY/base/deployment.yaml
fi

if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i "" '/| OAuth2-Proxy | common\/oauth2-proxy |/s|\[.*\](https://github.com/oauth2-proxy/oauth2-proxy/releases/tag/v.*)|['"${OAUTH2_PROXY_RELEASE#v}"'](https://github.com/oauth2-proxy/oauth2-proxy/releases/tag/'"${OAUTH2_PROXY_RELEASE}"')|' \
        ${MANIFESTS_DIRECTORY}/README.md
else
    sed -i '/| OAuth2-Proxy | common\/oauth2-proxy |/s|\[.*\](https://github.com/oauth2-proxy/oauth2-proxy/releases/tag/v.*)|['"${OAUTH2_PROXY_RELEASE#v}"'](https://github.com/oauth2-proxy/oauth2-proxy/releases/tag/'"${OAUTH2_PROXY_RELEASE}"')|' \
        ${MANIFESTS_DIRECTORY}/README.md
fi

commit_changes "$MANIFESTS_DIRECTORY" "Update common/oauth2-proxy manifests to ${OAUTH2_PROXY_RELEASE}" \
  "$DESTINATION_DIRECTORY" \
  "README.md"

echo "Synchronization completed successfully." 