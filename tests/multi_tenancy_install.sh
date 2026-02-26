#!/bin/bash
set -euxo pipefail

echo "Installing Profiles Controller with PSS (Pod Security Standards)"
kustomize build applications/profiles/pss | kubectl apply -f -
kubectl -n kubeflow rollout status deployment/profiles-deployment --timeout 180s
kubectl -n kubeflow wait --for=condition=Ready pods -l kustomize.component=profiles --timeout 180s

echo "Installing Multitenancy Kubeflow Roles"
kustomize build common/kubeflow-roles/base | kubectl apply -f -

