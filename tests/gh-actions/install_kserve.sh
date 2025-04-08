#!/bin/bash
set -euo pipefail
echo "Installing Kserve ..."
cd apps/kserve
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
kubectl wait --for=condition=Available deployment/kserve-controller-manager -n kubeflow --timeout=10s
kubectl wait --for=condition=Available deployment/kserve-models-web-app -n kubeflow --timeout=10s
kubectl get deployment -n kubeflow -l app.kubernetes.io/name=kserve
kubectl get crd | grep -E 'inferenceservice|servingruntimes'

# Return to the original directory
cd ../../
