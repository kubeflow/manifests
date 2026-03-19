#!/bin/bash
# NOTE: This script is invoked from CI workflows.
set -euo pipefail

KUSTOMIZE_VERSION="${1:-v5.8.1}"
TARGET_BINARY_DIRECTORY="${2:-/usr/local/bin}"

# In GitHub Actions, we always want to target /usr/local/bin
if [[ "${GITHUB_ACTIONS:-false}" == "true" ]]; then
    TARGET_BINARY_DIRECTORY="/usr/local/bin"
fi

INSTALL_COMMAND_PREFIX=""
if [[ -d "${TARGET_BINARY_DIRECTORY}" ]]; then
    if [[ ! -w "${TARGET_BINARY_DIRECTORY}" ]]; then
        if command -v sudo >/dev/null 2>&1; then
            INSTALL_COMMAND_PREFIX="sudo"
        else
            echo "Target directory ${TARGET_BINARY_DIRECTORY} is not writable and sudo is not available." >&2
            exit 1
        fi
    fi
else
    parent_directory_path="$(dirname "${TARGET_BINARY_DIRECTORY}")"
    if [[ ! -w "${parent_directory_path}" ]]; then
        if command -v sudo >/dev/null 2>&1; then
            INSTALL_COMMAND_PREFIX="sudo"
        else
            echo "Parent directory ${parent_directory_path} of target directory ${TARGET_BINARY_DIRECTORY} is not writable and sudo is not available." >&2
            exit 1
        fi
    fi
fi

${INSTALL_COMMAND_PREFIX} mkdir -p "${TARGET_BINARY_DIRECTORY}"

echo "Installing Kustomize ${KUSTOMIZE_VERSION} to ${TARGET_BINARY_DIRECTORY}..."

KUSTOMIZE_ASSET="kustomize_${KUSTOMIZE_VERSION}_linux_amd64.tar.gz"
KUSTOMIZE_TEMPORARY_DIRECTORY=""
trap 'if [[ -n "${KUSTOMIZE_TEMPORARY_DIRECTORY:-}" ]]; then rm -rf "${KUSTOMIZE_TEMPORARY_DIRECTORY}"; fi' EXIT

KUSTOMIZE_TEMPORARY_DIRECTORY="$(mktemp -d "${TMPDIR:-/tmp}/kustomize-install-XXXXXX")"

KUSTOMIZE_ARCHIVE_PATH="${KUSTOMIZE_TEMPORARY_DIRECTORY}/${KUSTOMIZE_ASSET}"
KUSTOMIZE_ALL_CHECKSUMS_FILE="${KUSTOMIZE_TEMPORARY_DIRECTORY}/checksums_all.txt"
KUSTOMIZE_CHECKSUMS_FILE="${KUSTOMIZE_TEMPORARY_DIRECTORY}/checksums.txt"

curl --fail --show-error --silent --location \
  "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/${KUSTOMIZE_ASSET}" \
  --output "${KUSTOMIZE_ARCHIVE_PATH}"

curl --fail --show-error --silent --location \
  "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/checksums.txt" \
  --output "${KUSTOMIZE_ALL_CHECKSUMS_FILE}"

if ! awk -v asset="${KUSTOMIZE_ASSET}" \
  'BEGIN { found=0 } $2 == asset { print; found=1 } END { if (!found) exit 1 }' \
  "${KUSTOMIZE_ALL_CHECKSUMS_FILE}" > "${KUSTOMIZE_CHECKSUMS_FILE}"; then
    echo "Failed to extract checksum entry for ${KUSTOMIZE_ASSET} from ${KUSTOMIZE_ALL_CHECKSUMS_FILE}"
    exit 1
fi

if ! ( cd "${KUSTOMIZE_TEMPORARY_DIRECTORY}" && sha256sum --check "checksums.txt" ); then
    echo "Failed to verify Kustomize checksums"
    exit 1
fi

tar -xzf "${KUSTOMIZE_ARCHIVE_PATH}" -C "${KUSTOMIZE_TEMPORARY_DIRECTORY}"


${INSTALL_COMMAND_PREFIX} mv "${KUSTOMIZE_TEMPORARY_DIRECTORY}/kustomize" "${TARGET_BINARY_DIRECTORY}/kustomize"
${INSTALL_COMMAND_PREFIX} chmod +x "${TARGET_BINARY_DIRECTORY}/kustomize"

"${TARGET_BINARY_DIRECTORY}/kustomize" version
