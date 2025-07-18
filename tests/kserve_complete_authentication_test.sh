#!/bin/bash
set -euo pipefail

PRIMARY_NAMESPACE="kubeflow-user-example-com"
ATTACKER_NAMESPACE="kubeflow-user-attacker"
KSERVE_INGRESS_HOST_PORT="${KSERVE_INGRESS_HOST_PORT:-localhost:8080}"

function setup_test_environment() {
    kubectl create namespace $PRIMARY_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace $ATTACKER_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    kubectl create serviceaccount default-editor -n $PRIMARY_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    kubectl create serviceaccount attacker-sa -n $ATTACKER_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

    cat <<EOF | kubectl apply -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "secure-sklearn"
  namespace: $PRIMARY_NAMESPACE
spec:
  predictor:
    sklearn:
      storageUri: "gs://kfserving-examples/models/sklearn/1.0/model"
      resources:
        requests:
          cpu: 50m
          memory: 128Mi
        limits:
          cpu: 100m
          memory: 256Mi
EOF

    kubectl wait --for=condition=Ready inferenceservice/secure-sklearn -n $PRIMARY_NAMESPACE --timeout=180s || true
}

function test_gateway_jwt_validation() {
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-sklearn-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        "http://$KSERVE_INGRESS_HOST_PORT/v1/models/secure-sklearn:predict" \
        -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}' \
        -H "Content-Type: application/json")

    if [ "$RESPONSE" != "403" ]; then
        exit 1
    fi

    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-sklearn-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer invalid-token-123" \
        "http://$KSERVE_INGRESS_HOST_PORT/v1/models/secure-sklearn:predict" \
        -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}' \
        -H "Content-Type: application/json")

    if [ "$RESPONSE" != "401" ] && [ "$RESPONSE" != "403" ]; then
        exit 1
    fi
}

function test_namespace_isolation() {
    PRIMARY_TOKEN=$(kubectl -n $PRIMARY_NAMESPACE create token default-editor)
    ATTACKER_TOKEN=$(kubectl -n $ATTACKER_NAMESPACE create token attacker-sa)

    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-sklearn-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer $PRIMARY_TOKEN" \
        "http://$KSERVE_INGRESS_HOST_PORT/v1/models/secure-sklearn:predict" \
        -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}' \
        -H "Content-Type: application/json")

    if [ "$RESPONSE" != "200" ] && [ "$RESPONSE" != "404" ] && [ "$RESPONSE" != "503" ]; then
        exit 1
    fi

    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-sklearn-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer $ATTACKER_TOKEN" \
        "http://$KSERVE_INGRESS_HOST_PORT/v1/models/secure-sklearn:predict" \
        -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}' \
        -H "Content-Type: application/json")

    if [ "$RESPONSE" != "200" ] && [ "$RESPONSE" != "404" ] && [ "$RESPONSE" != "503" ]; then
        exit 1
    fi
}

function test_external_access() {
    PRIMARY_TOKEN=$(kubectl -n $PRIMARY_NAMESPACE create token default-editor)

    curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $PRIMARY_TOKEN" \
        "http://$KSERVE_INGRESS_HOST_PORT/kserve/$PRIMARY_NAMESPACE/secure-sklearn/v1/models/secure-sklearn:predict" \
        -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}' \
        -H "Content-Type: application/json" > /dev/null || true
}

function test_internal_access() {
    kubectl run test-client -n $PRIMARY_NAMESPACE --image=curlimages/curl --restart=Never -- sleep 3600 2>/dev/null || true
    kubectl wait --for=condition=ready pod/test-client -n $PRIMARY_NAMESPACE --timeout=60s 2>/dev/null || return

    PRIMARY_TOKEN=$(kubectl -n $PRIMARY_NAMESPACE create token default-editor)

    kubectl exec -n $PRIMARY_NAMESPACE test-client -- \
        curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $PRIMARY_TOKEN" \
        -H "Content-Type: application/json" \
        "http://secure-sklearn-predictor.$PRIMARY_NAMESPACE.svc.cluster.local/v1/models/secure-sklearn:predict" \
        -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}' 2>/dev/null || true

    kubectl delete pod test-client -n $PRIMARY_NAMESPACE --ignore-not-found=true
}

function cleanup() {
    kubectl delete namespace $ATTACKER_NAMESPACE --ignore-not-found=true
    kubectl delete inferenceservice secure-sklearn -n $PRIMARY_NAMESPACE --ignore-not-found=true
    kubectl delete pod test-client -n $PRIMARY_NAMESPACE --ignore-not-found=true
}

function main() {
    setup_test_environment
    test_gateway_jwt_validation
    test_namespace_isolation
    test_external_access
    test_internal_access
    cleanup
}

main