#!/bin/bash

set -eux

# Install kubeflow using raw manifests
cd manifests
while ! kustomize build example | kubectl apply -f -; do echo "Retrying to apply resources"; sleep 10; done

# Wait for pods to be ready
kubectl -n kubeflow wait --for=condition=Ready pods --all --timeout=1200s