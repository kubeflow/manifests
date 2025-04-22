#!/bin/bash
set -euxo pipefail

NAMESPACES=("istio-system" "auth" "cert-manager" "oauth2-proxy" "kubeflow")
for NAMESPACE in "${NAMESPACES[@]}"; do
    if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
        if [ -f "./experimental/security/PSS/static/restricted/patches/${NAMESPACE}-labels.yaml" ]; then
            PATCH_OUTPUT=$(kubectl patch namespace $NAMESPACE --patch-file ./experimental/security/PSS/static/baseline/patches/${NAMESPACE}-labels.yaml 2>&1)
            if echo "$PATCH_OUTPUT" | grep -q "violate the new PodSecurity"; then
                echo "\nWARNING PSS VIOLATED\n"
            fi
        fi
    fi
done
