#!/bin/bash
set -euo pipefail

echo "Installing KNative with Istio-CNI ..."

# Retry mechanism for applying Knative manifests
set +e
for _ in {1..5}; do
    if kustomize build common/knative/knative-serving/base | kubectl apply -f -; then
        break
    fi
    echo "Retrying in 30 seconds..."
    sleep 30
done
set -e

kustomize build common/istio-cni-1-24/cluster-local-gateway/base | kubectl apply -f -
kustomize build common/istio-cni-1-24/kubeflow-istio-resources/base | kubectl apply -f -

kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=300s \
  --field-selector=status.phase!=Succeeded
kubectl patch cm config-domain --patch '{"data":{"example.com":""}}' -n knative-serving
