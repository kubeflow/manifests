#!/bin/bash
set -euo pipefail

cd apps/training-operator/upstream
kustomize build overlays/kubeflow | kubectl apply --server-side --force-conflicts -f -

kubectl wait --for=condition=Available deployment/training-operator -n kubeflow --timeout=120s

for i in {1..30}; do
  kubectl get crd | grep -q 'pytorchjobs.kubeflow.org' && break
  sleep 2
done

kubectl get deployment -n kubeflow training-operator
kubectl get pods -n kubeflow -l app=training-operator
kubectl get crd | grep -E 'tfjobs.kubeflow.org|pytorchjobs.kubeflow.org'

cd -
