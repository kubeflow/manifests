#!/bin/bash
set -euo pipefail
echo "Installing training operator ..."

cd apps/training-operator/upstream
kustomize build overlays/kubeflow | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=600s \
  --field-selector=status.phase!=Succeeded
cd -
