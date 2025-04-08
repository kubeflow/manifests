#!/bin/bash
set -euo pipefail

echo "Installing KNative ..."

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

kustomize build common/istio-1-24/cluster-local-gateway/base | kubectl apply -f -
kustomize build common/istio-1-24/kubeflow-istio-resources/base | kubectl apply -f -

kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=300s --field-selector=status.phase!=Succeeded
kubectl wait --for=condition=Available deployment/activator -n knative-serving --timeout=10s
kubectl wait --for=condition=Available deployment/autoscaler -n knative-serving --timeout=10s
kubectl wait --for=condition=Available deployment/controller -n knative-serving --timeout=10s
kubectl wait --for=condition=Available deployment/webhook -n knative-serving --timeout=10s
kubectl patch cm config-domain --patch '{"data":{"example.com":""}}' -n knative-serving # TODO this will be moved into the manifests in https://github.com/kubeflow/manifests/pull/3084
kubectl get deployment -n knative-serving
