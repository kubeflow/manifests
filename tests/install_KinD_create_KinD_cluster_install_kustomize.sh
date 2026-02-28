#!/bin/bash
set -e

KIND_VERSION="v0.31.0"
KUSTOMIZE_VERSION="v5.8.1"

error_exit() {
    echo "Error occurred in script at line: ${1}."
    exit 1
}

run_with_elevated_privileges() {
    if [ "$(id -u)" -eq 0 ]; then
        "$@"
    else
        sudo "$@"
    fi
}

trap 'error_exit $LINENO' ERR

echo "Install KinD..."

{
    curl -Lo ./kind-linux-amd64 https://kind.sigs.k8s.io/dl/$KIND_VERSION/kind-linux-amd64
    curl -Lo ./kind-linux-amd64.sha256sum https://kind.sigs.k8s.io/dl/$KIND_VERSION/kind-linux-amd64.sha256sum
    if ! sha256sum --check kind-linux-amd64.sha256sum; then
       echo "Failed to verify KinD checksums"
       exit 1
    fi
    chmod +x ./kind-linux-amd64
    run_with_elevated_privileges mv kind-linux-amd64 /usr/local/bin/kind
} || { echo "Failed to install KinD"; exit 1; }


echo "Creating KinD cluster ..."
echo "
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
# Configure registry for KinD.
containerdConfigPatches:
- |-
  [plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"REGISTRY_NAME:REGISTRY_PORT\"]
    endpoint = [\"http://REGISTRY_NAME:REGISTRY_PORT\"]
# This is needed in order to support projected volumes with service account tokens.
# See: https://kubernetes.slack.com/archives/CEKK1KTN2/p1600268272383600
kubeadmConfigPatches:
  - |
    apiVersion: kubeadm.k8s.io/v1beta2
    kind: ClusterConfiguration
    metadata:
      name: config
    apiServer:
      extraArgs:
        \"service-account-issuer\": \"https://kubernetes.default.svc\"
        \"service-account-signing-key-file\": \"/etc/kubernetes/pki/sa.key\"
nodes:
- role: control-plane
  image: kindest/node:v1.35.0@sha256:452d707d4862f52530247495d180205e029056831160e22870e37e3f6c1ac31f
- role: worker
  image: kindest/node:v1.35.0@sha256:452d707d4862f52530247495d180205e029056831160e22870e37e3f6c1ac31f
- role: worker
  image: kindest/node:v1.35.0@sha256:452d707d4862f52530247495d180205e029056831160e22870e37e3f6c1ac31f
" | kind create cluster --config - --wait 120s

kubectl cluster-info

echo "Install Kustomize ..."
{
    KUSTOMIZE_ASSET="kustomize_${KUSTOMIZE_VERSION}_linux_amd64.tar.gz"
    curl --fail --show-error --silent --location --remote-name "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/${KUSTOMIZE_ASSET}"
    curl --fail --show-error --silent --location "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/checksums.txt" | grep "  ${KUSTOMIZE_ASSET}$" > checksums.txt
    if [ "$(wc -l < checksums.txt)" -ne 1 ]; then
       echo "Failed to verify Kustomize checksums: expected exactly one checksum entry for ${KUSTOMIZE_ASSET}"
       exit 1
    fi
    if ! sha256sum --check checksums.txt; then
       echo "Failed to verify Kustomize checksums"
       exit 1
    fi
    tar -xzvf "${KUSTOMIZE_ASSET}"
    chmod a+x kustomize
    run_with_elevated_privileges mv kustomize /usr/local/bin/kustomize
} || { echo "Failed to install Kustomize"; exit 1; }
