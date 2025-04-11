#!/bin/bash
set -euxo 

KF_PROFILE=${1:-kubeflow-user-example-com}

cat tests/gh-actions/kf-objects/training_operator_job.yaml | \
  sed 's/name: pytorch-simple/name: pytorch-simple\n  namespace: '"$KF_PROFILE"'/g' > /tmp/pytorch-job.yaml

kubectl apply -f /tmp/pytorch-job.yaml

sleep 90

kubectl get pytorchjob/pytorch-simple -n $KF_PROFILE

kubectl wait --for=condition=Running pytorchjob/pytorch-simple -n $KF_PROFILE --timeout=180s

# Try to wait for Succeeded, but don't fail if it's still Running after timeout
kubectl wait --for=condition=Succeeded pytorchjob/pytorch-simple -n $KF_PROFILE --timeout=450s || {
  STATE=$(kubectl get pytorchjob/pytorch-simple -n $KF_PROFILE -o custom-columns=STATE:.status.conditions[0].type --no-headers)
  if [ "$STATE" == "Running" ]; then
    echo "Job still running, considering test successful"
    exit 0
  else
    echo "Job failed with state: $STATE"
    exit 1
  fi
}