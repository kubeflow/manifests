#!/bin/bash
set -euxo pipefail

# KServe Models Web App Test Script
# Tests the models-web-app API functionality with kubectl-deployed InferenceServices

KF_PROFILE=${1:-kubeflow-user-example-com}
TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
BASE_URL="localhost:8080/kserve-endpoints"

echo "=========================================="
echo "KServe Models Web App Test"
echo "Profile: ${KF_PROFILE}"
echo "=========================================="

# Pre-Test Setup: Create RBAC for default-editor to access InferenceServices
echo "Step 1: Setting up RBAC permissions for InferenceService access..."
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: inferenceservice-editor
  namespace: ${KF_PROFILE}
rules:
- apiGroups:
  - serving.kserve.io
  resources:
  - inferenceservices
  verbs:
  - get
  - list
  - watch
  - create
  - delete
  - deletecollection
  - patch
  - update
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: default-editor-inferenceservice-access
  namespace: ${KF_PROFILE}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: inferenceservice-editor
subjects:
- kind: ServiceAccount
  name: default-editor
  namespace: ${KF_PROFILE}
EOF

echo "✓ RBAC permissions configured"

# Pre-Test Setup: Deploy InferenceService via kubectl
echo ""
echo "Step 2: Deploying test InferenceService via kubectl..."
cat <<EOF | kubectl apply -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "sklearn-iris-private"
  namespace: ${KF_PROFILE}
spec:
  predictor:
    sklearn:
      storageUri: "gs://kfserving-examples/models/sklearn/1.0/model"
      resources:
        requests:
          cpu: "50m"
          memory: "128Mi"
        limits:
          cpu: "100m"
          memory: "256Mi"
EOF

echo "Waiting for InferenceService to be ready..."
kubectl wait --for=condition=Ready inferenceservice/sklearn-iris-private -n ${KF_PROFILE} --timeout=300s

echo "✓ InferenceService deployed successfully!"
kubectl get inferenceservice sklearn-iris-private -n ${KF_PROFILE}

# Get XSRF token for API calls
echo ""
echo "Step 3: Getting XSRF token..."
curl -s "http://${BASE_URL}/" \
  -H "Authorization: Bearer ${TOKEN}" \
  -v -c /tmp/kserve_xcrf.txt 2>&1 | grep -i "set-cookie" || true

if [ -f /tmp/kserve_xcrf.txt ]; then
  XSRFTOKEN=$(grep XSRF-TOKEN /tmp/kserve_xcrf.txt | awk '{print $NF}' || echo "")
  if [ -z "$XSRFTOKEN" ]; then
    echo "WARNING: Could not extract XSRF token, will try without it"
    XSRFTOKEN="dummy-token"
  else
    echo "✓ XSRF token retrieved: ${XSRFTOKEN:0:20}..."
  fi
else
  echo "WARNING: Cookie file not created, proceeding without XSRF token"
  XSRFTOKEN="dummy-token"
fi

# Test: Verify InferenceService appears in the models-web-app API
echo ""
echo "Step 4: Verifying InferenceService appears in models-web-app API..."

RESPONSE=$(curl -s --fail-with-body \
  "${BASE_URL}/api/namespaces/${KF_PROFILE}/inferenceservices" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${XSRFTOKEN}" \
  -H "Cookie: XSRF-TOKEN=${XSRFTOKEN}" 2>&1 || echo '{"error": "request failed"}')

echo "API Response:"
echo "$RESPONSE" | head -c 500
echo ""

# Check if sklearn-iris-private is in the response
if echo "$RESPONSE" | grep -q "sklearn-iris-private"; then
  echo "✓ SUCCESS: InferenceService 'sklearn-iris-private' found in models-web-app API response"
else
  echo "ERROR: InferenceService 'sklearn-iris-private' NOT found in API response"
  echo "Full response: $RESPONSE"
  exit 1
fi

# Verify Cluster State Consistency
echo ""
echo "Step 5: Verifying Cluster State Consistency..."

# Check that kubectl shows the InferenceService
if kubectl get inferenceservice sklearn-iris-private -n ${KF_PROFILE} > /dev/null 2>&1; then
  echo "✓ InferenceService exists in cluster"
else
  echo "ERROR: InferenceService not found in cluster"
  exit 1
fi

# Check that InferenceService is Ready
READY_STATUS=$(kubectl get inferenceservice sklearn-iris-private -n ${KF_PROFILE} -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>&1 || echo "Unknown")

if [[ "$READY_STATUS" == "True" ]]; then
  echo "✓ InferenceService is Ready"
else
  echo "WARNING: InferenceService Ready status is: $READY_STATUS"
fi

# Cleanup
echo ""
echo "Step 6: Cleanup - Deleting test resources..."

# Delete InferenceService
kubectl delete inferenceservice sklearn-iris-private -n ${KF_PROFILE}

echo "Waiting for InferenceService to be deleted..."
sleep 5

if kubectl get inferenceservice sklearn-iris-private -n ${KF_PROFILE} > /dev/null 2>&1; then
  echo "WARNING: InferenceService still exists after deletion"
else
  echo "✓ InferenceService successfully deleted"
fi

# Delete RBAC resources
kubectl delete role inferenceservice-editor -n ${KF_PROFILE} 2>/dev/null || true
kubectl delete rolebinding default-editor-inferenceservice-access -n ${KF_PROFILE} 2>/dev/null || true
echo "✓ RBAC permissions cleaned up"

echo ""
echo "=========================================="
echo "✓ All tests passed successfully!"
echo "=========================================="


echo ""
echo "=========================================="
echo "✓ KServe Models Web App Test PASSED"
echo "=========================================="
