#!/bin/bash
set -e
echo "Installing Istio-cni (with ExtAuthZ from oauth2-proxy) ..."
cd common/istio
kustomize build istio-crds/base | kubectl apply -f -
kustomize build istio-namespace/base | kubectl apply -f -
kustomize build istio-install/overlays/oauth2-proxy | kubectl apply -f -

echo "Waiting for all Istio Pods to become ready..."
kubectl wait --for=condition=Ready pods --all -n istio-system --timeout 180s
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=180s '--field-selector=status.phase!=Succeeded'
