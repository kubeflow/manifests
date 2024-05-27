#!/bin/bash
set -euxo pipefail

echo "Installing Profiles Controller"
kustomize build apps/profiles/upstream/overlays/kubeflow | kubectl apply -f -
sleep 30
kubectl -n kubeflow wait --for=condition=Ready pods -l kustomize.component=profiles --timeout 180s

echo "Installing Multitenancy Kubeflow Roles"
kustomize build common/kubeflow-roles/base | kubectl apply -f -
