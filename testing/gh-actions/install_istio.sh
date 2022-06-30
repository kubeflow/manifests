#!/bin/bash
set -e
echo "Installing Istio ..."
cd common/istio-1-14
kustomize build istio-crds/base | kubectl apply -f -
kustomize build istio-namespace/base | kubectl apply -f -
kustomize build istio-install/base | kubectl apply -f -