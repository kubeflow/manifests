#!/bin/bash
set -euxo

kustomize build common/user-namespace/base | kubectl apply -f -
sleep 30

PROFILE_CONTROLLER_POD=$(kubectl get pods -n kubeflow -o json | jq -r '.items[] | select(.metadata.name | startswith("profiles-deployment")) | .metadata.name')
if [[ -z "$PROFILE_CONTROLLER_POD" ]]; then
  echo "Error: profiles-deployment pod not found in kubeflow namespace."
  exit 1
fi

kubectl logs -n kubeflow "$PROFILE_CONTROLLER_POD"

KF_PROFILE=kubeflow-user-example-com
kubectl -n $KF_PROFILE get pods,configmaps,secrets 