#!/bin/bash
set -euxo pipefail

TRIVY_VERSION="v0.69.3"
USER_BINARY_DIRECTORY="$HOME/.local/bin"

if [[ "${GITHUB_ACTIONS:-false}" == "true" ]]; then
    USER_BINARY_DIRECTORY="/tmp/usr/local/bin"
fi

mkdir -p "${USER_BINARY_DIRECTORY}"
export PATH="${USER_BINARY_DIRECTORY}:${PATH}"

echo "Install Trivy ..."
{
    curl --fail --show-error --silent --location \
      https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | \
      sh -s -- -b "${USER_BINARY_DIRECTORY}" "${TRIVY_VERSION}"
} || { echo "Failed to install Trivy"; exit 1; }

