#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

cat tests/training_operator_job.yaml | \
sed 's/name: pytorch-simple/name: pytorch-simple\n  namespace: '"$KF_PROFILE"'/g' > /tmp/pytorch-job.yaml
kubectl apply -f /tmp/pytorch-job.yaml

kubectl wait --for=jsonpath='{.status.conditions[0].type}=Created' pytorchjob.kubeflow.org/pytorch-simple -n $KF_PROFILE --timeout=60s

kubectl get pods -n $KF_PROFILE --show-labels

kubectl wait --for=condition=Ready pod -l training.kubeflow.org/replica-type=worker -n $KF_PROFILE --timeout=240s

kubectl wait --for=condition=Succeeded pytorchjob/pytorch-simple -n $KF_PROFILE --timeout=7200s
