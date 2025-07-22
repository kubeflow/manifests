#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

kubectl get crd jobsets.jobset.x-k8s.io || echo "JobSet CRD not found!"
kubectl get service jobset-webhook-service -n kubeflow-system || echo "JobSet webhook service not found!"
kubectl get deployment -n kubeflow-system -l app.kubernetes.io/name=jobset || echo "JobSet deployment not found!"
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=jobset || echo "JobSet pods not found!"

kubectl get mutatingwebhookconfiguration | grep jobset || echo "JobSet mutating webhook not found!"
kubectl get validatingwebhookconfiguration | grep jobset || echo "JobSet validating webhook not found!"

if kubectl get deployment -n kubeflow-system -l app.kubernetes.io/name=jobset >/dev/null 2>&1; then
    echo "Waiting for JobSet webhook to be ready..."
    kubectl wait --for=condition=Available deployment -n kubeflow-system -l app.kubernetes.io/name=jobset --timeout=120s
    
    echo "Waiting for JobSet webhook endpoints..."
    kubectl wait --for=condition=Ready pod -n kubeflow-system -l app.kubernetes.io/name=jobset --timeout=60s
    sleep 5 
fi

kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE

sleep 10

kubectl get deployment kubeflow-trainer-controller-manager -n kubeflow-system
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer

kubectl get clustertrainingruntimes torch-distributed 

kubectl logs -n kubeflow-system -l app.kubernetes.io/name=trainer --tail=20

if kubectl get jobset pytorch-simple -n $KF_PROFILE >/dev/null 2>&1; then
    kubectl get pods -n $KF_PROFILE --show-labels
    kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=180s
    kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s
else
    echo "ERROR: JobSet was not created by trainer controller"
    echo "This is likely due to JobSet webhook service not being available"
    exit 1
fi