#!/bin/bash
set -e
echo "Installing Istio configured with external authorization..."
cd common/istio-1-22
kustomize build istio-crds/base | kubectl apply -f -
kustomize build istio-namespace/base | kubectl apply -f -
kustomize build istio-install/overlays/oauth2-proxy | kubectl apply -f -
cd -

echo "Waiting for all Istio Pods to become ready..."
kubectl wait --for=condition=Ready pods --all -n istio-system --timeout=300s \
  --field-selector=status.phase!=Succeeded

echo "Installing oauth2-proxy..."
cd common/
kustomize build oauth2-proxy/overlays/m2m-dex-and-kind/ | kubectl apply -f -

echo "Waiting for all oauth2-proxy pods to become ready..."
kubectl wait --for=condition=ready pod -l 'app.kubernetes.io/name=oauth2-proxy' --timeout=180s -n oauth2-proxy

echo "Waiting for all cluster-jwks-proxy pods to become ready..."
kubectl wait --for=condition=ready pod -l 'app.kubernetes.io/name=cluster-jwks-proxy' --timeout=180s -n istio-system