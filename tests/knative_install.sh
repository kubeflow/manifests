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

set +e
for ((i=1; i<=3; i++)); do
    if kustomize build common/knative/knative-eventing/overlays/security | kubectl apply -f -; then
        break
    fi
    if [[ "${i}" -eq 3 ]]; then
        echo "Failed to apply knative-eventing security overlay after 3 attempts" >&2
        exit 1
    fi
    sleep 10
done
set -e

kubectl rollout status deployment/eventing-controller -n knative-eventing --timeout=120s
kubectl rollout status deployment/eventing-webhook -n knative-eventing --timeout=120s
kubectl rollout status statefulset/request-reply -n knative-eventing --timeout=120s
kubectl get namespace knative-eventing -o jsonpath='{.metadata.labels.pod-security\.kubernetes\.io/enforce}' | grep -x restricted
kubectl get networkpolicy default-allow-same-namespace-knative-eventing -n knative-eventing
kubectl get networkpolicy webhook-apiserver -n knative-eventing
kubectl get deployment -n knative-serving
kubectl get deployment,statefulset -n knative-eventing
