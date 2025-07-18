#!/bin/bash
set -euo pipefail

set +e
for ((i=1; i<=3; i++)); do
    if kustomize build common/knative/knative-serving/overlays/gateways | kubectl apply -f -; then
        break
    fi
    kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=60s --field-selector=status.phase!=Succeeded
done
set -e

kustomize build common/istio/cluster-local-gateway/overlays/m2m-auth | kubectl apply -f -

kustomize build common/istio/kubeflow-istio-resources/base | kubectl apply -f -

kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=60s --field-selector=status.phase!=Succeeded
kubectl wait --for=condition=Available deployment/activator -n knative-serving --timeout=10s
kubectl wait --for=condition=Available deployment/autoscaler -n knative-serving --timeout=10s
kubectl wait --for=condition=Available deployment/controller -n knative-serving --timeout=10s
kubectl wait --for=condition=Available deployment/webhook -n knative-serving --timeout=10s
kubectl get deployment -n knative-serving
