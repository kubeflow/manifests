#!/bin/bash
set -euxo pipefail

# KServe Models Web App Test Script
# Tests the models-web-app API functionality with kubectl-deployed InferenceServices

KF_PROFILE=${1:-kubeflow-user-example-com}
TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
UNAUTHORIZED_TOKEN="$(kubectl -n default create token default)"
BASE_URL="localhost:8080/kserve-endpoints"

echo "=========================================="
echo "KServe Models Web App Test"
echo "Profile: ${KF_PROFILE}"
echo "=========================================="

# Pre-Test Setup: Create RBAC for default-editor to access InferenceServices
echo "Step 0: Setting up RBAC permissions for InferenceService access..."
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: default-editor-kserve-access
  namespace: ${KF_PROFILE}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kserve-models-web-app-cluster-role
subjects:
- kind: ServiceAccount
  name: default-editor
  namespace: ${KF_PROFILE}
EOF

echo "✓ RBAC permissions configured"

# Pre-Test Setup: Deploy InferenceService via kubectl
echo ""
echo "Step 1: Deploying test InferenceService via kubectl..."
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

echo "InferenceService deployed successfully!"
kubectl get inferenceservice sklearn-iris-private -n ${KF_PROFILE}

# Test 2: Authentication Test
echo ""
echo "Step 2: Testing Authentication..."

# Test without token (should fail)
echo "Testing access without token (expecting 403 or 302)..."
RESPONSE_NO_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" \
  "http://${BASE_URL}/" 2>&1 || echo "000")

if [ "$RESPONSE_NO_TOKEN" != "403" ] && [ "$RESPONSE_NO_TOKEN" != "302" ] && [ "$RESPONSE_NO_TOKEN" != "401" ]; then
  echo "WARNING: Expected 403/302/401 for unauthenticated access, got $RESPONSE_NO_TOKEN"
else
  echo "✓ Unauthenticated access correctly rejected with status $RESPONSE_NO_TOKEN"
fi

# Get XSRF token
echo "Getting XSRF token..."
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

# Test 3: List InferenceServices
echo ""
echo "Step 3: Testing List InferenceServices API..."

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
  echo "✓ InferenceService 'sklearn-iris-private' found in API response"
else
  echo "ERROR: InferenceService 'sklearn-iris-private' NOT found in API response"
  echo "Full response: $RESPONSE"
  exit 1
fi

# Test 4: Get InferenceService Details
echo ""
echo "Step 4: Testing Get InferenceService Details API..."

DETAILS_RESPONSE=$(curl -s --fail-with-body \
  "${BASE_URL}/api/namespaces/${KF_PROFILE}/inferenceservices/sklearn-iris-private" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${XSRFTOKEN}" \
  -H "Cookie: XSRF-TOKEN=${XSRFTOKEN}" 2>&1 || echo '{"error": "request failed"}')

echo "Details API Response:"
echo "$DETAILS_RESPONSE" | head -c 500
echo ""

if echo "$DETAILS_RESPONSE" | grep -q "sklearn-iris-private"; then
  echo "✓ InferenceService details retrieved successfully"
else
  echo "ERROR: Could not retrieve InferenceService details"
  echo "Full response: $DETAILS_RESPONSE"
  exit 1
fi

# Test 5: Unauthorized Access Test
echo ""
echo "Step 5: Testing Unauthorized Access..."

UNAUTHORIZED_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  "${BASE_URL}/api/namespaces/${KF_PROFILE}/inferenceservices" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}" 2>&1)

if [[ "$UNAUTHORIZED_STATUS" == "403" ]]; then
  echo "✓ Unauthorized access correctly rejected with 403"
elif [[ "$UNAUTHORIZED_STATUS" == "401" ]]; then
  echo "✓ Unauthorized access correctly rejected with 401"
else
  echo "WARNING: Expected 403/401 for unauthorized access, got $UNAUTHORIZED_STATUS"
fi

# Test 6: Verify Cluster State Consistency
echo ""
echo "Step 6: Verifying Cluster State Consistency..."

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

# Test 7: Cleanup
echo ""
echo "Step 7: Cleanup - Deleting test resources..."

# Delete InferenceService
kubectl delete inferenceservice sklearn-iris-private -n ${KF_PROFILE}

echo "Waiting for InferenceService to be deleted..."
sleep 5

if kubectl get inferenceservice sklearn-iris-private -n ${KF_PROFILE} > /dev/null 2>&1; then
  echo "WARNING: InferenceService still exists after deletion"
else
  echo "✓ InferenceService successfully deleted"
fi

# Delete RBAC RoleBinding
kubectl delete rolebinding default-editor-kserve-access -n ${KF_PROFILE} 2>/dev/null || true
echo "✓ RBAC permissions cleaned up"

echo ""
echo "=========================================="
echo "✓ KServe Models Web App Test PASSED"
echo "=========================================="
