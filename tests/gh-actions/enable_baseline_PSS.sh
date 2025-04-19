#!/bin/bash
set -euxo pipefail

NAMESPACES=("istio-system" "auth" "cert-manager" "oauth2-proxy" "kubeflow")

for NAMESPACE in "${NAMESPACES[@]}"; do
    if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
        # Check if the patch file exists
        if [ -f "./experimental/security/PSS/static/baseline/patches/${NAMESPACE}-labels.yaml" ]; then
            echo "Patching the PSS-baseline labels for namespace $NAMESPACE..."
            kubectl patch namespace $NAMESPACE --patch-file ./experimental/security/PSS/static/baseline/patches/${NAMESPACE}-labels.yaml
        else
            echo "Patch file for namespace $NAMESPACE not found, skipping..."
        fi
    fi
done

VIOLATIONS=$(kubectl get pods --all-namespaces -o json | jq -r '.items[] | select(.metadata.namespace as $ns | ["istio-system", "auth", "cert-manager", "oauth2-proxy", "kubeflow"] | index($ns) != null) | select(.status.message != null and ((.status.message | type) == "string") and (.status.message | contains("violate PodSecurity")))')

if [ -n "$VIOLATIONS" ]; then
    echo "$VIOLATIONS" | jq -r '.metadata.namespace + "/" + .metadata.name + ": " + .status.message'
    exit 1
fi
