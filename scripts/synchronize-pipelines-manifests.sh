#!/usr/bin/env bash
# This script helps to create a PR to update the Kubeflow Pipelines manifests

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

setup_error_handling

COMPONENT_NAME="pipelines"
REPO_NAME="kubeflow/pipelines"
REPO_URL="https://github.com/kubeflow/pipelines.git"
COMMIT="2.4.1"
REPO_DIR="pipelines"
SRC_DIR=${SRC_DIR:=/tmp/kubeflow-${COMPONENT_NAME}}
BRANCH=${BRANCH:=synchronize-kubeflow-${COMPONENT_NAME}-manifests-${COMMIT?}}

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)
SRC_MANIFESTS_PATH="manifests/kustomize"
DST_MANIFESTS_PATH="apps/pipeline/upstream"

# README update patterns
SRC_TXT="\[.*\](https://github.com/${REPO_NAME}/tree/.*/manifests/kustomize)"
DST_TXT="\[${COMMIT}\](https://github.com/${REPO_NAME}/tree/${COMMIT}/manifests/kustomize)"

create_branch "$BRANCH"

clone_and_checkout "$SRC_DIR" "$REPO_URL" "$REPO_DIR" "$COMMIT"

echo "Copying ${COMPONENT_NAME} manifests..."
copy_manifests "${SRC_DIR}/${REPO_DIR}/${SRC_MANIFESTS_PATH}" "${MANIFESTS_DIR}/${DST_MANIFESTS_PATH}"

update_readme "$MANIFESTS_DIR" "$SRC_TXT" "$DST_TXT"

commit_changes "$MANIFESTS_DIR" "Update ${REPO_NAME} manifests from ${COMMIT}" \
  "apps" \
  "README.md"

echo "Synchronization completed successfully."
