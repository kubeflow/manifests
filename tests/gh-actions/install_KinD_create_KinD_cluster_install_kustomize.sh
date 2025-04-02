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
cat << EOF > kind-config.yaml
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
# Configure registry for KinD.
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."REGISTRY_NAME:REGISTRY_PORT"]
    endpoint = ["http://REGISTRY_NAME:REGISTRY_PORT"]
# This is needed in order to support projected volumes with service account tokens.
kubeadmConfigPatches:
  - |
    apiVersion: kubeadm.k8s.io/v1beta2
    kind: ClusterConfiguration
    metadata:
      name: config
    apiServer:
      extraArgs:
        "service-account-issuer": "https://kubernetes.default.svc"
        "service-account-signing-key-file": "/etc/kubernetes/pki/sa.key"
nodes:
- role: control-plane
  image: kindest/node:v1.27.3@sha256:3966ac761ae0136263ffdb6cfd4db23ef8a83cba8a463690e98317add2c9ba72
  kubeadmConfigPatches:
    - |
      kind: InitConfiguration
      nodeRegistration:
        kubeletExtraArgs:
          node-labels: "ingress-ready=true"
  extraPortMappings:
    - containerPort: 80
      hostPort: 80
      protocol: TCP
    - containerPort: 443
      hostPort: 443
      protocol: TCP
- role: worker
  image: kindest/node:v1.27.3@sha256:3966ac761ae0136263ffdb6cfd4db23ef8a83cba8a463690e98317add2c9ba72
- role: worker
  image: kindest/node:v1.27.3@sha256:3966ac761ae0136263ffdb6cfd4db23ef8a83cba8a463690e98317add2c9ba72
EOF

# Create the cluster with the new config
kind create cluster --name "${KIND_CLUSTER_NAME:-kubeflow}" --config kind-config.yaml

# Verify cluster is running
echo "Verifying cluster status..."
kubectl cluster-info
kubectl get nodes

echo "Install Kustomize ..."
{
    curl --silent --location --remote-name "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.4.3/kustomize_v5.4.3_linux_amd64.tar.gz"
    tar -xzvf kustomize_v5.4.3_linux_amd64.tar.gz
    chmod a+x kustomize
    sudo mv kustomize /usr/local/bin/kustomize
} || { echo "Failed to install Kustomize"; exit 1; }

# Save cluster configuration for other jobs
mkdir -p cluster-config
kind get kubeconfig --name="${KIND_CLUSTER_NAME:-kubeflow}" > cluster-config/kind-config
kubectl cluster-info > cluster-config/cluster-info.txt
kubectl get nodes -o wide > cluster-config/nodes.txt
kubectl version --output=json > cluster-config/version.json
