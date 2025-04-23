#!/usr/bin/env bash
# This script helps to create a PR to update the Istio CNI manifests

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIR}/lib.sh"

setup_error_handling

COMPONENT_NAME="istio-cni"
COMMIT="1.24.3"
CURRENT_VERSION="1-24" 
NEW_VERSION="1-24" # Must be a release
SRC_DIR=${SRC_DIR:=/tmp/${COMPONENT_NAME}}
BRANCH=${BRANCH:=${COMPONENT_NAME}-${COMMIT?}}

# Path configurations
MANIFESTS_DIR=$(dirname $SCRIPT_DIR)
ISTIO_OLD=$MANIFESTS_DIR/common/${COMPONENT_NAME}-${CURRENT_VERSION}
ISTIO_NEW=$MANIFESTS_DIR/common/${COMPONENT_NAME}-${NEW_VERSION}

if [ ! -d "$ISTIO_NEW" ]; then
  cp -a $ISTIO_OLD $ISTIO_NEW
fi

create_branch "$BRANCH"

echo "Checking out in $SRC_DIR to $COMMIT..."
mkdir -p $SRC_DIR
cd $SRC_DIR
if [ ! -d "istio-${COMMIT}" ]; then
    wget "https://github.com/istio/istio/releases/download/${COMMIT}/istio-${COMMIT}-linux-amd64.tar.gz"
    tar xvfz istio-${COMMIT}-linux-amd64.tar.gz
fi

ISTIOCTL=$SRC_DIR/istio-${COMMIT}/bin/istioctl
cd $ISTIO_NEW

$ISTIOCTL manifest generate -f profile.yaml -f profile-overlay.yaml --set components.cni.enabled=true --set components.cni.namespace=kube-system > dump.yaml
./split-istio-packages -f dump.yaml
mv $ISTIO_NEW/crd.yaml $ISTIO_NEW/istio-crds/base
mv $ISTIO_NEW/install.yaml $ISTIO_NEW/istio-install/base
mv $ISTIO_NEW/cluster-local-gateway.yaml $ISTIO_NEW/cluster-local-gateway/base
rm dump.yaml

check_uncommitted_changes

echo "Updating README..."
SRC_TXT="\[.*\](https://github.com/istio/istio/releases/tag/.*)"
DST_TXT="\[$COMMIT\](https://github.com/istio/istio/releases/tag/$COMMIT)"

update_readme "$MANIFESTS_DIR" "$SRC_TXT" "$DST_TXT"

echo "Synchronizing directory names..."
find "$MANIFESTS_DIR" -type f -not -path '*/.git/*' -exec sed -i "s/istio-cni-${CURRENT_VERSION}/istio-cni-${NEW_VERSION}/g" {} +

cd "$MANIFESTS_DIR"
if [ "$CURRENT_VERSION" != "$NEW_VERSION" ]; then
  rm -rf $ISTIO_OLD
fi
commit_changes "$MANIFESTS_DIR" "Upgrade istio-cni to v.${COMMIT}" "."

echo "Synchronization completed successfully."