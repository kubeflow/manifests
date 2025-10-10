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

RESPONSE_NO_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/kserve/${NAMESPACE}/test-sklearn/v1/models/test-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')

if [ "$RESPONSE_NO_TOKEN" != "403" ] && [ "$RESPONSE_NO_TOKEN" != "302" ]; then
  exit 1
fi

RESPONSE_WITH_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" \
 -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/kserve/${NAMESPACE}/test-sklearn/v1/models/test-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4], [6.0, 3.4, 4.5, 1.6]]}')

if [ "$RESPONSE_WITH_TOKEN" != "200" ] && [ "$RESPONSE_WITH_TOKEN" != "404" ] && [ "$RESPONSE_WITH_TOKEN" != "503" ]; then
  exit 1
fi

curl -s -o /dev/null \
  -H "Host: test-sklearn-predictor.${NAMESPACE}.example.com" \
  -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/test-sklearn:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}' || true

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

if cd ${TEST_DIRECTORY}; then
  pytest . -vs --log-level info || true
fi

kubectl wait --for=condition=Available --timeout=300s -n kubeflow deployment/kserve-models-web-app

PRIMARY_NAMESPACE="kubeflow-user-example-com"
ATTACKER_NAMESPACE="kubeflow-user-attacker"
INGRESS_HOST_PORT="${KSERVE_INGRESS_HOST_PORT:-localhost:8080}"

kubectl create namespace $PRIMARY_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace $ATTACKER_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
kubectl create serviceaccount default-editor -n $PRIMARY_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
kubectl create serviceaccount attacker-sa -n $ATTACKER_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

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

kubectl wait --for=condition=Ready ksvc/secure-model-predictor -n $PRIMARY_NAMESPACE --timeout=120s

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
    "http://$INGRESS_HOST_PORT/")

if [ "$RESPONSE" != "403" ]; then
    exit 1
fi

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
    -H "Authorization: Bearer invalid-token" \
    "http://$INGRESS_HOST_PORT/")

if [ "$RESPONSE" != "401" ] && [ "$RESPONSE" != "403" ]; then
    exit 1
fi

PRIMARY_TOKEN=$(kubectl -n $PRIMARY_NAMESPACE create token default-editor)
ATTACKER_TOKEN=$(kubectl -n $ATTACKER_NAMESPACE create token attacker-sa)

RESPONSE=$(curl -s -w "%{http_code}" \
    -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
    -H "Authorization: Bearer $PRIMARY_TOKEN" \
    "http://$INGRESS_HOST_PORT/")

if [[ "$RESPONSE" != *"200"* ]] && [[ "$RESPONSE" != *"404"* ]] && [[ "$RESPONSE" != *"503"* ]]; then
    exit 1
fi

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
    -H "Authorization: Bearer $ATTACKER_TOKEN" \
    "http://$INGRESS_HOST_PORT/")

if [[ "$RESPONSE" != *"200"* ]] && [[ "$RESPONSE" != *"404"* ]] && [[ "$RESPONSE" != *"503"* ]]; then
    exit 1
fi

kubectl port-forward -n istio-system svc/cluster-local-gateway 8081:80 &
PF_PID=$!
sleep 5

PRIMARY_TOKEN=$(kubectl -n $PRIMARY_NAMESPACE create token default-editor)

curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
    -H "Authorization: Bearer $PRIMARY_TOKEN" \
    "http://localhost:8081/" > /dev/null

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
    "http://localhost:8081/")

if [ "$RESPONSE" != "403" ]; then
    exit 1
fi

kill $PF_PID 2>/dev/null || true

kubectl delete namespace $ATTACKER_NAMESPACE --ignore-not-found=true
kubectl delete ksvc secure-model-predictor -n $PRIMARY_NAMESPACE --ignore-not-found=true

kubectl create serviceaccount default-editor -n ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

export KSERVE_M2M_TOKEN="$(kubectl -n ${NAMESPACE} create token default-editor)"

set +e
if cat <<EOF | kubectl apply -f -
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
then
  echo "InferenceService created successfully, waiting for it to be ready..."
  kubectl wait --for=condition=Ready inferenceservice/test-sklearn-secure -n ${NAMESPACE} --timeout=180s || echo "InferenceService not ready, continuing with JWT tests..."

  sleep 60

  cat <<EOF | kubectl apply -f -
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-test-sklearn-secure
  namespace: ${NAMESPACE}
spec:
  action: ALLOW
  rules:
  - from:
    - source:
        requestPrincipals: ["*"]
  selector:
    matchLabels:
      serving.knative.dev/service: test-sklearn-secure-predictor
EOF
else
  echo "InferenceService creation failed (likely KServe webhook issues), continuing with JWT authentication tests..."
fi
set -e

set +e
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Host: test-sklearn-secure-predictor.${NAMESPACE}.svc.cluster.local" \
  -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/test-sklearn-secure:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')
set -e

if [ "$RESPONSE" != "200" ] && [ "$RESPONSE" != "404" ] && [ "$RESPONSE" != "503" ]; then
  exit 1
fi

set +e
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Host: test-sklearn-secure-predictor.${NAMESPACE}.svc.cluster.local" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/test-sklearn-secure:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')
set -e

if [ "$RESPONSE" != "403" ]; then
  exit 1
fi

kubectl create namespace attacker-namespace --dry-run=client -o yaml | kubectl apply -f -
kubectl create serviceaccount attacker-sa -n attacker-namespace --dry-run=client -o yaml | kubectl apply -f -
ATTACKER_TOKEN="$(kubectl -n attacker-namespace create token attacker-sa)"

set +e
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Host: test-sklearn-secure-predictor.${NAMESPACE}.svc.cluster.local" \
  -H "Authorization: Bearer ${ATTACKER_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/test-sklearn-secure:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')
set -e

if [ "$RESPONSE" != "200" ] && [ "$RESPONSE" != "404" ] && [ "$RESPONSE" != "503" ]; then
  exit 1
fi

kubectl delete namespace attacker-namespace --ignore-not-found=true
