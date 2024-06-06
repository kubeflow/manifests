#!/bin/bash
set -e
echo "Installing Istio configured with external authorization..."
cd common/istio-1-21
kustomize build istio-crds/base | kubectl apply -f -
kustomize build istio-namespace/base | kubectl apply -f -
kustomize build istio-install/overlays/oauth2-proxy | kubectl apply -f -
cd -

echo "Waiting for all Istio Pods to become ready..."
kubectl wait --for=condition=Ready pods --all -n istio-system --timeout 300s

echo "Installing oauth2-proxy..."
cd common/oidc-client
kustomize build oauth2-proxy/overlays/m2m-self-signed/ | kubectl apply -f -
kubectl wait --for=condition=ready pod -l 'app.kubernetes.io/name=oauth2-proxy' --timeout=180s -n oauth2-proxy
