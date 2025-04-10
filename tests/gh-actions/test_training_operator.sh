#!/bin/bash
set -euo pipefail

KF_PROFILE=${1:-kubeflow-user-example-com}

for i in {1..10}; do
  kubectl api-resources | grep -q pytorchjob && break
  sleep 5
done

kubectl get pytorchjobs --all-namespaces 

sed 's/namespace: .*/namespace: '"$KF_PROFILE"'/g' tests/gh-actions/kf-objects/training_operator_job.yaml > /tmp/pytorch-job.yaml
kubectl apply -f /tmp/pytorch-job.yaml

sleep 5
kubectl wait --for=condition=Succeeded pytorchjob/pytorch-simple -n $KF_PROFILE --timeout=600s 