#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

kubectl get crd jobsets.jobset.x-k8s.io || echo "JobSet CRD not found!"
kubectl get service jobset-webhook-service -n kubeflow-system || echo "JobSet webhook service not found!"

kubectl get deployment -A | grep jobset || echo "JobSet deployment not found in any namespace!"
kubectl get pods -A | grep jobset || echo "JobSet pods not found in any namespace!"

echo "=== Checking JobSet webhook configuration ==="
kubectl get mutatingwebhookconfiguration | grep jobset || echo "JobSet mutating webhook not found!"
kubectl get validatingwebhookconfiguration | grep jobset || echo "JobSet validating webhook not found!"

if ! kubectl get deployment -A | grep -q jobset; then
    echo "=== JobSet deployment missing, checking if we can install it ==="
    if kubectl get service jobset-webhook-service -n kubeflow-system >/dev/null 2>&1; then
        echo "JobSet service exists but deployment is missing - this indicates a broken installation"
        echo "Attempting to check JobSet webhook readiness differently..."
        
        if kubectl get endpoints jobset-webhook-service -n kubeflow-system -o jsonpath='{.subsets[*].addresses[*].ip}' | grep -q .; then
            echo "JobSet webhook endpoints exist"
        else
            echo "WARNING: JobSet webhook service has no endpoints - this will cause webhook failures"
        fi
    fi
fi

kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE

sleep 15  # Give more time for processing

kubectl get deployment kubeflow-trainer-controller-manager -n kubeflow-system
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer

kubectl get clustertrainingruntimes torch-distributed 

kubectl logs -n kubeflow-system -l app.kubernetes.io/name=trainer --tail=30

kubectl get trainjob pytorch-simple -n $KF_PROFILE -o yaml | tail -20

if kubectl get jobset pytorch-simple -n $KF_PROFILE >/dev/null 2>&1; then
    echo "SUCCESS: JobSet was created!"
    kubectl get pods -n $KF_PROFILE --show-labels
    kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=180s
    kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s
else
    echo "ERROR: JobSet was not created by trainer controller"
    echo "This is likely due to JobSet webhook deployment being missing"
    
    kubectl get all -n kubeflow-system | grep jobset || echo "No JobSet resources in kubeflow-system"
    kubectl describe service jobset-webhook-service -n kubeflow-system || echo "Cannot describe JobSet service"
    
    exit 1
fi