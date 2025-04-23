#!/usr/bin/env bash
# This script helps to create a PR to update the KServe manifests

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

setup_error_handling

COMPONENT_NAME="kserve"
REPO_NAME="kserve/kserve"
REPO_URL="https://github.com/kserve/kserve.git"
KSERVE_VERSION="v0.15.0"
COMMIT="v0.15.0"
REPO_DIR="kserve"
SRC_DIR=${SRC_DIR:=/tmp/${COMPONENT_NAME}}
BRANCH=${BRANCH:=synchronize-${COMPONENT_NAME}-manifests-${COMMIT?}}

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)
SRC_MANIFESTS_PATH="install/${KSERVE_VERSION}"
DST_MANIFESTS_PATH="apps/${COMPONENT_NAME}/${COMPONENT_NAME}"

# README update patterns
SRC_TXT="\[.*\](https://github.com/${REPO_NAME}/releases/tag/.*)"
DST_TXT="\[${COMMIT}\](https://github.com/${REPO_NAME}/releases/tag/${COMMIT}/install/${KSERVE_VERSION})"

create_branch "$BRANCH"

clone_and_checkout "$SRC_DIR" "$REPO_URL" "$REPO_DIR" "$COMMIT"

echo "Copying kserve manifests..."
DST_DIR=$MANIFESTS_DIR/$DST_MANIFESTS_PATH
if [ -d "$DST_DIR" ]; then
    rm -rf "$DST_DIR"/kserve*
fi
cp $SRC_DIR/$REPO_DIR/$SRC_MANIFESTS_PATH/* $DST_DIR -r

echo "Successfully copied all manifests."

update_readme "$MANIFESTS_DIR" "$SRC_TXT" "$DST_TXT"

commit_changes "$MANIFESTS_DIR" "Update ${REPO_NAME} manifests from ${KSERVE_VERSION}" \
  "apps/${COMPONENT_NAME}" \
  "README.md" \
  "scripts"

echo "Synchronization completed successfully."
