#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE

sleep 10

kubectl get deployment kubeflow-trainer-controller-manager -n kubeflow-system
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer

kubectl get clustertrainingruntimes torch-distributed 

kubectl logs -n kubeflow-system -l app.kubernetes.io/name=trainer --tail=20

kubectl get trainjob pytorch-simple -n $KF_PROFILE -o yaml

kubectl get jobset pytorch-simple -n $KF_PROFILE || echo "JobSet not found yet"

if kubectl get jobset pytorch-simple -n $KF_PROFILE >/dev/null 2>&1; then
    kubectl get pods -n $KF_PROFILE --show-labels
    kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=180s
    kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s
else
    echo "ERROR: JobSet was not created by trainer controller"
    exit 1
fi