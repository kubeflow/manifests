#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE

timeout 60 bash -c 'until kubectl get jobset pytorch-simple -n '"$KF_PROFILE"' >/dev/null 2>&1; do sleep 3; done' 2>/dev/null
kubectl get jobset pytorch-simple -n $KF_PROFILE
kubectl get pods -n $KF_PROFILE --show-labels
kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=180s

kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s