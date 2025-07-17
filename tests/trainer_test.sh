#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE

kubectl wait --for=jsonpath='{.status.conditions[0].type}=Created' trainjob/pytorch-simple -n $KF_PROFILE --timeout=60s
kubectl get jobset pytorch-simple -n $KF_PROFILE
kubectl get pods -n $KF_PROFILE --show-labels
kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=180s

kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s