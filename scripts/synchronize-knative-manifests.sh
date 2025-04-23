#!/usr/bin/env bash
# This script helps to create a PR to update the Knative manifests

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

setup_error_handling

COMPONENT_NAME="knative"
KN_SERVING_RELEASE="v1.16.2" # Must be a release
KN_EXTENSION_RELEASE="v1.16.0" # Must be a release
KN_EVENTING_RELEASE="v1.16.4" # Must be a release
BRANCH=${BRANCH:=synchronize-${COMPONENT_NAME}-manifests-${KN_SERVING_RELEASE?}}

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)
DST_DIR=$MANIFESTS_DIR/common/${COMPONENT_NAME}

create_branch "$BRANCH"

check_uncommitted_changes

# Clean up existing files (keep README and OWNERS)
if [ -d "$DST_DIR" ]; then
    rm -r "$DST_DIR/knative-serving/base/upstream" 2>/dev/null || true
    rm "$DST_DIR/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml" 2>/dev/null || true
    rm -r "$DST_DIR/knative-eventing/base/upstream" 2>/dev/null || true
    rm "$DST_DIR/knative-eventing-post-install-jobs/base/eventing-post-install.yaml" 2>/dev/null || true
fi

# Create required directories
mkdir -p "$DST_DIR/knative-serving/base/upstream"
mkdir -p "$DST_DIR/knative-serving-post-install-jobs/base"
mkdir -p "$DST_DIR/knative-eventing/base/upstream"
mkdir -p "$DST_DIR/knative-eventing-post-install-jobs/base"

echo "Downloading knative-serving manifests..."
wget -O $DST_DIR/knative-serving/base/upstream/serving-core.yaml "https://github.com/knative/serving/releases/download/knative-$KN_SERVING_RELEASE/serving-core.yaml"
wget -O $DST_DIR/knative-serving/base/upstream/net-istio.yaml "https://github.com/knative-extensions/net-istio/releases/download/knative-$KN_EXTENSION_RELEASE/net-istio.yaml"
wget -O $DST_DIR/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml "https://github.com/knative/serving/releases/download/knative-$KN_SERVING_RELEASE/serving-post-install-jobs.yaml"

yq eval -i '... comments=""' $DST_DIR/knative-serving/base/upstream/serving-core.yaml
yq eval -i '... comments=""' $DST_DIR/knative-serving/base/upstream/net-istio.yaml
yq eval -i '... comments=""' $DST_DIR/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml

yq eval -i 'explode(.)' $DST_DIR/knative-serving/base/upstream/serving-core.yaml
yq eval -i 'explode(.)' $DST_DIR/knative-serving/base/upstream/net-istio.yaml
yq eval -i 'explode(.)' $DST_DIR/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml

yq eval -i 'select(.kind == "Job" and .metadata.generateName == "storage-version-migration-serving-") | .metadata.name = "storage-version-migration-serving"' $DST_DIR/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml

echo "Downloading knative-eventing manifests..."
wget -O $DST_DIR/knative-eventing/base/upstream/eventing-core.yaml "https://github.com/knative/eventing/releases/download/knative-$KN_EVENTING_RELEASE/eventing-core.yaml"
wget -O $DST_DIR/knative-eventing/base/upstream/in-memory-channel.yaml "https://github.com/knative/eventing/releases/download/knative-$KN_EVENTING_RELEASE/in-memory-channel.yaml"
wget -O $DST_DIR/knative-eventing/base/upstream/mt-channel-broker.yaml "https://github.com/knative/eventing/releases/download/knative-$KN_EVENTING_RELEASE/mt-channel-broker.yaml"
wget -O $DST_DIR/knative-eventing-post-install-jobs/base/eventing-post-install.yaml "https://github.com/knative/eventing/releases/download/knative-$KN_EVENTING_RELEASE/eventing-post-install.yaml"

yq eval -i '... comments=""' $DST_DIR/knative-eventing/base/upstream/eventing-core.yaml
yq eval -i '... comments=""' $DST_DIR/knative-eventing/base/upstream/in-memory-channel.yaml
yq eval -i '... comments=""' $DST_DIR/knative-eventing/base/upstream/mt-channel-broker.yaml
yq eval -i '... comments=""' $DST_DIR/knative-eventing-post-install-jobs/base/eventing-post-install.yaml

yq eval -i 'explode(.)' $DST_DIR/knative-eventing/base/upstream/eventing-core.yaml
yq eval -i 'explode(.)' $DST_DIR/knative-eventing/base/upstream/in-memory-channel.yaml
yq eval -i 'explode(.)' $DST_DIR/knative-eventing/base/upstream/mt-channel-broker.yaml
yq eval -i 'explode(.)' $DST_DIR/knative-eventing-post-install-jobs/base/eventing-post-install.yaml

yq eval -i 'select(.kind == "Job" and .metadata.generateName == "storage-version-migration-eventing-") | .metadata.name = "storage-version-migration-eventing"' $DST_DIR/knative-eventing-post-install-jobs/base/eventing-post-install.yaml

yq eval -i 'select((.kind == "ConfigMap" and .metadata.name == "config-observability") | not)' $DST_DIR/knative-eventing/base/upstream/in-memory-channel.yaml 
yq eval -i 'select((.kind == "ConfigMap" and .metadata.name == "config-tracing") | not)' $DST_DIR/knative-eventing/base/upstream/in-memory-channel.yaml 

echo "Successfully copied all manifests."

echo "Updating READMEs..."
# Helper function to replace text in files
replace_in_file() {
  local SRC_TXT=$1
  local DST_TXT=$2
  local FILE=$3
  sed -i "s|$SRC_TXT|$DST_TXT|g" $FILE
}

replace_in_file \
  "\[.*\](https://github.com/knative/serving/releases/tag/knative-.*) <" \
  "\[$KN_SERVING_RELEASE\](https://github.com/knative/serving/releases/tag/knative-$KN_SERVING_RELEASE) <" \
  ${MANIFESTS_DIR}/README.md

replace_in_file \
  "> \[.*\](https://github.com/knative/eventing/releases/tag/knative-.*)" \
  "> \[$KN_EVENTING_RELEASE\](https://github.com/knative/eventing/releases/tag/knative-$KN_EVENTING_RELEASE)" \
  ${MANIFESTS_DIR}/README.md

replace_in_file \
  "\[Knative serving (v.*)\](https://github.com/knative/serving/releases/tag/knative-v.*)" \
  "\[Knative serving ($KN_SERVING_RELEASE)\](https://github.com/knative/serving/releases/tag/knative-$KN_SERVING_RELEASE)" \
  $DST_DIR/README.md

replace_in_file \
  "\[Knative ingress controller for Istio (v.*)\](https://github.com/knative-extensions/net-istio/releases/tag/knative-v.*)" \
  "\[Knative ingress controller for Istio ($KN_EXTENSION_RELEASE)\](https://github.com/knative-extensions/net-istio/releases/tag/knative-$KN_EXTENSION_RELEASE)" \
  $DST_DIR/README.md

replace_in_file \
  "The manifests for Knative Eventing are based off the \[v.* release\](https://github.com/knative/eventing/releases/tag/knative-v.*)" \
  "The manifests for Knative Eventing are based off the \[$KN_EVENTING_RELEASE release\](https://github.com/knative/eventing/releases/tag/knative-$KN_EVENTING_RELEASE)" \
  $DST_DIR/README.md

commit_changes "$MANIFESTS_DIR" "Update common/knative manifests from ${KN_SERVING_RELEASE}/${KN_EVENTING_RELEASE}" \
  "$DST_DIR" \
  "README.md"

echo "Synchronization completed successfully."
