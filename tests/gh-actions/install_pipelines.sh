#!/bin/bash
set -euo pipefail
echo "Installing Pipelines ..."
cd apps/pipeline/upstream
kubectl create ns kubeflow
kubectl apply -f third-party/metacontroller/base/crd.yaml
echo "Waiting for crd/compositecontrollers.metacontroller.k8s.io to be available ..."
kubectl wait --for condition=established --timeout=30s crd/compositecontrollers.metacontroller.k8s.io
kustomize build env/cert-manager/platform-agnostic-multi-user | kubectl apply -f -
sleep 60
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout 600s
