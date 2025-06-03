#!/usr/bin/env bash
# This script helps to create a PR to update the unified Istio manifests

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="istio"
COMMIT="1.27.0"  # Update this for new versions
CURRENT_VERSION="1-26" 
NEW_VERSION="1-27"  # Update this for new versions
SOURCE_DIRECTORY=${SOURCE_DIRECTORY:=/tmp/${COMPONENT_NAME}}
BRANCH_NAME=${BRANCH_NAME:=${COMPONENT_NAME}-${COMMIT?}}

# Path configurations
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
ISTIO_OLD=$MANIFESTS_DIRECTORY/common/${COMPONENT_NAME}-${CURRENT_VERSION}
ISTIO_NEW=$MANIFESTS_DIRECTORY/common/${COMPONENT_NAME}-${NEW_VERSION}

if [ ! -d "$ISTIO_NEW" ]; then
  cp -a $ISTIO_OLD $ISTIO_NEW
fi

create_branch "$BRANCH_NAME"

echo "Checking out in $SOURCE_DIRECTORY to $COMMIT..."
mkdir -p $SOURCE_DIRECTORY
cd $SOURCE_DIRECTORY
if [ ! -d "istio-${COMMIT}" ]; then
    wget "https://github.com/istio/istio/releases/download/${COMMIT}/istio-${COMMIT}-linux-amd64.tar.gz"
    tar xvfz istio-${COMMIT}-linux-amd64.tar.gz
fi

ISTIOCTL=$SOURCE_DIRECTORY/istio-${COMMIT}/bin/istioctl
cd $ISTIO_NEW

echo "Generating CNI manifests (default)..."
$ISTIOCTL manifest generate -f profile.yaml -f profile-overlay.yaml \
  --set components.cni.enabled=true \
  --set components.cni.namespace=kube-system > dump.yaml
./split-istio-packages -f dump.yaml
mv $ISTIO_NEW/crd.yaml $ISTIO_NEW/istio-crds/base/
mv $ISTIO_NEW/install.yaml $ISTIO_NEW/istio-install/base/
mv $ISTIO_NEW/cluster-local-gateway.yaml $ISTIO_NEW/cluster-local-gateway/base/
rm dump.yaml

echo "Generating non-CNI manifests (insecure overlay)..."
$ISTIOCTL manifest generate -f profile.yaml -f profile-overlay.yaml \
  --set components.cni.enabled=false > istio-install/overlays/insecure/install-insecure.yaml

check_uncommitted_changes

SOURCE_TEXT="\[.*\](https://github.com/istio/istio/releases/tag/.*)"
DESTINATION_TEXT="\[$COMMIT\](https://github.com/istio/istio/releases/tag/$COMMIT)"

update_readme "$MANIFESTS_DIRECTORY" "$SOURCE_TEXT" "$DESTINATION_TEXT"

echo "Synchronizing directory names..."
find "$MANIFESTS_DIRECTORY" -type f -not -path '*/.git/*' -exec sed -i "s/istio-${CURRENT_VERSION}/istio-${NEW_VERSION}/g" {} +

cd "$MANIFESTS_DIRECTORY"
if [ "$CURRENT_VERSION" != "$NEW_VERSION" ]; then
  rm -rf $ISTIO_OLD
fi
commit_changes "$MANIFESTS_DIRECTORY" "Upgrade istio to v.${COMMIT}" "."

echo "Synchronization completed successfully."