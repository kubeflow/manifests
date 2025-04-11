#!/bin/bash
set -euo pipefail
echo "Installing training operator ..."

cd apps/training-operator/upstream
kustomize build overlays/kubeflow | kubectl apply --server-side --force-conflicts -f -
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=600s \
  --field-selector=status.phase!=Succeeded
kubectl wait --for=condition=Available deployment/training-operator -n kubeflow --timeout=10s
kubectl get crd | grep -E 'tfjobs.kubeflow.org|pytorchjobs.kubeflow.org'
kubectl get deployment -n kubeflow training-operator
cd -
