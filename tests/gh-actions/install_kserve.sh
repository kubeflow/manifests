#!/bin/bash
set -euo pipefail
echo "Installing Kserve ..."
cd contrib/kserve
kubectl create ns kubeflow
set +e
kustomize build kserve | kubectl apply -f -
kubectl wait --for condition=established --timeout=30s crd/clusterservingruntimes.serving.kserve.io
set -e
kustomize build kserve | kubectl apply -f -
kustomize build models-web-app/overlays/kubeflow | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout 180s