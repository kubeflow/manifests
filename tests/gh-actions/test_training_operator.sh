#!/bin/bash
set -euxo 

KF_PROFILE=${1:-kubeflow-user-example-com}

cat tests/gh-actions/kf-objects/training_operator_job.yaml | \
  sed 's/name: pytorch-simple/name: pytorch-simple\n  namespace: '"$KF_PROFILE"'/g' > /tmp/pytorch-job.yaml

kubectl apply -f /tmp/pytorch-job.yaml

sleep 90

kubectl get pytorchjob/pytorch-simple -n $KF_PROFILE

kubectl wait --for=condition=Running pytorchjob/pytorch-simple -n $KF_PROFILE --timeout=180s

kubectl wait --for=condition=Succeeded pytorchjob/pytorch-simple -n $KF_PROFILE --timeout=450s 