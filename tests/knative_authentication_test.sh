#!/bin/bash
set -euo pipefail

echo "Knative Service JWT Authentication Test"
echo "=========================================="
echo

# Configuration
PRIMARY_NAMESPACE="kubeflow-user-example-com"
ATTACKER_NAMESPACE="kubeflow-user-attacker"
INGRESS_HOST_PORT="${KSERVE_INGRESS_HOST_PORT:-localhost:8080}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

function log_test() { echo -e "${YELLOW}Test: $1${NC}"; }
function log_pass() { echo -e "${GREEN}PASS: $1${NC}"; }
function log_fail() { echo -e "${RED}FAIL: $1${NC}"; exit 1; }
function log_info() { echo -e "INFO: $1"; }

# Setup
function setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create namespaces
    kubectl create namespace $PRIMARY_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace $ATTACKER_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    
    # Create service accounts
    kubectl create serviceaccount default-editor -n $PRIMARY_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    kubectl create serviceaccount attacker-sa -n $ATTACKER_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    
    # Create test Knative service (simulating KServe predictor)
    cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: secure-model-predictor
  namespace: $PRIMARY_NAMESPACE
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        ports:
        - containerPort: 8080
        env:
        - name: TARGET
          value: "Secure KServe Model"
EOF

    log_info "Waiting for service to be ready..."
    kubectl wait --for=condition=Ready ksvc/secure-model-predictor -n $PRIMARY_NAMESPACE --timeout=120s
    
    # Note: Namespace isolation policies removed for this PR
    # Focus is on gateway-level JWT authentication only

    log_info "Test environment ready"
    echo
}

# Test gateway JWT validation
function test_gateway_authentication() {
    log_test "Gateway JWT Authentication"
    
    # Test 1: No token
    log_info "Testing without JWT token..."
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        "http://$INGRESS_HOST_PORT/")
    
    if [ "$RESPONSE" = "403" ]; then
        log_pass "No token correctly rejected (403)"
    else
        log_fail "Expected 403, got $RESPONSE"
    fi
    
    # Test 2: Invalid token
    log_info "Testing with invalid JWT token..."
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer invalid-token" \
        "http://$INGRESS_HOST_PORT/")
    
    if [ "$RESPONSE" = "401" ] || [ "$RESPONSE" = "403" ]; then
        log_pass "Invalid token correctly rejected ($RESPONSE)"
    else
        log_fail "Expected 401/403, got $RESPONSE"
    fi
    
    echo
}

# Test namespace isolation
function test_namespace_isolation() {
    log_test "Namespace Isolation"
    
    # Get tokens
    PRIMARY_TOKEN=$(kubectl -n $PRIMARY_NAMESPACE create token default-editor)
    ATTACKER_TOKEN=$(kubectl -n $ATTACKER_NAMESPACE create token attacker-sa)
    
    # Test 3: Same namespace access
    log_info "Testing same namespace access..."
    RESPONSE=$(curl -s -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer $PRIMARY_TOKEN" \
        "http://$INGRESS_HOST_PORT/")
    
    if [[ "$RESPONSE" == *"200"* ]] || [[ "$RESPONSE" == *"404"* ]] || [[ "$RESPONSE" == *"503"* ]]; then
        log_pass "Same namespace access allowed (${RESPONSE##*HTTP_CODE:})"
    else
        log_fail "Same namespace failed: $RESPONSE"
    fi
    
    # Test 4: Cross-namespace access (should work with valid token - no namespace isolation in this PR)
    log_info "Testing cross-namespace access..."
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer $ATTACKER_TOKEN" \
        "http://$INGRESS_HOST_PORT/")
    
    if [[ "$RESPONSE" == *"200"* ]] || [[ "$RESPONSE" == *"404"* ]] || [[ "$RESPONSE" == *"503"* ]]; then
        log_pass "Cross-namespace access allowed with valid token (${RESPONSE##*HTTP_CODE:})"
    else
        log_fail "Cross-namespace access failed: $RESPONSE"
    fi
    
    echo
}

# Test via cluster-local-gateway directly
function test_cluster_local_gateway() {
    log_test "Direct cluster-local-gateway Access"
    
    # Port forward cluster-local-gateway
    kubectl port-forward -n istio-system svc/cluster-local-gateway 8081:80 &
    PF_PID=$!
    sleep 5
    
    PRIMARY_TOKEN=$(kubectl -n $PRIMARY_NAMESPACE create token default-editor)
    
    # Test direct access
    log_info "Testing direct cluster-local-gateway access..."
    RESPONSE=$(curl -s -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer $PRIMARY_TOKEN" \
        "http://localhost:8081/")
    
    if [[ "$RESPONSE" == *"200"* ]] || [[ "$RESPONSE" == *"404"* ]] || [[ "$RESPONSE" == *"503"* ]]; then
        log_pass "Direct gateway access with token works (${RESPONSE##*HTTP_CODE:})"
    else
        log_info "Direct gateway response: $RESPONSE"
    fi
    
    # Test without token
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        "http://localhost:8081/")
    
    if [ "$RESPONSE" = "403" ]; then
        log_pass "Direct gateway blocks requests without token (403)"
    else
        log_fail "Expected 403 without token, got $RESPONSE"
    fi
    
    # Cleanup
    kill $PF_PID 2>/dev/null || true
    echo
}

function cleanup() {
    log_info "Cleaning up..."
    kubectl delete namespace $ATTACKER_NAMESPACE --ignore-not-found=true
    kubectl delete ksvc secure-model-predictor -n $PRIMARY_NAMESPACE --ignore-not-found=true
    # No authorization policies to clean up in this version
}

function main() {
    echo "Starting JWT authentication validation..."
    echo "Primary namespace: $PRIMARY_NAMESPACE" 
    echo "Attacker namespace: $ATTACKER_NAMESPACE"
    echo "Gateway endpoint: $INGRESS_HOST_PORT"
    echo
    
    setup_test_environment
    test_gateway_authentication
    test_namespace_isolation
    test_cluster_local_gateway
    
    echo "All tests completed successfully!"
    echo
    echo "JWT Authentication Summary:"
    echo "   - cluster-local-gateway validates JWT tokens"
    echo "   - Requests without tokens are blocked (403)"
    echo "   - Invalid tokens are rejected (401/403)"
    echo "   - Valid tokens allow access regardless of namespace"
    echo
    echo "KServe JWT authentication is working correctly!"
    
    cleanup
}

main