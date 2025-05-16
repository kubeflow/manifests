#!/bin/bash
set -euxo pipefail

kustomize build ./common/dex/overlays/oauth2-proxy | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all --timeout=180s -n auth 