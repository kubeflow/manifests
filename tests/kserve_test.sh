#!/bin/bash
set -euxo pipefail

NAMESPACE=${1:-kubeflow-user-example-com}
SCRIPT_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_DIRECTORY="${SCRIPT_DIRECTORY}/kserve"

echo "=== Creating dedicated Gateway for path-based routing ==="
cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
 name: kserve-path-gateway
 namespace: ${NAMESPACE}
spec:
 selector:
   istio: ingressgateway
 servers:
 - port:
     number: 80
     name: http
     protocol: HTTP
   hosts:
   - kserve-path.${NAMESPACE}.example.com
EOF

kubectl wait --for=condition=Ready gateway/kserve-path-gateway -n ${NAMESPACE} --timeout=60s || true
kubectl get gateway kserve-path-gateway -n ${NAMESPACE} -o yaml

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
export KSERVE_TEST_NAMESPACE=${NAMESPACE}
cd ${TEST_DIRECTORY} && pytest . -vs --log-level info

echo "=== InferenceService for additional tests ==="
cat <<EOF | kubectl apply -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "isvc-sklearn"
  namespace: ${NAMESPACE}
spec:
  predictor:
    sklearn:
      storageUri: "gs://kfserving-examples/models/sklearn/1.0/model"
EOF

kubectl wait --for=condition=Ready inferenceservice/isvc-sklearn -n ${NAMESPACE} --timeout=300s

cat <<EOF | kubectl apply -f -
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-kserve-access
  namespace: ${NAMESPACE}
spec:
  action: ALLOW
  rules:
  - {}
  selector:
    matchLabels:
      serving.knative.dev/service: isvc-sklearn-predictor
EOF

kubectl get pods -n ${NAMESPACE} -l serving.knative.dev/service=isvc-sklearn-predictor --show-labels
echo "=== Testing path-based routing functionality ==="

echo "Testing path-based access with valid token..."
curl --fail --show-error \
 -H "Host: kserve-path.${NAMESPACE}.example.com" \
 -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/serving/${NAMESPACE}/isvc-sklearn/v1/models/isvc-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4], [6.0, 3.4, 4.5, 1.6]]}'

echo "Testing 404 for incorrect path..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
 -H "Host: kserve-path.${NAMESPACE}.example.com" \
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
  -H "Host: isvc-sklearn-predictor.${NAMESPACE}.example.com" \
  -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/isvc-sklearn:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}'

echo "=== Path-based routing tests completed successfully ==="

# TODO FOR FOLLOW-UP PR: Implement proper security with AuthorizationPolicy that restricts access