#!/usr/bin/env bash
# This script helps to create a PR to update cert-manager manifests.

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="cert-manager"
CERT_MANAGER_RELEASE="v1.19.4" # Must be a release tag in cert-manager/cert-manager
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${CERT_MANAGER_RELEASE?}}

MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
DESTINATION_DIRECTORY=$MANIFESTS_DIRECTORY/common/${COMPONENT_NAME}
DESTINATION_FILE="$DESTINATION_DIRECTORY/base/upstream/cert-manager.yaml"

create_branch "$BRANCH_NAME"
check_uncommitted_changes

echo "Downloading cert-manager manifest ${CERT_MANAGER_RELEASE}..."
wget -O "$DESTINATION_FILE" \
  "https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_RELEASE}/cert-manager.yaml"

# Update top-level component version table.
if [[ "$OSTYPE" == "darwin"* ]]; then
  sed -i "" '/| Cert Manager | common\/cert-manager |/s|\[.*\](https://github.com/cert-manager/cert-manager/releases/tag/v.*)|['"${CERT_MANAGER_RELEASE#v}"'](https://github.com/cert-manager/cert-manager/releases/tag/'"${CERT_MANAGER_RELEASE}"')|' \
    "${MANIFESTS_DIRECTORY}/README.md"
else
  sed -i '/| Cert Manager | common\/cert-manager |/s|\[.*\](https://github.com/cert-manager/cert-manager/releases/tag/v.*)|['"${CERT_MANAGER_RELEASE#v}"'](https://github.com/cert-manager/cert-manager/releases/tag/'"${CERT_MANAGER_RELEASE}"')|' \
    "${MANIFESTS_DIRECTORY}/README.md"
fi

commit_changes "$MANIFESTS_DIRECTORY" "Update common/cert-manager manifests to ${CERT_MANAGER_RELEASE}" \
  "$DESTINATION_DIRECTORY" \
  "README.md"

echo "Synchronization completed successfully."
