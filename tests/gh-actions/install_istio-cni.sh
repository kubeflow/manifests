#!/bin/bash
set -e
echo "Installing Istio-cni ..."
cd common/istio-cni-1-21
kustomize build istio-crds/base | kubectl apply -f -
kustomize build istio-namespace/base | kubectl apply -f -
kustomize build istio-install/base | kubectl apply -f -