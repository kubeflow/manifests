#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

kubectl get crd jobsets.jobset.x-k8s.io
kubectl get service jobset-webhook-service -n kubeflow-system
kubectl get mutatingwebhookconfiguration jobset-mutating-webhook-configuration
kubectl get validatingwebhookconfiguration jobset-validating-webhook-configuration


kubectl wait --for=condition=Available deployment/jobset-controller-manager -n kubeflow-system --timeout=120s
kubectl wait --for=condition=Ready pod -l control-plane=controller-manager -n kubeflow-system --timeout=60s

sleep 10
kubectl get endpoints jobset-webhook-service -n kubeflow-system

TRAINER_POD=$(kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer -o jsonpath='{.items[0].metadata.name}')
if [ -z "$TRAINER_POD" ]; then
    echo "ERROR: Trainer pod not found"
    exit 1
fi


kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE
sleep 15

kubectl get deployment kubeflow-trainer-controller-manager -n kubeflow-system
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer
kubectl get clustertrainingruntimes torch-distributed
kubectl get trainjob pytorch-simple -n $KF_PROFILE -o yaml | tail -20

if ! kubectl get jobset pytorch-simple -n $KF_PROFILE >/dev/null 2>&1; then
    echo "ERROR: JobSet was not created by trainer controller"
    
    kubectl get all -n kubeflow-system | grep jobset 
    kubectl describe service jobset-webhook-service -n kubeflow-system
    kubectl logs -n kubeflow-system -l control-plane=controller-manager --tail=20
    kubectl get mutatingwebhookconfiguration jobset-mutating-webhook-configuration -o yaml
    
    exit 1
fi

kubectl get pods -n $KF_PROFILE --show-labels
kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=180s
kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s