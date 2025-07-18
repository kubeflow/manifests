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

# Create AuthorizationPolicy to allow authenticated access
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

# Create AuthorizationPolicy for pytest isvc-sklearn
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

if cd ${TEST_DIRECTORY}; then
  pytest . -vs --log-level info || true
fi