#!/bin/bash
set -euo pipefail

cd applications/trainer/upstream

kustomize build base/crds | kubectl apply --server-side --force-conflicts -f -

sleep 5
kubectl wait --for condition=established crd/trainjobs.trainer.kubeflow.org --timeout=60s

kustomize build overlays/manager | kubectl apply --server-side --force-conflicts -f -

cd ../../../

kustomize build common/kubeflow-system-namespace/base | kubectl apply -f -

kustomize build common/networkpolicies/kubeflow-system | kubectl apply -f -

cd applications/trainer/upstream

kubectl wait --for=condition=Available deployment/kubeflow-trainer-controller-manager -n kubeflow-system --timeout=180s

kubectl wait --for=condition=Available deployment/jobset-controller-manager -n kubeflow-system --timeout=180s
sleep 15

kubectl apply -f tests/trainer_rbac_patch.yaml

kustomize build overlays/runtimes | kubectl apply --server-side --force-conflicts -f -

kubectl apply -f overlays/kubeflow-platform/kubeflow-trainer-roles.yaml


kubectl get deployment -n kubeflow-system kubeflow-trainer-controller-manager
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer
kubectl get crd | grep -E 'trainer.kubeflow.org'
kubectl get clustertrainingruntimes

cd -