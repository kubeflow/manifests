#!/bin/bash
set -e
echo "Installing Istio (with ExtAuthZ from oauth2-proxy) ..."
cd common/istio-cni-1-23
kustomize build istio-crds/base | kubectl apply -f -
kustomize build istio-namespace/base | kubectl apply -f -
kustomize build istio-install/overlays/oauth2-proxy | kubectl apply -f -

echo "Waiting for all Istio Pods to become ready..."
kubectl wait --for=condition=Ready pods --all -n istio-system --timeout 300s
