#!/bin/bash
set -euo pipefail
echo "Installing KNative ..."

kustomize build common/knative/knative-serving/base | kubectl apply -f -

kustomize build common/istio-1-16/cluster-local-gateway/base | kubectl apply -f -
kustomize build common/istio-1-16/kubeflow-istio-resources/base | kubectl apply -f -