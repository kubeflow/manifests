#!/bin/bash
set -e
echo "Installing cert-manager ..."
cd common/cert-manager
kubectl create namespace cert-manager
kustomize build base | kubectl apply -f -

echo "Waiting for cert-manager to be ready ..."
kubectl wait --for=condition=Ready pod -l 'app in (cert-manager,webhook)' --timeout=180s -n cert-manager
kubectl wait --for=jsonpath='{.subsets[0].addresses[0].targetRef.kind}'=Pod endpoints -l 'app in (cert-manager,webhook)' --timeout=180s -n cert-manager

echo "Deploy clusterissuer.cert-manager.io/kubeflow-self-signing-issuer"
kustomize build kubeflow-issuer/base | kubectl apply -f -