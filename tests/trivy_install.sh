#!/bin/bash
set -euo pipefail

# Install trivy
echo "Installing Trivy..."
TRIVY_VERSION="v0.65.0"
TRIVY_VERSION_PLAIN="0.65.0"
TARGET_BINARY_DIRECTORY="/usr/local/bin"

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

TRIVY_ASSET="trivy_${TRIVY_VERSION_PLAIN}_Linux-64bit.tar.gz"
TRIVY_TEMPORARY_DIRECTORY=""
trap 'if [[ -n "${TRIVY_TEMPORARY_DIRECTORY:-}" ]]; then rm -rf "${TRIVY_TEMPORARY_DIRECTORY}"; fi' EXIT

TRIVY_TEMPORARY_DIRECTORY="$(mktemp -d "${TMPDIR:-/tmp}/trivy-install-XXXXXX")"

TRIVY_ARCHIVE_PATH="${TRIVY_TEMPORARY_DIRECTORY}/${TRIVY_ASSET}"
TRIVY_ALL_CHECKSUMS_FILE="${TRIVY_TEMPORARY_DIRECTORY}/checksums_all.txt"
TRIVY_CHECKSUMS_FILE="${TRIVY_TEMPORARY_DIRECTORY}/checksums.txt"

curl --fail --show-error --silent --location \
  "https://github.com/aquasecurity/trivy/releases/download/${TRIVY_VERSION}/${TRIVY_ASSET}" \
  --output "${TRIVY_ARCHIVE_PATH}"

curl --fail --show-error --silent --location \
  "https://github.com/aquasecurity/trivy/releases/download/${TRIVY_VERSION}/trivy_${TRIVY_VERSION_PLAIN}_checksums.txt" \
  --output "${TRIVY_ALL_CHECKSUMS_FILE}"

if ! awk -v asset="${TRIVY_ASSET}" \
  'BEGIN { found=0 } $2 == asset { print; found=1 } END { if (!found) exit 1 }' \
  "${TRIVY_ALL_CHECKSUMS_FILE}" > "${TRIVY_CHECKSUMS_FILE}"; then
    echo "Failed to extract checksum entry for ${TRIVY_ASSET} from ${TRIVY_ALL_CHECKSUMS_FILE}"
    exit 1
fi

if ! ( cd "${TRIVY_TEMPORARY_DIRECTORY}" && sha256sum --check "checksums.txt" ); then
    echo "Failed to verify Trivy checksums"
    exit 1
fi

tar -xzf "${TRIVY_ARCHIVE_PATH}" -C "${TRIVY_TEMPORARY_DIRECTORY}"

${INSTALL_COMMAND_PREFIX} mv "${TRIVY_TEMPORARY_DIRECTORY}/trivy" "${TARGET_BINARY_DIRECTORY}/trivy"
${INSTALL_COMMAND_PREFIX} chmod +x "${TARGET_BINARY_DIRECTORY}/trivy"

"${TARGET_BINARY_DIRECTORY}/trivy" version
