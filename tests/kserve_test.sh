#!/bin/bash
set -euxo pipefail

NAMESPACE=${1:-kubeflow-user-example-com}
SCRIPT_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if ! command -v pytest &> /dev/null; then
  pip install "pytest>=7.0.0" "kserve>=0.16.0" "kubernetes>=18.20.0" "requests>=2.18.4"
fi

export KSERVE_INGRESS_HOST_PORT=${KSERVE_INGRESS_HOST_PORT:-localhost:8080}
export KSERVE_M2M_TOKEN="$(kubectl -n ${NAMESPACE} create token default-editor)"
export KSERVE_TEST_NAMESPACE=${NAMESPACE}

# ============================================================
# Test 1: Model Inference via KServe SDK
# ============================================================
# Runs kserve_sklearn_test.py which:
#   1. Deploys an sklearn InferenceService (isvc-sklearn)
#   2. Creates an Istio AuthorizationPolicy to allow authenticated traffic
#   3. Runs a prediction and validates the output
#
# The test uses Host-header routing:
#   Host: isvc-sklearn.<namespace>.example.com
# which matches the Knative-managed VirtualService.
#
# An AuthorizationPolicy is created inside the test because Kubeflow
# installs a mesh-wide global-deny-all policy that blocks all pod traffic
# unless an explicit ALLOW is present.
pytest "${SCRIPT_DIRECTORY}/kserve_sklearn_test.py" -vs --log-level info

# ============================================================
# Test 2: Ingress Gateway Security (AuthorizationPolicy)
# ============================================================
# The AuthorizationPolicy (allow-isvc-sklearn) was created by pytest above.
# It allows requests with any valid requestPrincipal to reach the predictor.
#
# This test validates:
#   a) Requests WITHOUT a token → should be blocked (403/302)
#   b) Requests WITH a valid token → should succeed (200)
#
# We use Host-header routing to match the Knative-managed VirtualService.

sleep 10

HOST_HEADER="Host: isvc-sklearn.${NAMESPACE}.example.com"

# Request without token should be rejected
RESPONSE_NO_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" \
 -H "${HOST_HEADER}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/isvc-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')

if [ "$RESPONSE_NO_TOKEN" != "403" ] && [ "$RESPONSE_NO_TOKEN" != "302" ]; then
  echo "FAIL: Expected 403/302 without token, got $RESPONSE_NO_TOKEN"
  exit 1
fi

# Request with valid token should succeed
RESPONSE_WITH_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" \
 -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
 -H "${HOST_HEADER}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/isvc-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4], [6.0, 3.4, 4.5, 1.6]]}')

if [ "$RESPONSE_WITH_TOKEN" != "200" ] && [ "$RESPONSE_WITH_TOKEN" != "404" ] && [ "$RESPONSE_WITH_TOKEN" != "503" ]; then
  echo "FAIL: Expected 200/404/503 with token, got $RESPONSE_WITH_TOKEN"
  exit 1
fi

# ============================================================
# Test 3: KServe Models Web Application API
# ============================================================
kubectl wait --for=condition=Available --timeout=300s -n kubeflow deployment/kserve-models-web-app

TOKEN="$(kubectl -n ${NAMESPACE} create token default-editor)"
BASE_URL="localhost:8080/kserve-endpoints"

cat <<EOF | kubectl apply -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "sklearn-iris"
  namespace: ${NAMESPACE}
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

kubectl wait --for=condition=Ready inferenceservice/sklearn-iris -n ${NAMESPACE} --timeout=120s
kubectl get inferenceservice sklearn-iris -n ${NAMESPACE}

# Get XSRF token for API calls
curl -s "http://${BASE_URL}/" \
  -H "Authorization: Bearer ${TOKEN}" \
  -v -c /tmp/kserve_xcrf.txt 2>&1 | grep -i "set-cookie"
XSRFTOKEN=$(grep XSRF-TOKEN /tmp/kserve_xcrf.txt | awk '{print $NF}')

RESPONSE=$(curl -s --fail-with-body \
  "${BASE_URL}/api/namespaces/${NAMESPACE}/inferenceservices" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${XSRFTOKEN}" \
  -H "Cookie: XSRF-TOKEN=${XSRFTOKEN}")

echo "$RESPONSE" | grep -q "sklearn-iris" || exit 1
kubectl get inferenceservice sklearn-iris -n ${NAMESPACE} || exit 1
READY=$(kubectl get isvc sklearn-iris -n ${NAMESPACE} -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')
[[ "$READY" == "True" ]] || {
  echo "FAILURE: InferenceService sklearn-iris Ready status is: $READY"
  exit 1
}

kubectl delete inferenceservice sklearn-iris -n ${NAMESPACE} || exit 1

# Test unauthorized access to models web app
UNAUTH_TOKEN="$(kubectl -n default create token default)"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/namespaces/${NAMESPACE}/inferenceservices" -H "Authorization: Bearer ${UNAUTH_TOKEN}")
[[ "$HTTP_CODE" == "403" || "$HTTP_CODE" == "401" ]] || { echo "FAILURE: Expected 401/403, got $HTTP_CODE"; exit 1; }
echo "Models Web App: Token from unauthorized ServiceAccount cannot list InferenceServices in $NAMESPACE namespace."

# ============================================================
# Test 4: Knative Service authentication via cluster-local-gateway
# ============================================================
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: secure-model-predictor
  namespace: ${NAMESPACE}
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

kubectl wait --for=condition=Ready ksvc/secure-model-predictor -n ${NAMESPACE} --timeout=120s

# Verify unauthenticated access is blocked
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    "http://${KSERVE_INGRESS_HOST_PORT}/")

if [ "$RESPONSE" != "403" ]; then
    echo "FAIL: Unauthenticated access should return 403, got $RESPONSE"
    exit 1
fi

# Verify invalid token is rejected
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    -H "Authorization: Bearer invalid-token" \
    "http://${KSERVE_INGRESS_HOST_PORT}/")

if [ "$RESPONSE" != "401" ] && [ "$RESPONSE" != "403" ]; then
    echo "FAIL: Invalid token should return 401/403, got $RESPONSE"
    exit 1
fi

# ============================================================
# Test 5: Cluster-local-gateway requires authentication
# ============================================================
kubectl port-forward -n istio-system svc/cluster-local-gateway 8081:80 &
PF_PID=$!
sleep 5

curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    -H "Authorization: Bearer $KSERVE_M2M_TOKEN" \
    "http://localhost:8081/" > /dev/null

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    "http://localhost:8081/")

if [ "$RESPONSE" != "403" ]; then
    echo "FAIL: Cluster-local-gateway unauthenticated access should return 403, got $RESPONSE"
    exit 1
fi

# ============================================================
# Test 6: Namespace isolation - attacker should NOT have access
# ============================================================
ATTACKER_NAMESPACE="attacker-namespace"
kubectl create namespace ${ATTACKER_NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: attacker-service-account
  namespace: ${ATTACKER_NAMESPACE}
EOF

# Unauthenticated request from attacker namespace should be REJECTED
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    "http://localhost:8081/")

if [ "$RESPONSE" == "200" ]; then
    echo "FAIL: Unauthenticated attacker namespace request should be rejected, got $RESPONSE"
    exit 1
fi

# Authenticated request from attacker namespace should ALSO be REJECTED
ATTACKER_TOKEN=$(kubectl -n ${ATTACKER_NAMESPACE} create token attacker-service-account)

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    -H "Authorization: Bearer $ATTACKER_TOKEN" \
    "http://localhost:8081/")

if [ "$RESPONSE" == "200" ]; then
    echo "FAIL: Attacker namespace token should be rejected (namespace isolation), got $RESPONSE"
    exit 1
fi

# ============================================================
# Cleanup
# ============================================================
kill $PF_PID 2>/dev/null || true

kubectl delete namespace ${ATTACKER_NAMESPACE} --ignore-not-found=true
kubectl delete ksvc secure-model-predictor -n ${NAMESPACE} --ignore-not-found=true
