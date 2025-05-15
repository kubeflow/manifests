#!/bin/bash
set -euxo
kustomize build common/user-namespace/base | kubectl apply -f -
sleep 30 # Let the profile controler reconcile the namespace
PROFILE_CONTROLLER_POD=$(kubectl get pods -n kubeflow -o json | jq -r '.items[] | select(.metadata.name | startswith("profiles-deployment")) | .metadata.name')
kubectl logs -n kubeflow "$PROFILE_CONTROLLER_POD"
KF_PROFILE=kubeflow-user-example-com
kubectl -n $KF_PROFILE get pods,configmaps,secrets
