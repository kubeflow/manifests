#!/bin/bash
set -euxo pipefail

curl -sL https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml > metrics-server.yaml
sed -i 's/- args:/- args:\n        - --kubelet-insecure-tls/' metrics-server.yaml
kubectl apply -f metrics-server.yaml
kubectl wait --for=condition=Ready pods -l k8s-app=metrics-server -n kube-system --timeout=180s