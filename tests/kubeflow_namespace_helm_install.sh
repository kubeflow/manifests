#!/usr/bin/env bash
# Helm counterpart of the "Create Kubeflow Namespace" step: the
# kubeflow-namespaces chart owns every platform namespace and the network
# policies, proven equivalent to common/kubeflow-namespace/base by
# tests/run_helm_kustomize_comparison.py kubeflow-namespaces.
set -euxo pipefail
helm install kubeflow-namespaces common/kubeflow-namespace/helm \
  --namespace default \
  --values common/kubeflow-namespace/helm/ci/values-default.yaml \
  --wait --timeout 5m
