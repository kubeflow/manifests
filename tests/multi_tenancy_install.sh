#!/bin/bash
set -euo pipefail

echo "Installing Profiles Controller with PSS (Pod Security Standards)"
kustomize build applications/profiles/pss | kubectl apply -f -
kubectl -n kubeflow wait --for=condition=Ready pods -l kustomize.component=profiles --timeout 180s

echo "Installing Multitenancy Kubeflow Roles"
kustomize build common/kubeflow-roles/base | kubectl apply -f -

echo "Installing Multitenancy Network policies"
# Create namespaces if they don't exist (required for network policies)
for ns in auth cert-manager istio-system knative-serving kubeflow-system oauth2-proxy; do
  kubectl create namespace "$ns" --dry-run=client -o yaml | kubectl apply -f -
done
kustomize build common/networkpolicies/base | kubectl apply -f -
