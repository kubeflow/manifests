#!/bin/bash
set -e

echo "Installing Central Dashboard..."
kustomize build apps/centraldashboard/upstream/overlays/kserve | kubectl apply -f -

echo "Waiting for pods to become ready..."
kubectl wait --for=condition=Ready pods --all -n kubeflow --timeout=180s

echo "Central Dashboard installation completed." 