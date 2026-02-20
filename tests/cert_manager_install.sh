#!/bin/bash
set -euo pipefail
echo "Installing cert-manager ..."
cd common/cert-manager
kubectl create namespace cert-manager || true
kustomize build overlays/kubeflow | kubectl apply -f -
echo "Waiting for cert-manager-webhook to be ready ..."
kubectl wait --for=condition=Ready pod -l 'app=webhook' --timeout=180s -n cert-manager

echo "Waiting for all cert-manager components to be ready ..."
kubectl wait --for=condition=Ready pod -l 'app.kubernetes.io/instance=cert-manager' --timeout=180s -n cert-manager
kubectl wait --for=jsonpath='{.subsets[0].addresses[0].targetRef.kind}'=Pod endpoints -l 'app in (cert-manager,webhook)' --timeout=180s -n cert-manager
