#!/bin/bash
set -e
echo "Installing Istio ..."
cd common/istio-1-21
kustomize build istio-crds/base | kubectl apply -f -
kustomize build istio-namespace/base | kubectl apply -f -
kustomize build istio-install/base | kubectl apply -f -

echo "Waiting for all Istio Pods to become ready..."
kubectl wait --for=condition=Ready pods --all -n istio-system --timeout 300s
