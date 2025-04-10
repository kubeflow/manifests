#!/bin/bash
set -euxo pipefail


cd apps/volumes-web-app/upstream
kustomize build overlays/istio | kubectl apply -f -
cd ../../../

sleep 5 