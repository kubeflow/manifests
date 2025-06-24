#!/bin/bash
set -euxo pipefail

NAMESPACE=${1:-kubeflow-user-example-com}
SCRIPT_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_DIRECTORY="${SCRIPT_DIRECTORY}/kserve"

if ! command -v pytest &> /dev/null; then
  echo "Installing test dependencies..."
  pip install -r ${TEST_DIRECTORY}/requirements.txt
fi

export KSERVE_INGRESS_HOST_PORT=${KSERVE_INGRESS_HOST_PORT:-localhost:8080}
export KSERVE_M2M_TOKEN="$(kubectl -n ${NAMESPACE} create token default-editor)"
export KSERVE_TEST_NAMESPACE=${NAMESPACE}

# Deploy test InferenceService
cat <<EOF | kubectl apply -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "test-sklearn-secure"
  namespace: ${NAMESPACE}
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

kubectl wait --for=condition=Ready inferenceservice/test-sklearn-secure -n ${NAMESPACE} --timeout=300s

# Test 1: Access with valid token should succeed
echo "Test 1: Testing access with valid token..."
curl -v --fail --show-error \
  -H "Host: test-sklearn-secure-predictor.${NAMESPACE}.example.com" \
  -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/test-sklearn-secure:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}'

# Test 2: Access without token should fail with 403
echo "Test 2: Testing access without token (should fail with 403)..."
set +e
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Host: test-sklearn-secure-predictor.${NAMESPACE}.example.com" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/test-sklearn-secure:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')
set -e

if [ "$RESPONSE" = "403" ]; then
  echo "Test passed: Request without token was correctly rejected with 403"
else
  echo "Test failed: Expected 403 but got $RESPONSE"
  exit 1
fi

# Test 3: Access from unauthorized namespace should fail
echo "Test 3: Testing access from unauthorized namespace..."
kubectl create namespace attacker-namespace --dry-run=client -o yaml | kubectl apply -f -
kubectl create serviceaccount attacker-sa -n attacker-namespace --dry-run=client -o yaml | kubectl apply -f -
ATTACKER_TOKEN="$(kubectl -n attacker-namespace create token attacker-sa)"

set +e
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Host: test-sklearn-secure-predictor.${NAMESPACE}.example.com" \
  -H "Authorization: Bearer ${ATTACKER_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/test-sklearn-secure:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')
set -e

if [ "$RESPONSE" = "403" ]; then
  echo "Test passed: Request from unauthorized namespace was correctly rejected with 403"
else
  echo "Test failed: Expected 403 but got $RESPONSE"
  exit 1
fi

# Clean up
kubectl delete namespace attacker-namespace --ignore-not-found=true

echo "All security tests passed!"

# Run existing pytest tests
cd ${TEST_DIRECTORY} && pytest . -vs --log-level info