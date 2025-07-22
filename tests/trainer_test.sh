#!/bin/bash
set -euo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

kubectl get deployment -n kubeflow-system kubeflow-trainer-controller-manager >/dev/null
kubectl get deployment -n kubeflow-system jobset-controller-manager >/dev/null
kubectl get clustertrainingruntimes >/dev/null


kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE

for i in {1..60}; do
    if kubectl get jobset pytorch-simple -n $KF_PROFILE >/dev/null 2>&1; then
        break
    fi
    if [ $i -eq 60 ]; then
        kubectl get trainjob pytorch-simple -n $KF_PROFILE -o yaml 
        kubectl get events -n $KF_PROFILE --field-selector involvedObject.name=pytorch-simple 
        kubectl logs -n kubeflow-system -l app.kubernetes.io/name=trainer --tail=50 
        exit 1
    fi
    sleep 3
done

kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=180s

kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s
