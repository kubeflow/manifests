#!/bin/bash
set -e

echo "Creating Kubeflow user profile..."
kustomize build common/user-namespace/base | kubectl apply -f -
sleep 30

# Verify profile controller is working
PROFILE_CONTROLLER_POD=$(kubectl get pods -n kubeflow -o json | jq -r '.items[] | select(.metadata.name | startswith("profiles-deployment")) | .metadata.name')
if [[ -z "$PROFILE_CONTROLLER_POD" ]]; then
  echo "Error: profiles-deployment pod not found in kubeflow namespace."
  exit 1
fi

echo "Profile controller logs:"
kubectl logs -n kubeflow "$PROFILE_CONTROLLER_POD"

# Set profile name and check resources
KF_PROFILE=kubeflow-user-example-com
echo "Checking resources in profile namespace $KF_PROFILE:"
kubectl -n $KF_PROFILE get pods,configmaps,secrets

echo "Kubeflow profile creation completed." 