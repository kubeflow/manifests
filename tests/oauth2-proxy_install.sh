#!/bin/bash
set -e

echo "Installing oauth2-proxy..."
cd common/
kustomize build oauth2-proxy/overlays/m2m-dex-and-kind/ | kubectl apply -f -

echo "Waiting for all oauth2-proxy pods to become ready..."
kubectl wait --for=condition=Ready pod -l 'app.kubernetes.io/name=oauth2-proxy' --timeout=180s -n oauth2-proxy

echo "Waiting for all cluster-jwks-proxy pods to become ready..."
kubectl wait --for=condition=Ready pod -l 'app.kubernetes.io/name=cluster-jwks-proxy' --timeout=180s -n istio-system 
kubectl wait --for=condition=Available deployment -n oauth2-proxy oauth2-proxy --timeout=180s
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=180s '--field-selector=status.phase!=Succeeded'
