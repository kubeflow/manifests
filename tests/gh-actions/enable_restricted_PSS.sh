#!/bin/bash

NAMESPACES=("istio-system" "auth" "cert-manager" "oauth2-proxy" "kubeflow")

for NAMESPACE in "${NAMESPACES[@]}"; do
    if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
        if [ -f "./experimental/security/PSS/static/restricted/patches/${NAMESPACE}-labels.yaml" ]; then
            echo "Patching the PSS-restricted labels for namespace $NAMESPACE..."
            kubectl patch namespace $NAMESPACE --patch-file ./experimental/security/PSS/static/restricted/patches/${NAMESPACE}-labels.yaml
        else
            echo "Patch file for namespace $NAMESPACE not found, skipping..."
        fi
    fi
done