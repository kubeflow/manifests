#!/usr/bin/env bash
# This script helps to create a PR to update the Kubeflow manifests
SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"
setup_error_handling
COMPONENT_NAME="kubeflow"
REPOSITORY_NAME="kubeflow/kubeflow"
REPOSITORY_URL="https://github.com/kubeflow/kubeflow.git"
COMMIT="v1.10.0"
REPOSITORY_DIRECTORY="kubeflow"
SOURCE_DIRECTORY=${SOURCE_DIRECTORY:=/tmp/${COMPONENT_NAME}-${COMPONENT_NAME}}
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-${COMPONENT_NAME}-manifests-${COMMIT?}}
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
create_branch "$BRANCH_NAME"
check_uncommitted_changes
clone_and_checkout "$SOURCE_DIRECTORY" "$REPOSITORY_URL" "$REPOSITORY_DIRECTORY" "$COMMIT"
copy_component_manifests() {
    local source_path=$1
    local destination_path=$2
    local readme_path_pattern=$3
    local destination_directory="${MANIFESTS_DIRECTORY}/${destination_path}"
    if [ -d "$destination_directory" ]; then
        rm -r "$destination_directory"
    fi
    mkdir -p "$destination_directory"
    cp "${SOURCE_DIRECTORY}/${REPOSITORY_DIRECTORY}/${source_path}/"* "$destination_directory" -r
    local source_text="\[.*\](https://github.com/${REPOSITORY_NAME}/tree/.*/components/${readme_path_pattern})"
    local destination_text="\[${COMMIT}\](https://github.com/${REPOSITORY_NAME}/tree/${COMMIT}/components/${readme_path_pattern})"
    update_readme "$MANIFESTS_DIRECTORY" "$source_text" "$destination_text"
}
copy_component_manifests "components/admission-webhook/manifests" \
    "applications/admission-webhook/upstream" \
    "admission-webhook/manifests"
copy_component_manifests "components/centraldashboard/manifests" \
    "applications/centraldashboard/upstream" \
    "centraldashboard/manifests"
copy_component_manifests "components/crud-web-applications/jupyter/manifests" \
    "applications/jupyter/jupyter-web-app/upstream" \
    "crud-web-applications/jupyter/manifests"
copy_component_manifests "components/crud-web-applications/volumes/manifests" \
    "applications/volumes-web-app/upstream" \
    "crud-web-applications/volumes/manifests"
copy_component_manifests "components/crud-web-applications/tensorboards/manifests" \
    "applications/tensorboard/tensorboards-web-app/upstream" \
    "crud-web-applications/tensorboards/manifests"
copy_component_manifests "components/profile-controller/config" \
    "applications/profiles/upstream" \
    "profile-controller/config"
copy_component_manifests "components/notebook-controller/config" \
    "applications/jupyter/notebook-controller/upstream" \
    "notebook-controller/config"
copy_component_manifests "components/tensorboard-controller/config" \
    "applications/tensorboard/tensorboard-controller/upstream" \
    "tensorboard-controller/config"
copy_component_manifests "components/pvcviewer-controller/config" \
    "applications/pvcviewer-controller/upstream" \
    "pvcviewer-controller/config"
commit_changes "$MANIFESTS_DIRECTORY" "Update ${REPOSITORY_NAME} manifests from ${COMMIT}" \
  "applications" \
  "README.md"
echo "Synchronization completed successfully."
