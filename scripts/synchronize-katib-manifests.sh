#!/usr/bin/env bash
# This script helps to create a PR to update the Katib manifests

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

setup_error_handling

COMPONENT_NAME="katib"
REPO_NAME="kubeflow/katib"
REPO_URL="https://github.com/kubeflow/katib.git"
COMMIT="v0.18.0"
REPO_DIR="katib"
SRC_DIR=${SRC_DIR:=/tmp/kubeflow-${COMPONENT_NAME}}
BRANCH=${BRANCH:=synchronize-${COMPONENT_NAME}-manifests-${COMMIT?}}

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)
SRC_MANIFESTS_PATH="manifests/v1beta1"
DST_MANIFESTS_PATH="apps/${COMPONENT_NAME}/upstream"

# README update patterns
SRC_TXT="\[.*\](https://github.com/${REPO_NAME}/tree/.*/manifests/v1beta1)"
DST_TXT="\[${COMMIT}\](https://github.com/${REPO_NAME}/tree/${COMMIT}/manifests/v1beta1)"

create_branch "$BRANCH"

clone_and_checkout "$SRC_DIR" "$REPO_URL" "$REPO_DIR" "$COMMIT"

copy_manifests "${SRC_DIR}/${REPO_DIR}/${SRC_MANIFESTS_PATH}" "${MANIFESTS_DIR}/${DST_MANIFESTS_PATH}"

update_readme "$MANIFESTS_DIR" "$SRC_TXT" "$DST_TXT"

commit_changes "$MANIFESTS_DIR" "Update ${REPO_NAME} manifests from ${COMMIT}" \
  "apps" \
  "README.md"

echo "Synchronization completed successfully."
