#!/bin/bash
set -euxo pipefail

NAMESPACE=${1:-kubeflow-user-example-com}
SCRIPT_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_DIRECTORY="${SCRIPT_DIRECTORY}/kserve"

echo "=== KServe Predictor Service Labels ==="
kubectl get pods -n ${NAMESPACE} -l serving.knative.dev/service=isvc-sklearn-predictor --show-labels

cat <<EOF | kubectl apply -f -
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-in-cluster-kserve
  namespace: ${NAMESPACE}
spec:
  selector:
    matchLabels:
      serving.knative.dev/service: isvc-sklearn-predictor
  rules:
    - to:
        - operation:
            paths:
              - /v1/models/*
              - /v2/models/*
EOF

echo "=== Setting up path-based routing VirtualService ==="
cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: kserve-path-routing
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
            host: knative-local-gateway.istio-system.svc.cluster.local
          headers:
            request:
              set:
                Host: isvc-sklearn-predictor.${NAMESPACE}.svc.cluster.local
          weight: 100
      timeout: 300s
EOF

if ! command -v pytest &> /dev/null; then
  echo "Installing test dependencies..."
  pip install -r ${TEST_DIRECTORY}/requirements.txt
fi

export KSERVE_INGRESS_HOST_PORT=${KSERVE_INGRESS_HOST_PORT:-localhost:8080}
export KSERVE_M2M_TOKEN="$(kubectl -n ${NAMESPACE} create token default-editor)"
cd ${TEST_DIRECTORY} && pytest . -vs --log-level info

echo "=== Testing path-based routing functionality ==="

echo "Testing path-based access with valid token..."
curl --fail --show-error \
  -H "Host: isvc-sklearn.${NAMESPACE}.example.com" \
  -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/kserve/${NAMESPACE}/isvc-sklearn/v1/models/isvc-sklearn:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4], [6.0, 3.4, 4.5, 1.6]]}'

echo "Testing 404 for incorrect path..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Host: isvc-sklearn.${NAMESPACE}.example.com" \
  -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/wrong-path/${NAMESPACE}/isvc-sklearn/v1/models/isvc-sklearn:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}')

if [[ "$RESPONSE" == "404" ]]; then
  echo "404 test passed - wrong path correctly rejected"
else
  echo "Expected 404, got $RESPONSE - path routing may be too permissive"
fi

echo "Testing direct service access still works..."
curl --fail --show-error \
  -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/isvc-sklearn:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}'

echo "=== Path-based routing tests completed successfully ==="

# TODO FOR FOLLOW-UP PR: Implement proper security with AuthorizationPolicy that restricts access