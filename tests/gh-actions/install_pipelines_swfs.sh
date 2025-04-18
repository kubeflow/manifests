#!/bin/bash
set -euo pipefail
echo "Installing Pipelines ..."
kubectl apply -f apps/pipeline/upstream/third-party/metacontroller/base/crd.yaml
echo "Waiting for crd/compositecontrollers.metacontroller.k8s.io to be available ..."
kubectl wait --for condition=established --timeout=30s crd/compositecontrollers.metacontroller.k8s.io
kustomize build experimental/seaweedfs/istio | kubectl apply -f -
sleep 60
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=600s \
  --field-selector=status.phase!=Succeeded
