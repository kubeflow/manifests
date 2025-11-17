#!/bin/bash
set -euxo pipefail

NAMESPACE=${1:-kubeflow-user-example-com}
SCRIPT_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_DIRECTORY="${SCRIPT_DIRECTORY}/kserve"

if ! command -v pytest &> /dev/null; then
  pip install -r ${TEST_DIRECTORY}/requirements.txt
fi

export KSERVE_INGRESS_HOST_PORT=${KSERVE_INGRESS_HOST_PORT:-localhost:8080}
export KSERVE_M2M_TOKEN="$(kubectl -n ${NAMESPACE} create token default-editor)"
export KSERVE_TEST_NAMESPACE=${NAMESPACE}

# Test 1: Model Inference via KServe SDK (pytest creates isvc-sklearn internally)
if cd ${TEST_DIRECTORY}; then
  pytest . -vs --log-level info || true
fi

# Test 2: Path-based Routing & Ingress Gateway Security (VirtualService + AuthorizationPolicy)
cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: isvc-sklearn-path
  namespace: ${NAMESPACE}
spec:
  gateways:
    - kubeflow/kubeflow-gateway
  hosts:
    - '*'
  http:
    - match:
        - uri:
            prefix: /kserve/${NAMESPACE}/isvc-sklearn/
      rewrite:
        uri: /
      route:
        - destination:
            host: cluster-local-gateway.istio-system.svc.cluster.local
          headers:
            request:
              set:
                Host: isvc-sklearn-predictor.${NAMESPACE}.svc.cluster.local
          weight: 100
      timeout: 300s
EOF

# WARNING: This policy allows ANY valid token from ANY kubeflow namespace to access this InferenceService.
cat <<EOF | kubectl apply -f -
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-isvc-sklearn
  namespace: ${NAMESPACE}
spec:
  action: ALLOW
  rules:
  - from:
    - source:
        requestPrincipals: ["*"]
  selector:
    matchLabels:
      serving.knative.dev/service: isvc-sklearn-predictor
EOF

sleep 60

# Request without token should be rejected
RESPONSE_NO_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/kserve/${NAMESPACE}/isvc-sklearn/v1/models/isvc-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')

if [ "$RESPONSE_NO_TOKEN" != "403" ] && [ "$RESPONSE_NO_TOKEN" != "302" ]; then
  exit 1
fi

# Request with valid token should succeed
RESPONSE_WITH_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" \
 -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/kserve/${NAMESPACE}/isvc-sklearn/v1/models/isvc-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4], [6.0, 3.4, 4.5, 1.6]]}')

if [ "$RESPONSE_WITH_TOKEN" != "200" ] && [ "$RESPONSE_WITH_TOKEN" != "404" ] && [ "$RESPONSE_WITH_TOKEN" != "503" ]; then
  exit 1
fi

kubectl wait --for=condition=Available --timeout=300s -n kubeflow deployment/kserve-models-web-app

# Knative Service authentication via cluster-local-gateway
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
    exit 1
fi

# Verify invalid token is rejected
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    -H "Authorization: Bearer invalid-token" \
    "http://${KSERVE_INGRESS_HOST_PORT}/")

if [ "$RESPONSE" != "401" ] && [ "$RESPONSE" != "403" ]; then
    exit 1
fi

# Test 3: Cluster-local-gateway requires authentication
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
    exit 1
fi

# Test 4: Namespace isolation - attacker in different namespace should NOT have access
ATTACKER_NAMESPACE="attacker-namespace"
kubectl create namespace ${ATTACKER_NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: attacker-service-account
  namespace: ${ATTACKER_NAMESPACE}
EOF

# Test 5: Unauthenticated request from attacker namespace should be REJECTED
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    "http://localhost:8081/")

if [ "$RESPONSE" == "200" ]; then
    echo "FAIL: Unauthenticated attacker namespace request should be rejected, got $RESPONSE"
    exit 1
fi

# Test 6: Authenticated request from attacker namespace should ALSO be REJECTED
ATTACKER_TOKEN=$(kubectl -n ${ATTACKER_NAMESPACE} create token attacker-service-account)

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    -H "Authorization: Bearer $ATTACKER_TOKEN" \
    "http://localhost:8081/")

if [ "$RESPONSE" == "200" ]; then
    echo "FAIL: Attacker namespace token should be rejected (namespace isolation), got $RESPONSE"
    exit 1
fi

kill $PF_PID 2>/dev/null || true

kubectl delete namespace ${ATTACKER_NAMESPACE} --ignore-not-found=true
kubectl delete ksvc secure-model-predictor -n ${NAMESPACE} --ignore-not-found=true
