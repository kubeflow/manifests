#!/usr/bin/env bash
# Template for Kubeflow manifests synchronization scripts
# Usage: Copy this file and adjust the variables for your specific component

# Source the common library functions
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

# Initialize error handling
setup_error_handling

# Configuration variables (adjust these for your specific component)
COMPONENT_NAME="component-name"  # Name of the component (e.g., katib, training-operator)
REPO_NAME="repo-name"            # Repository name (e.g., kubeflow/katib)
REPO_URL="https://github.com/org/repo.git"  # Repository URL to clone from
COMMIT="v0.0.0"                  # Version/commit to synchronize
REPO_DIR="${COMPONENT_NAME}"     # Directory name within SRC_DIR
SRC_DIR=${SRC_DIR:=/tmp/kubeflow-${COMPONENT_NAME}}  # Where to clone the source repo
BRANCH=${BRANCH:=synchronize-${COMPONENT_NAME}-manifests-${COMMIT?}}  # Branch name for the PR

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)
SRC_MANIFESTS_PATH="manifests"   # Path within source repo where manifests are
DST_MANIFESTS_PATH="apps/${COMPONENT_NAME}/upstream"  # Destination path within manifests repo

# README update patterns
SRC_TXT="\[.*\](https://github.com/${REPO_NAME}/tree/.*/manifests)"
DST_TXT="\[${COMMIT}\](https://github.com/${REPO_NAME}/tree/${COMMIT}/manifests)"

# Create branch
create_branch "$BRANCH"

# Clone repository and checkout to specific commit
clone_and_checkout "$SRC_DIR" "$REPO_URL" "$REPO_DIR" "$COMMIT"

# Copy manifests
copy_manifests "${SRC_DIR}/${REPO_DIR}/${SRC_MANIFESTS_PATH}" "${MANIFESTS_DIR}/${DST_MANIFESTS_PATH}"

# Update README
update_readme "$MANIFESTS_DIR" "$SRC_TXT" "$DST_TXT"

# Commit changes
commit_changes "$MANIFESTS_DIR" "Update ${REPO_NAME} manifests from ${COMMIT}" \
  "${DST_MANIFESTS_PATH}" \
  "README.md"

echo "Synchronization completed successfully." 