#!/bin/bash
set -e

echo "Installing oauth2-proxy..."
cd common/
kustomize build oauth2-proxy/overlays/m2m-dex-and-kind/ | kubectl apply -f -

echo "Waiting for all oauth2-proxy pods to become ready..."
kubectl wait --for=condition=ready pod -l 'app.kubernetes.io/name=oauth2-proxy' --timeout=180s -n oauth2-proxy

echo "Waiting for all cluster-jwks-proxy pods to become ready..."
kubectl wait --for=condition=ready pod -l 'app.kubernetes.io/name=cluster-jwks-proxy' --timeout=180s -n istio-system