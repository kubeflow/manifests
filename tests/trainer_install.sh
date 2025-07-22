#!/bin/bash
set -euo pipefail

cd applications/trainer/upstream

kustomize build base/crds | kubectl apply --server-side --force-conflicts -f -

sleep 5
kubectl wait --for condition=established crd/trainjobs.trainer.kubeflow.org --timeout=60s

kustomize build overlays/manager | kubectl apply --server-side --force-conflicts -f -
kubectl wait --for=condition=Available deployment/kubeflow-trainer-controller-manager -n kubeflow-system --timeout=180s


kustomize build overlays/runtimes | kubectl apply --server-side --force-conflicts -f -


kubectl apply -f overlays/kubeflow-platform/kubeflow-trainer-roles.yaml

cd -

kubectl apply -f common/networkpolicies/base/trainer-webhook-kubeflow-system.yaml
kubectl apply -f common/networkpolicies/base/default-allow-same-namespace-kubeflow-system.yaml

kubectl get deployment -n kubeflow-system kubeflow-trainer-controller-manager
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer
kubectl get crd | grep -E 'trainer.kubeflow.org'
kubectl get clustertrainingruntimes