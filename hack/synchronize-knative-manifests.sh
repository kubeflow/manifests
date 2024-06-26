#!/usr/bin/env bash

# This script aims at helping create a PR to update the manifests of knative.
# This script:
# 1. Checks out a new branch
# 2. Download files into the correct places
# 3. Commits the changes
#
# Afterwards the developers can submit the PR to the kubeflow/manifests
# repository, based on that local branch
# It must be executed directly from its directory

# strict mode http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euxo pipefail
IFS=$'\n\t'

KN_SERVING_RELEASE="v1.12.4" # Must be a release
KN_EXTENSION_RELEASE="v1.12.3" # Must be a release
KN_EVENTING_RELEASE="v1.12.6" # Must be a release
BRANCH=${BRANCH:=synchronize-knative-manifests-${KN_SERVING_RELEASE?}}

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)

# replace source regex ($1) with target regex ($2)
# in file ($3)
replace_in_file() {
  SRC_TXT=$1
  DST_TXT=$2
  sed -i "s|$SRC_TXT|$DST_TXT|g" $3
}

echo "Creating branch: ${BRANCH}"

if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi
if [ `git branch --list $BRANCH` ]
then
   echo "WARNING: Branch $BRANCH already exists."
fi

# Create the branch in the manifests repository
if ! git show-ref --verify --quiet refs/heads/$BRANCH; then
    git checkout -b $BRANCH
else
    echo "Branch $BRANCH already exists."
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "WARNING: You have uncommitted changes"
fi

DST_DIR=$MANIFESTS_DIR/common/knative
if [ -d "$DST_DIR" ]; then
    # keep README and OWNERS file
    rm -r "$DST_DIR/knative-serving/base/upstream"
    rm "$DST_DIR/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml"
    rm -r "$DST_DIR/knative-eventing/base/upstream"
    rm "$DST_DIR/knative-eventing-post-install-jobs/base/eventing-post-install.yaml"
fi

mkdir -p "$DST_DIR/knative-serving/base/upstream"
mkdir -p "$DST_DIR/knative-serving-post-install-jobs/base"
mkdir -p "$DST_DIR/knative-eventing/base/upstream"
mkdir -p "$DST_DIR/knative-eventing-post-install-jobs/base"

echo "Downloading knative-serving manifests..."
# No need to install serving-crds.
# See: https://github.com/knative/serving/issues/9945
wget -O $DST_DIR/knative-serving/base/upstream/serving-core.yaml "https://github.com/knative/serving/releases/download/knative-$KN_SERVING_RELEASE/serving-core.yaml"
wget -O $DST_DIR/knative-serving/base/upstream/net-istio.yaml "https://github.com/knative-extensions/net-istio/releases/download/knative-$KN_EXTENSION_RELEASE/net-istio.yaml"
wget -O $DST_DIR/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml "https://github.com/knative/serving/releases/download/knative-$KN_SERVING_RELEASE/serving-post-install-jobs.yaml"

yq eval -i '... comments=""' $DST_DIR/knative-serving/base/upstream/serving-core.yaml
yq eval -i '... comments=""' $DST_DIR/knative-serving/base/upstream/net-istio.yaml
yq eval -i '... comments=""' $DST_DIR/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml

yq eval -i 'explode(.)' $DST_DIR/knative-serving/base/upstream/serving-core.yaml
yq eval -i 'explode(.)' $DST_DIR/knative-serving/base/upstream/net-istio.yaml
yq eval -i 'explode(.)' $DST_DIR/knative-serving-post-install-jobs/base/serving-post-install-jobs.yaml

# We are not using the '|=' operator because it generates an empty object
# ({}) which crashes kustomize.
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

# We are not using the '|=' operator because it generates an empty object
# ({}) which crashes kustomize.
yq eval -i 'select(.kind == "Job" and .metadata.generateName == "storage-version-migration-eventing-") | .metadata.name = "storage-version-migration-eventing"' $DST_DIR/knative-eventing-post-install-jobs/base/eventing-post-install.yaml

yq eval -i 'select((.kind == "ConfigMap" and .metadata.name == "config-observability") | not)' $DST_DIR/knative-eventing/base/upstream/in-memory-channel.yaml 
yq eval -i 'select((.kind == "ConfigMap" and .metadata.name == "config-tracing") | not)' $DST_DIR/knative-eventing/base/upstream/in-memory-channel.yaml 

echo "Successfully copied all manifests."

echo "Updating README..."

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

echo "Committing the changes..."
cd $MANIFESTS_DIR
git add $DST_DIR
git add README.md
git commit -s -m "Update common/knative manifests from ${KN_SERVING_RELEASE}/${KN_EVENTING_RELEASE}"
