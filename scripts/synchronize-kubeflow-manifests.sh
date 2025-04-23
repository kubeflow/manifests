#!/usr/bin/env bash
# This script helps to create a PR to update the Kubeflow manifests

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

setup_error_handling

COMPONENT_NAME="kubeflow"
REPO_NAME="kubeflow/kubeflow"
REPO_URL="https://github.com/kubeflow/kubeflow.git"
COMMIT="v1.10.0"
REPO_DIR="kubeflow"
SRC_DIR=${SRC_DIR:=/tmp/${COMPONENT_NAME}-${COMPONENT_NAME}}
BRANCH=${BRANCH:=synchronize-${COMPONENT_NAME}-${COMPONENT_NAME}-manifests-${COMMIT?}}

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)

create_branch "$BRANCH"

clone_and_checkout "$SRC_DIR" "$REPO_URL" "$REPO_DIR" "$COMMIT"

# Function to copy manifests for a specific component
copy_component_manifests() {
    local component_name=$1
    local src_path=$2
    local dst_path=$3
    local readme_path_pattern=$4
    
    echo "Copying ${component_name} manifests..."
    
    local dst_dir="${MANIFESTS_DIR}/${dst_path}"
    if [ -d "$dst_dir" ]; then
        rm -r "$dst_dir"
    fi
    mkdir -p "$dst_dir"
    
    cp "${SRC_DIR}/${REPO_DIR}/${src_path}/"* "$dst_dir" -r
    
    echo "Updating README for ${component_name}..."
    local src_txt="\[.*\](https://github.com/${REPO_NAME}/tree/.*/components/${readme_path_pattern})"
    local dst_txt="\[${COMMIT}\](https://github.com/${REPO_NAME}/tree/${COMMIT}/components/${readme_path_pattern})"
    
    update_readme "$MANIFESTS_DIR" "$src_txt" "$dst_txt"
}

copy_component_manifests "admission-webhook" \
    "components/admission-webhook/manifests" \
    "apps/admission-webhook/upstream" \
    "admission-webhook/manifests"

copy_component_manifests "centraldashboard" \
    "components/centraldashboard/manifests" \
    "apps/centraldashboard/upstream" \
    "centraldashboard/manifests"

copy_component_manifests "jupyter-web-app" \
    "components/crud-web-apps/jupyter/manifests" \
    "apps/jupyter/jupyter-web-app/upstream" \
    "crud-web-apps/jupyter/manifests"

copy_component_manifests "volumes-web-app" \
    "components/crud-web-apps/volumes/manifests" \
    "apps/volumes-web-app/upstream" \
    "crud-web-apps/volumes/manifests"

copy_component_manifests "tensorboards-web-app" \
    "components/crud-web-apps/tensorboards/manifests" \
    "apps/tensorboard/tensorboards-web-app/upstream" \
    "crud-web-apps/tensorboards/manifests"

copy_component_manifests "profile-controller" \
    "components/profile-controller/config" \
    "apps/profiles/upstream" \
    "profile-controller/config"

copy_component_manifests "notebook-controller" \
    "components/notebook-controller/config" \
    "apps/jupyter/notebook-controller/upstream" \
    "notebook-controller/config"

copy_component_manifests "tensorboard-controller" \
    "components/tensorboard-controller/config" \
    "apps/tensorboard/tensorboard-controller/upstream" \
    "tensorboard-controller/config"

copy_component_manifests "pvcviewer-controller" \
    "components/pvcviewer-controller/config" \
    "apps/pvcviewer-controller/upstream" \
    "pvcviewer-controller/config"

echo "Successfully copied all manifests."

commit_changes "$MANIFESTS_DIR" "Update ${REPO_NAME} manifests from ${COMMIT}" \
  "apps" \
  "README.md"

echo "Synchronization completed successfully."
