#!/bin/bash
set -euo pipefail


PSS_LEVEL="${1:-baseline}"

if [[ "$PSS_LEVEL" != "baseline" && "$PSS_LEVEL" != "restricted" ]]; then
    echo "ERROR: Invalid PSS level '$PSS_LEVEL'. Must be 'baseline' or 'restricted'."
    echo "Usage: $0 [baseline|restricted]"
    exit 1
fi

if [[ "$PSS_LEVEL" == "baseline" ]]; then
    NAMESPACES=("istio-system" "auth" "cert-manager" "oauth2-proxy" "kubeflow" "knative-serving" "kubeflow-system" "kubeflow-user-example-com")
else
    NAMESPACES=("istio-system" "auth" "cert-manager" "oauth2-proxy" "kubeflow" "knative-serving" "kubeflow-system")
fi

echo "Enabling Pod Security Standards with level: $PSS_LEVEL"
echo "Namespaces to process: ${NAMESPACES[*]}"

for NAMESPACE in "${NAMESPACES[@]}"; do
    if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
        echo "Processing namespace: $NAMESPACE"
        PATCH_OUTPUT=$(kubectl label namespace "$NAMESPACE" "pod-security.kubernetes.io/enforce=$PSS_LEVEL" --overwrite 2>&1)
        if echo "$PATCH_OUTPUT" | grep -q "violate the new PodSecurity"; then
            echo "ERROR: PSS violation detected for namespace $NAMESPACE"
            echo "$PATCH_OUTPUT" | grep -A 5 "violate the new PodSecurity"
            exit 1
        else
            echo "✅ Namespace '$NAMESPACE' labeled successfully with $PSS_LEVEL PSS."
        fi
    else
        echo "Namespace '$NAMESPACE' not found, skipping."
    fi
done

echo "Pod Security Standards ($PSS_LEVEL) enforcement completed successfully!"