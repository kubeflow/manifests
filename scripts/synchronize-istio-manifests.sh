#!/usr/bin/env bash
# This script helps to create a PR to update the unified Istio manifests

SCRIPT_DIRECTORY=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${SCRIPT_DIRECTORY}/library.sh"

setup_error_handling

COMPONENT_NAME="istio"
COMMIT="1.28.0"  # Update this for new versions
SOURCE_DIRECTORY=${SOURCE_DIRECTORY:=/tmp/${COMPONENT_NAME}}
BRANCH_NAME=${BRANCH_NAME:=${COMPONENT_NAME}-${COMMIT?}}

# Path configurations
MANIFESTS_DIRECTORY=$(dirname $SCRIPT_DIRECTORY)
ISTIO_DIRECTORY=$MANIFESTS_DIRECTORY/common/${COMPONENT_NAME}

create_branch "$BRANCH_NAME"

echo "Checking out in $SOURCE_DIRECTORY to $COMMIT..."
mkdir -p $SOURCE_DIRECTORY
cd $SOURCE_DIRECTORY
if [ ! -d "istio-${COMMIT}" ]; then
    wget "https://github.com/istio/istio/releases/download/${COMMIT}/istio-${COMMIT}-linux-amd64.tar.gz"
    tar xvfz istio-${COMMIT}-linux-amd64.tar.gz
fi

ISTIOCTL=$SOURCE_DIRECTORY/istio-${COMMIT}/bin/istioctl
cd $ISTIO_DIRECTORY

echo "Generating CNI manifests (default)..."
$ISTIOCTL manifest generate -f profile.yaml -f profile-overlay.yaml \
  --set components.cni.enabled=true \
  --set components.cni.namespace=kube-system > dump.yaml
./split-istio-packages -f dump.yaml
mv $ISTIO_DIRECTORY/crd.yaml $ISTIO_DIRECTORY/istio-crds/base/
mv $ISTIO_DIRECTORY/install.yaml $ISTIO_DIRECTORY/istio-install/base/
mv $ISTIO_DIRECTORY/cluster-local-gateway.yaml $ISTIO_DIRECTORY/cluster-local-gateway/base/
rm dump.yaml

echo "Generating ztunnel manifests (ambient mode)..."
$ISTIOCTL manifest generate -f profile.yaml -f profile-overlay.yaml \
  --set components.cni.enabled=true \
  --set components.ztunnel.enabled=true > dump-ztunnel.yaml
./split-istio-packages -f dump-ztunnel.yaml
mv $ISTIO_DIRECTORY/ztunnel.yaml $ISTIO_DIRECTORY/istio-install/components/ambient-mode/
rm dump-ztunnel.yaml crd.yaml install.yaml cluster-local-gateway.yaml

check_uncommitted_changes

echo "Updating tag in istio-sidecar-injector-patch.yaml..."
sed -i "s/\"tag\": \".*\"/\"tag\": \"$COMMIT\"/" $ISTIO_DIRECTORY/istio-install/base/patches/istio-sidecar-injector-patch.yaml

SOURCE_TEXT="\[.*\](https://github.com/istio/istio/releases/tag/.*)"
DESTINATION_TEXT="\[$COMMIT\](https://github.com/istio/istio/releases/tag/$COMMIT)"

update_readme "$MANIFESTS_DIRECTORY" "$SOURCE_TEXT" "$DESTINATION_TEXT"

commit_changes "$MANIFESTS_DIRECTORY" "Upgrade istio to v.${COMMIT}" "."

echo "Synchronization completed successfully."
