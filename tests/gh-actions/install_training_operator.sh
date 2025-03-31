#!/bin/bash
set -euo pipefail
echo "Installing training operator ..."

cd apps/training-operator/upstream
kustomize build overlays/kubeflow | kubectl apply --server-side --force-conflicts -f -
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=600s \
  --field-selector=status.phase!=Succeeded

# Verify Training Operator components
echo "Verifying Training Operator components..."
kubectl wait --for=condition=Available deployment/training-operator -n kubeflow --timeout=300s || echo "Training operator not yet available"

# Check CRDs are properly installed
echo "Verifying Training Operator CRDs..."
kubectl get crd | grep -E 'tfjobs.kubeflow.org|pytorchjobs.kubeflow.org' || echo "Some training operator CRDs may not be available"

# Display Training Operator status
echo "Training Operator status:"
kubectl get deployment -n kubeflow training-operator

cd -
