#!/bin/bash
set -euxo pipefail

# Test Model Registry API and UI integration
# This script can be used for local testing without GitHub Actions
# Prerequisites: Model Registry must be installed (run install.sh first)
# Usage: ./tests/model_registry_test/test.sh

echo "=== Model Registry Integration Tests ==="

# ---- Test 1: Direct API access via port-forward ----
echo "Test 1: Direct Model Registry API access..."
nohup kubectl port-forward svc/model-registry-service -n kubeflow 8081:8080 &
PORT_FORWARD_PID=$!

MAX_RETRIES=30
RETRY_COUNT=0
while ! curl -s localhost:8081 > /dev/null 2>&1; do
    echo "Waiting for port-forwarding to be ready... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 1
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        echo "ERROR: Port-forwarding to model-registry-service failed"
        kill $PORT_FORWARD_PID 2>/dev/null || true
        exit 1
    fi
done
echo "Port-forwarding ready on 8081!"

# Test registered models endpoint
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  "http://localhost:8081/api/model_registry/v1alpha3/registered_models?pageSize=100&orderBy=ID&sortOrder=DESC")

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "PASS: Model Registry API responding (HTTP $HTTP_CODE)"
    curl -s "http://localhost:8081/api/model_registry/v1alpha3/registered_models" | head -c 500
    echo ""
else
    echo "FAIL: Model Registry API returned unexpected status: $HTTP_CODE"
    kill $PORT_FORWARD_PID 2>/dev/null || true
    exit 1
fi

kill $PORT_FORWARD_PID 2>/dev/null || true

# ---- Test 2: API access through Istio gateway ----
echo ""
echo "Test 2: Model Registry API via Istio gateway..."
INGRESS_GATEWAY_SERVICE=$(kubectl get svc --namespace istio-system \
  --selector="app=istio-ingressgateway" \
  --output jsonpath='{.items[0].metadata.name}')

nohup kubectl port-forward --namespace istio-system "svc/${INGRESS_GATEWAY_SERVICE}" 8080:80 &
GATEWAY_PID=$!

RETRY_COUNT=0
while ! curl -s localhost:8080 > /dev/null 2>&1; do
    echo "Waiting for gateway port-forwarding... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 1
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        echo "ERROR: Gateway port-forwarding failed"
        kill $GATEWAY_PID 2>/dev/null || true
        exit 1
    fi
done
echo "Gateway port-forwarding ready on 8080!"

# Test authenticated access (authorized SA)
export KF_PROFILE=kubeflow-user-example-com
export KF_TOKEN="$(kubectl -n "$KF_PROFILE" create token default-editor)"

STATUS_CODE=$(curl -s -o /dev/stderr -w "%{http_code}" \
    "localhost:8080/model-registry/api/v1/model_registry?namespace=${KF_PROFILE}" \
    -H "Authorization: Bearer ${KF_TOKEN}" 2>/dev/null)

if [ "$STATUS_CODE" -eq 200 ]; then
    echo "PASS: Authorized access to Model Registry via gateway (HTTP $STATUS_CODE)"
else
    echo "FAIL: Expected HTTP 200 for authorized access, got: $STATUS_CODE"
    kill $GATEWAY_PID 2>/dev/null || true
    exit 1
fi

# Test unauthorized access (default SA - should be 403)
export KF_TOKEN_UNAUTH="$(kubectl -n default create token default)"

STATUS_CODE=$(curl -s -o /dev/stderr -w "%{http_code}" \
    "localhost:8080/model-registry/api/v1/model_registry?namespace=${KF_PROFILE}" \
    -H "Authorization: Bearer ${KF_TOKEN_UNAUTH}" 2>/dev/null)

if [ "$STATUS_CODE" -eq 403 ]; then
    echo "PASS: Unauthorized access correctly denied (HTTP $STATUS_CODE)"
else
    echo "FAIL: Expected HTTP 403 for unauthorized access, got: $STATUS_CODE"
    kill $GATEWAY_PID 2>/dev/null || true
    exit 1
fi

kill $GATEWAY_PID 2>/dev/null || true

echo ""
echo "=== All Model Registry tests passed! ==="
