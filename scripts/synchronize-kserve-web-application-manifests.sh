#!/usr/bin/env bash
# This script helps to create a PR to update the KServe Models Web App manifests

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

setup_error_handling

COMPONENT_NAME="models-web-app"
REPO_NAME="kserve/models-web-app"
REPO_URL="https://github.com/kserve/models-web-app.git"
COMMIT="v0.14.0"
REPO_DIR="models-web-app"
SRC_DIR=${SRC_DIR:=/tmp/kserve-${COMPONENT_NAME}}
BRANCH=${BRANCH:=synchronize-kserve-${COMPONENT_NAME}-manifests-${COMMIT?}}

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)
SRC_MANIFESTS_PATH="config"
DST_MANIFESTS_PATH="apps/kserve/${COMPONENT_NAME}"

# README update patterns
SRC_TXT="\[.*\](https://github.com/${REPO_NAME}/tree/.*)"
DST_TXT="\[${COMMIT}\](https://github.com/${REPO_NAME}/tree/${COMMIT}/${SRC_MANIFESTS_PATH})"

create_branch "$BRANCH"

clone_and_checkout "$SRC_DIR" "$REPO_URL" "$REPO_DIR" "$COMMIT"

echo "Copying manifests"
copy_manifests "${SRC_DIR}/${REPO_DIR}/${SRC_MANIFESTS_PATH}" "${MANIFESTS_DIR}/${DST_MANIFESTS_PATH}"

update_readme "$MANIFESTS_DIR" "$SRC_TXT" "$DST_TXT"

commit_changes "$MANIFESTS_DIR" "Update kserve models web application manifests from ${COMMIT}" \
  "${DST_MANIFESTS_PATH}" \
  "README.md"

echo "Synchronization completed successfully."
