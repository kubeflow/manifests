#!/usr/bin/env bash
# This script helps to create a PR to update the Knative manifests

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="knative"
KN_SERVING_RELEASE="v1.20.0" # Must be a release
KN_EXTENSION_RELEASE="v1.20.1" # Must be a release
KN_EVENTING_RELEASE="v1.20.0" # Must be a release
BRANCH_NAME=${BRANCH_NAME:=synchronize-${COMPONENT_NAME}-manifests-${KN_SERVING_RELEASE?}}

# Path configurations
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
DESTINATION_DIRECTORY=$MANIFESTS_DIRECTORY/common/${COMPONENT_NAME}

create_branch "$BRANCH_NAME"

check_uncommitted_changes

# Clean up existing files (keep README and OWNERS)
if [ -d "$DESTINATION_DIRECTORY" ]; then
    rm -r "$DESTINATION_DIRECTORY/knative-serving/base/upstream"
    rm "$DESTINATION_DIRECTORY/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml"
    rm -r "$DESTINATION_DIRECTORY/knative-eventing/base/upstream"
    rm "$DESTINATION_DIRECTORY/knative-eventing-post-install-jobs/base/eventing-post-install.yaml"
fi

# Create required directories
mkdir -p "$DESTINATION_DIRECTORY/knative-serving/base/upstream"
mkdir -p "$DESTINATION_DIRECTORY/knative-serving-post-install-jobs/base"
mkdir -p "$DESTINATION_DIRECTORY/knative-eventing/base/upstream"
mkdir -p "$DESTINATION_DIRECTORY/knative-eventing-post-install-jobs/base"

echo "Downloading knative-serving manifests..."
wget -O $DESTINATION_DIRECTORY/knative-serving/base/upstream/serving-core.yaml "https://github.com/knative/serving/releases/download/knative-$KN_SERVING_RELEASE/serving-core.yaml"
wget -O $DESTINATION_DIRECTORY/knative-serving/base/upstream/net-istio.yaml "https://github.com/knative-extensions/net-istio/releases/download/knative-$KN_EXTENSION_RELEASE/net-istio.yaml"
wget -O $DESTINATION_DIRECTORY/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml "https://github.com/knative/serving/releases/download/knative-$KN_SERVING_RELEASE/serving-post-install-jobs.yaml"

yq eval -i '... comments=""' $DESTINATION_DIRECTORY/knative-serving/base/upstream/serving-core.yaml
yq eval -i '... comments=""' $DESTINATION_DIRECTORY/knative-serving/base/upstream/net-istio.yaml
yq eval -i '... comments=""' $DESTINATION_DIRECTORY/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml

yq eval -i 'explode(.)' $DESTINATION_DIRECTORY/knative-serving/base/upstream/serving-core.yaml
yq eval -i 'explode(.)' $DESTINATION_DIRECTORY/knative-serving/base/upstream/net-istio.yaml
yq eval -i 'explode(.)' $DESTINATION_DIRECTORY/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml

yq eval -i 'select(.kind == "Job" and .metadata.generateName == "storage-version-migration-serving-") | .metadata.name = "storage-version-migration-serving"' $DESTINATION_DIRECTORY/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml

echo "Downloading knative-eventing manifests..."
wget -O $DESTINATION_DIRECTORY/knative-eventing/base/upstream/eventing-core.yaml "https://github.com/knative/eventing/releases/download/knative-$KN_EVENTING_RELEASE/eventing-core.yaml"
wget -O $DESTINATION_DIRECTORY/knative-eventing/base/upstream/in-memory-channel.yaml "https://github.com/knative/eventing/releases/download/knative-$KN_EVENTING_RELEASE/in-memory-channel.yaml"
wget -O $DESTINATION_DIRECTORY/knative-eventing/base/upstream/mt-channel-broker.yaml "https://github.com/knative/eventing/releases/download/knative-$KN_EVENTING_RELEASE/mt-channel-broker.yaml"
wget -O $DESTINATION_DIRECTORY/knative-eventing-post-install-jobs/base/eventing-post-install.yaml "https://github.com/knative/eventing/releases/download/knative-$KN_EVENTING_RELEASE/eventing-post-install.yaml"

yq eval -i '... comments=""' $DESTINATION_DIRECTORY/knative-eventing/base/upstream/eventing-core.yaml
yq eval -i '... comments=""' $DESTINATION_DIRECTORY/knative-eventing/base/upstream/in-memory-channel.yaml
yq eval -i '... comments=""' $DESTINATION_DIRECTORY/knative-eventing/base/upstream/mt-channel-broker.yaml
yq eval -i '... comments=""' $DESTINATION_DIRECTORY/knative-eventing-post-install-jobs/base/eventing-post-install.yaml

yq eval -i 'explode(.)' $DESTINATION_DIRECTORY/knative-eventing/base/upstream/eventing-core.yaml
yq eval -i 'explode(.)' $DESTINATION_DIRECTORY/knative-eventing/base/upstream/in-memory-channel.yaml
yq eval -i 'explode(.)' $DESTINATION_DIRECTORY/knative-eventing/base/upstream/mt-channel-broker.yaml
yq eval -i 'explode(.)' $DESTINATION_DIRECTORY/knative-eventing-post-install-jobs/base/eventing-post-install.yaml

yq eval -i 'select(.kind == "Job" and .metadata.generateName == "storage-version-migration-eventing-") | .metadata.name = "storage-version-migration-eventing"' $DESTINATION_DIRECTORY/knative-eventing-post-install-jobs/base/eventing-post-install.yaml

yq eval -i 'select((.kind == "ConfigMap" and .metadata.name == "config-observability") | not)' $DESTINATION_DIRECTORY/knative-eventing/base/upstream/in-memory-channel.yaml 
yq eval -i 'select((.kind == "ConfigMap" and .metadata.name == "config-tracing") | not)' $DESTINATION_DIRECTORY/knative-eventing/base/upstream/in-memory-channel.yaml 

# Helper function to replace text in files
replace_in_file() {
  local SOURCE_TEXT=$1
  local DESTINATION_TEXT=$2
  local FILE=$3
  sed -i "s|$SOURCE_TEXT|$DESTINATION_TEXT|g" $FILE
}

replace_in_file \
  "\[.*\](https://github.com/knative/serving/releases/tag/knative-.*) <" \
  "\[$KN_SERVING_RELEASE\](https://github.com/knative/serving/releases/tag/knative-$KN_SERVING_RELEASE) <" \
  ${MANIFESTS_DIRECTORY}/README.md

replace_in_file \
  "> \[.*\](https://github.com/knative/eventing/releases/tag/knative-.*)" \
  "> \[$KN_EVENTING_RELEASE\](https://github.com/knative/eventing/releases/tag/knative-$KN_EVENTING_RELEASE)" \
  ${MANIFESTS_DIRECTORY}/README.md

replace_in_file \
  "\[Knative serving (v.*)\](https://github.com/knative/serving/releases/tag/knative-v.*)" \
  "\[Knative serving ($KN_SERVING_RELEASE)\](https://github.com/knative/serving/releases/tag/knative-$KN_SERVING_RELEASE)" \
  $DESTINATION_DIRECTORY/README.md

replace_in_file \
  "\[Knative ingress controller for Istio (v.*)\](https://github.com/knative-extensions/net-istio/releases/tag/knative-v.*)" \
  "\[Knative ingress controller for Istio ($KN_EXTENSION_RELEASE)\](https://github.com/knative-extensions/net-istio/releases/tag/knative-$KN_EXTENSION_RELEASE)" \
  $DESTINATION_DIRECTORY/README.md

replace_in_file \
  "The manifests for Knative Eventing are based off the \[v.* release\](https://github.com/knative/eventing/releases/tag/knative-v.*)" \
  "The manifests for Knative Eventing are based off the \[$KN_EVENTING_RELEASE release\](https://github.com/knative/eventing/releases/tag/knative-$KN_EVENTING_RELEASE)" \
  $DESTINATION_DIRECTORY/README.md

commit_changes "$MANIFESTS_DIRECTORY" "Update common/knative manifests from ${KN_SERVING_RELEASE}/${KN_EVENTING_RELEASE}" \
  "$DESTINATION_DIRECTORY" \
  "README.md"

echo "Synchronization completed successfully."
