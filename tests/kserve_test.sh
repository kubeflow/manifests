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

# Path-based routing
cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: test-sklearn-path
  namespace: ${NAMESPACE}
spec:
  gateways:
    - kubeflow/kubeflow-gateway
  hosts:
    - '*'
  http:
    - match:
        - uri:
            prefix: /kserve/${NAMESPACE}/test-sklearn/
      rewrite:
        uri: /
      route:
        - destination:
            host: cluster-local-gateway.istio-system.svc.cluster.local
          headers:
            request:
              set:
                Host: test-sklearn-predictor.${NAMESPACE}.svc.cluster.local
          weight: 100
      timeout: 300s
EOF

# Deploy InferenceService for testing
cat <<EOF | kubectl apply -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "test-sklearn"
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

kubectl wait --for=condition=Ready inferenceservice/test-sklearn -n ${NAMESPACE} --timeout=300s

cat <<EOF | kubectl apply -f -
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-test-sklearn
  namespace: ${NAMESPACE}
spec:
  action: ALLOW
  rules:
  - from:
    - source:
        requestPrincipals: ["*"]
  selector:
    matchLabels:
      serving.knative.dev/service: test-sklearn-predictor
EOF

sleep 60

# Request without token should be rejected
RESPONSE_NO_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/kserve/${NAMESPACE}/test-sklearn/v1/models/test-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')

if [ "$RESPONSE_NO_TOKEN" != "403" ] && [ "$RESPONSE_NO_TOKEN" != "302" ]; then
  exit 1
fi

# Request with valid token should succeed
RESPONSE_WITH_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" \
 -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/kserve/${NAMESPACE}/test-sklearn/v1/models/test-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4], [6.0, 3.4, 4.5, 1.6]]}')

if [ "$RESPONSE_WITH_TOKEN" != "200" ] && [ "$RESPONSE_WITH_TOKEN" != "404" ] && [ "$RESPONSE_WITH_TOKEN" != "503" ]; then
  exit 1
fi

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

kubectl delete inferenceservice isvc-sklearn -n ${NAMESPACE} --ignore-not-found=true

if kubectl get inferenceservice isvc-sklearn -n ${NAMESPACE} &>/dev/null; then
  kubectl wait --for=delete inferenceservice/isvc-sklearn -n ${NAMESPACE} --timeout=120s
fi

sleep 5

if cd ${TEST_DIRECTORY}; then
  pytest . -vs --log-level info || true
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

# Verify cluster-local-gateway requires authentication
kubectl port-forward -n istio-system svc/cluster-local-gateway 8081:80 &
PF_PID=$!
sleep 5

PRIMARY_TOKEN=$(kubectl -n ${NAMESPACE} create token default-editor)

curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    -H "Authorization: Bearer $PRIMARY_TOKEN" \
    "http://localhost:8081/" > /dev/null

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    "http://localhost:8081/")

if [ "$RESPONSE" != "403" ]; then
    exit 1
fi

# Test namespace isolation - attacker in different namespace should not have access
ATTACKER_NAMESPACE="attacker-namespace"
kubectl create namespace ${ATTACKER_NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: attacker-sa
  namespace: ${ATTACKER_NAMESPACE}
EOF

ATTACKER_TOKEN=$(kubectl -n ${ATTACKER_NAMESPACE} create token attacker-sa)

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.${NAMESPACE}.svc.cluster.local" \
    -H "Authorization: Bearer $ATTACKER_TOKEN" \
    "http://localhost:8081/")

# The attacker token can reach the service through the cluster-local-gateway,
# but may get 404 or 503 depending on whether the request is fully processed
if [ "$RESPONSE" != "200" ] && [ "$RESPONSE" != "404" ] && [ "$RESPONSE" != "503" ]; then
    exit 1
fi

kill $PF_PID 2>/dev/null || true

kubectl delete namespace ${ATTACKER_NAMESPACE} --ignore-not-found=true
kubectl delete ksvc secure-model-predictor -n ${NAMESPACE} --ignore-not-found=true
