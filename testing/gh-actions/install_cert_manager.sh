#!/bin/bash
set -e
echo "Installing cert-manager ..."
cd common/cert-manager
kubectl create namespace cert-manager
kustomize build cert-manager/base | kubectl apply -f -
echo "Waiting for cert-manager to be ready ..."
kubectl wait --for=condition=ready pod -l 'app in (cert-manager,webhook)' --timeout=180s -n cert-manager
kustomize build kubeflow-issuer/base | kubectl apply -f -