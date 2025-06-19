#!/bin/bash
set -euxo pipefail
echo "Installing Kserve ..."
cd applications/kserve
set +e
for ((i=1; i<=3; i++)); do
    if kustomize build kserve | kubectl apply --server-side --force-conflicts -f -; then
        break
    fi
    kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=60s --field-selector=status.phase!=Succeeded
    kubectl wait --for=condition=Ready certificate/serving-cert -n kubeflow --timeout=60s
    kubectl get secret kserve-webhook-server-cert -n kubeflow -o name
done
set -e

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
