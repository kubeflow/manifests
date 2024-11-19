#!/bin/bash
set -euo pipefail

echo "Installing KNative ..."

# Retry mechanism for applying Knative manifests
set +e
for i in {1..5}; do
    kustomize build common/knative/knative-serving/base | kubectl apply -f -
    if [[ $? -eq 0 ]]; then
        break
    fi
    echo "Retrying in 30 seconds..."
    sleep 30
done
set -e

kustomize build common/istio-cni-1-23/cluster-local-gateway/base | kubectl apply -f -
kustomize build common/istio-cni-1-23/kubeflow-istio-resources/base | kubectl apply -f -

kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=300s \
  --field-selector=status.phase!=Succeeded
kubectl patch cm config-domain --patch '{"data":{"example.com":""}}' -n knative-serving
