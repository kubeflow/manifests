#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

cat tests/gh-actions/kf-objects/training_operator_job.yaml | \
sed 's/name: pytorch-simple/name: pytorch-simple\n  namespace: '"$KF_PROFILE"'/g' > /tmp/pytorch-job.yaml
kubectl apply -f /tmp/pytorch-job.yaml

kubectl wait --for=jsonpath='{.status.conditions[0].type}=Created' pytorchjob.kubeflow.org/pytorch-simple -n $KF_PROFILE --timeout=60s

echo "Checking for PyTorch pods with the following commands:"
echo "kubectl get pods -n $KF_PROFILE -o wide"
kubectl get pods -n $KF_PROFILE -o wide
echo "kubectl get pods -n $KF_PROFILE --show-labels"
kubectl get pods -n $KF_PROFILE --show-labels

kubectl wait --for=condition=Ready pod -l pytorch-replica-type=worker -n $KF_PROFILE --timeout=180s

kubectl wait --for=condition=Succeeded pytorchjob/pytorch-simple -n $KF_PROFILE --timeout=450s