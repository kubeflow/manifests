#!/bin/bash
set -euo pipefail

cd applications/training-operator/upstream
kustomize build overlays/kubeflow | kubectl apply --server-side --force-conflicts -f -

kubectl wait --for=condition=Available deployment/training-operator -n kubeflow --timeout=180s


kubectl get deployment -n kubeflow training-operator
kubectl get pods -n kubeflow -l app=training-operator
kubectl get crd | grep -E 'tfjobs.kubeflow.org|pytorchjobs.kubeflow.org'

cd -