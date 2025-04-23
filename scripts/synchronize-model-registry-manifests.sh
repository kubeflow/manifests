#!/usr/bin/env bash
# This script helps to create a PR to update the Model Registry manifests

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

setup_error_handling

COMPONENT_NAME="model-registry"
REPO_NAME="kubeflow/model-registry"
REPO_URL="https://github.com/kubeflow/model-registry.git"
COMMIT="v0.2.13"
REPO_DIR="model-registry"
DEV_MODE=${DEV_MODE:=false}
SRC_DIR=${SRC_DIR:=/tmp/kubeflow-${COMPONENT_NAME}}
BRANCH=${BRANCH:=synchronize-kubeflow-${COMPONENT_NAME}-manifests-${COMMIT?}}

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)
SRC_MANIFESTS_PATH="manifests/kustomize"
DST_MANIFESTS_PATH="apps/${COMPONENT_NAME}/upstream"

# README update patterns
SRC_TXT="\[.*\](https://github.com/${REPO_NAME}/tree/.*/manifests/kustomize)"
DST_TXT="\[${COMMIT}\](https://github.com/${REPO_NAME}/tree/${COMMIT}/manifests/kustomize)"

create_branch "$BRANCH"

clone_and_checkout "$SRC_DIR" "$REPO_URL" "$REPO_DIR" "$COMMIT"

echo "Copying ${COMPONENT_NAME} manifests..."
DST_DIR=$MANIFESTS_DIR/$DST_MANIFESTS_PATH
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR"
fi
mkdir -p $DST_DIR
cp -r "$SRC_DIR/$REPO_DIR/$SRC_MANIFESTS_PATH/"* "$DST_DIR"

echo "Successfully copied all manifests."

update_readme "$MANIFESTS_DIR" "$SRC_TXT" "$DST_TXT"

commit_changes "$MANIFESTS_DIR" "Update ${REPO_NAME} manifests from ${COMMIT}" \
  "apps" \
  "README.md"

echo "Synchronization completed successfully."
