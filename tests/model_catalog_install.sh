#!/bin/bash
set -euxo pipefail

(
    cd applications/model-registry/upstream/options/catalog/base
    kustomize build . | kubectl apply -n kubeflow -f -
)

kubectl wait --for=condition=Available deployment/model-catalog-server -n kubeflow --timeout=120s