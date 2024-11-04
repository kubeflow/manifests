#!/bin/bash
set -euo pipefail
echo "Installing Kserve ..."
cd contrib/kserve
set +e
kustomize build kserve | kubectl apply --server-side --force-conflicts -f -
sleep 30
kustomize build kserve | kubectl apply --server-side --force-conflicts -f -
set -e
echo "Waiting for crd/clusterservingruntimes.serving.kserve.io to be available ..."
kubectl wait --for condition=established --timeout=30s crd/clusterservingruntimes.serving.kserve.io
kustomize build kserve | kubectl apply --server-side --force-conflicts -f -
kustomize build models-web-app/overlays/kubeflow | kubectl apply --server-side --force-conflicts -f -
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=600s \
  --field-selector=status.phase!=Succeeded
