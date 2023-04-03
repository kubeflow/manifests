#!/bin/bash
set -euo pipefail
echo "Installing Kserve ..."
cd contrib/kserve
set +e
kustomize build kserve | kubectl apply -f -
set -e
echo "Waiting for crd/clusterservingruntimes.serving.kserve.io to be available ..."
kubectl wait --for condition=established --timeout=30s crd/clusterservingruntimes.serving.kserve.io
kustomize build kserve | kubectl apply -f -
kustomize build models-web-app/overlays/kubeflow | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout 600s
