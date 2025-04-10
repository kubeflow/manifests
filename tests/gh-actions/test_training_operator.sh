#!/bin/bash
set -euo pipefail

KF_PROFILE=${1:-kubeflow-user-example-com}

for i in {1..30}; do
  kubectl api-resources | grep -q "pytorchjobs.*kubeflow" && break
  sleep 5
done

JOB_YAML=$(sed 's/namespace: .*/namespace: '"$KF_PROFILE"'/g' tests/gh-actions/kf-objects/training_operator_job.yaml)
echo "$JOB_YAML" | kubectl apply -f -

sleep 30

kubectl get pytorchjob/pytorch-simple -n $KF_PROFILE


kubectl wait --for=condition=Succeeded pytorchjob/pytorch-simple -n $KF_PROFILE --timeout=600s 