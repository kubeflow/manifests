#!/bin/bash
set -euo pipefail

cd applications/trainer/upstream

kustomize build base/crds | kubectl apply --server-side --force-conflicts -f -

sleep 5
kubectl wait --for condition=established crd/trainjobs.trainer.kubeflow.org --timeout=60s

kustomize build overlays/kubeflow-platform | kubectl apply --server-side --force-conflicts -f -
kubectl wait --for=condition=Available deployment/kubeflow-trainer-controller-manager -n kubeflow --timeout=180s

kubectl get deployment -n kubeflow kubeflow-trainer-controller-manager
kubectl get pods -n kubeflow -l app.kubernetes.io/name=trainer
kubectl get crd | grep -E 'trainer.kubeflow.org'
kubectl get clustertrainingruntimes

cd -