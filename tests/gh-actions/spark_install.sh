#!/bin/bash
set -euxo

# Install Spark operator
kustomize build spark-operator/overlays/standalone | kubectl -n kubeflow apply --server-side -f -

# Wait for the operator controller to be ready.
kubectl -n kubeflow wait --for=condition=available --timeout=60s deploy/spark-operator-controller
kubectl -n kubeflow get pod -l app.kubernetes.io/name=spark-operator

# Wait for the operator webhook to be ready.
kubectl -n kubeflow wait --for=condition=available --timeout=30s deploy/spark-operator-webhook
kubectl -n kubeflow get pod -l app.kubernetes.io/name=spark-operator
