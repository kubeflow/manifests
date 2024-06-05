#!/bin/bash
set -euo pipefail
echo "Installing KNative with istio-cni ..."
set +e
kustomize build common/knative/knative-serving/base | kubectl apply -f -
set -e
kustomize build common/knative/knative-serving/base | kubectl apply -f -

kustomize build common/istio-cni-1-21/cluster-local-gateway/base | kubectl apply -f -
kustomize build common/istio-cni-1-21/kubeflow-istio-resources/base | kubectl apply -f -

kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout 600s
kubectl patch cm config-domain --patch '{"data":{"example.com":""}}' -n knative-serving
