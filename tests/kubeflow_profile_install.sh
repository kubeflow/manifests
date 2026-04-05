#!/bin/bash
set -euxo pipefail
kustomize build common/user-namespace/base | kubectl apply -f -
sleep 30 # Let the profile controler reconcile the namespace
PROFILE_CONTROLLER_POD=$(kubectl get pods -n kubeflow -o json | jq -r '.items[] | select(.metadata.name | startswith("profiles-deployment")) | .metadata.name')
kubectl logs -n kubeflow "$PROFILE_CONTROLLER_POD"
KF_PROFILE=kubeflow-user-example-com
kubectl -n $KF_PROFILE get pods,configmaps,secrets

echo "Verifying PSS Restricted enforcement on namespace $KF_PROFILE..."
# Profiles controller should automatically add the label via the 'pss' overlay
MAX_RETRIES=10
for i in $(seq 1 $MAX_RETRIES); do
    PSS_LABEL=$(kubectl get namespace "$KF_PROFILE" -o jsonpath='{.metadata.labels.pod-security\.kubernetes\.io/enforce}')
    if [[ "$PSS_LABEL" == "restricted" ]]; then
        echo "✅ Namespace $KF_PROFILE is correctly labeled as restricted."
        exit 0
    fi
    echo "Wait for Profiles controller to label the namespace (attempt $i/$MAX_RETRIES)..."
    sleep 5
done

echo "❌ ERROR: Namespace $KF_PROFILE is NOT labeled as restricted."
kubectl get namespace "$KF_PROFILE" -o yaml
exit 1
