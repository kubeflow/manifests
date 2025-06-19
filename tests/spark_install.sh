#!/bin/bash
set -euxo

REPOSITORY_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "${GITHUB_WORKSPACE:-$(pwd)}")
cd "${REPOSITORY_ROOT}"

# Install Spark operator
kustomize build applications/spark/spark-operator/overlays/standalone | kubectl -n kubeflow apply --server-side -f -

# Wait for the operator controller to be ready.
kubectl -n kubeflow wait --for=condition=available --timeout=60s deploy/spark-operator-controller
kubectl -n kubeflow get pod -l app.kubernetes.io/name=spark-operator

# Wait for the operator webhook to be ready.
kubectl -n kubeflow wait --for=condition=available --timeout=30s deploy/spark-operator-webhook
kubectl -n kubeflow get pod -l app.kubernetes.io/name=spark-operator
