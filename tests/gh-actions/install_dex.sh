#!/bin/bash
set -e

echo "Installing Dex..."
kustomize build ./common/dex/overlays/oauth2-proxy | kubectl apply -f -

echo "Waiting for pods in auth namespace to become ready..."
kubectl wait --for=condition=Ready pods --all --timeout=180s -n auth

echo "Dex installation completed." 