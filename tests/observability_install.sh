#!/bin/bash
set -euxo pipefail
echo "Installing observability stack..."
# The base component now includes everything needed, including CRDs and Operators.
# We apply the overlay which includes the base.
cd common/observability
kustomize build overlays/kubeflow | kubectl apply --server-side --force-conflicts -f -

echo "Waiting for operators to be ready..."
kubectl wait --for=condition=Ready pod -l 'app.kubernetes.io/name=prometheus-operator' --timeout=180s -n kubeflow-monitoring-system
kubectl wait --for=condition=Ready pod -l 'control-plane=controller-manager' --timeout=180s -n kubeflow-monitoring-system
echo "Observability stack installed successfully."
