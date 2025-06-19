#!/bin/bash
set -euxo pipefail

cd applications/volumes-web-app/upstream
kustomize build overlays/istio | kubectl apply -f -
cd ../../../

kubectl wait --for=condition=Available deployment/volumes-web-app-deployment -n kubeflow --timeout=180s 
