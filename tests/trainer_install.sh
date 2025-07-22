#!/bin/bash
set -euo pipefail

cd applications/trainer/upstream

kustomize build base/crds | kubectl apply --server-side --force-conflicts -f -

sleep 5
kubectl wait --for condition=established crd/trainjobs.trainer.kubeflow.org --timeout=60s

kustomize build overlays/manager | kubectl apply --server-side --force-conflicts -f -
kubectl wait --for=condition=Available deployment/kubeflow-trainer-controller-manager -n kubeflow-system --timeout=180s

kubectl get crd jobsets.jobset.x-k8s.io 

if kubectl get deployment -A | grep -q jobset; then
    echo "JobSet deployment found:"
    kubectl get deployment -A | grep jobset
    
    JOBSET_NAMESPACE=$(kubectl get deployment -A | grep jobset | awk '{print $1}' | head -1)
    JOBSET_DEPLOYMENT=$(kubectl get deployment -A | grep jobset | awk '{print $2}' | head -1)
    
    echo "Waiting for JobSet deployment $JOBSET_DEPLOYMENT in namespace $JOBSET_NAMESPACE to be ready..."
    kubectl wait --for=condition=Available deployment/$JOBSET_DEPLOYMENT -n $JOBSET_NAMESPACE --timeout=180s
else
    echo "WARNING: JobSet deployment not found in any namespace!"
fi

kustomize build overlays/runtimes | kubectl apply --server-side --force-conflicts -f -

kubectl apply -f overlays/kubeflow-platform/kubeflow-trainer-roles.yaml

cd -

kubectl apply -f common/networkpolicies/base/trainer-webhook-kubeflow-system.yaml
kubectl apply -f common/networkpolicies/base/default-allow-same-namespace-kubeflow-system.yaml
kubectl apply -f common/networkpolicies/base/jobset-webhook-kubeflow-system.yaml

kubectl get deployment -n kubeflow-system kubeflow-trainer-controller-manager
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer
kubectl get crd | grep -E 'trainer.kubeflow.org'
kubectl get clustertrainingruntimes


kubectl rollout restart deployment/jobset-controller-manager -n kubeflow-system
kubectl wait --for=condition=Available deployment/jobset-controller-manager -n kubeflow-system --timeout=120s