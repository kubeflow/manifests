#!/bin/bash
set -e

error_exit() {
    echo "Error occurred in script at line: ${1}."
    exit 1
}

trap 'error_exit $LINENO' ERR

echo "Install KinD..."
sudo swapoff -a

# This conditional helps running GH Workflows through
# [act](https://github.com/nektos/act)
if [ -e /swapfile ]; then
    sudo rm -f /swapfile
    sudo mkdir -p /tmp/etcd
    sudo mount -t tmpfs tmpfs /tmp/etcd
fi

{
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64
    chmod +x ./kind
    sudo mv kind /usr/local/bin
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
  image: kindest/node:v1.32.2@sha256:f226345927d7e348497136874b6d207e0b32cc52154ad8323129352923a3142f
- role: worker
  image: kindest/node:v1.32.2@sha256:f226345927d7e348497136874b6d207e0b32cc52154ad8323129352923a3142f
- role: worker
  image: kindest/node:v1.32.2@sha256:f226345927d7e348497136874b6d207e0b32cc52154ad8323129352923a3142f
" | kind create cluster --config - --wait 120s

kubectl cluster-info

echo "Install Kustomize ..."
{
    curl --silent --location --remote-name "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.4.3/kustomize_v5.4.3_linux_amd64.tar.gz"
    tar -xzvf kustomize_v5.4.3_linux_amd64.tar.gz
    chmod a+x kustomize
    sudo mv kustomize /usr/local/bin/kustomize
} || { echo "Failed to install Kustomize"; exit 1; }
