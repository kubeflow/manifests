#!/bin/bash
set -euo pipefail

cd apps/training-operator/upstream
kustomize build overlays/kubeflow | kubectl apply --server-side --force-conflicts -f -

kubectl wait --for=condition=Available deployment/training-operator -n kubeflow --timeout=180s

for i in {1..60}; do
  if kubectl get crd pytorchjobs.kubeflow.org >/dev/null 2>&1 && 
     kubectl api-resources | grep -q "pytorchjobs.*kubeflow"; then
    break
  fi
  
  [ $i -eq 60 ] && exit 1
  sleep 2
done

kubectl get deployment -n kubeflow training-operator
kubectl get pods -n kubeflow -l app=training-operator
kubectl get crd | grep -E 'tfjobs.kubeflow.org|pytorchjobs.kubeflow.org'

cd -
