#!/bin/bash
set -euxo pipefail

NAMESPACE=${1:-kubeflow-user-example-com}
SCRIPT_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_DIRECTORY="${SCRIPT_DIRECTORY}/kserve"

echo "=== KServe Predictor Service Labels ==="
kubectl get pods -n ${NAMESPACE} -l serving.knative.dev/service=isvc-sklearn-predictor --show-labels || true

export KSERVE_INGRESS_HOST_PORT=${KSERVE_INGRESS_HOST_PORT:-localhost:8080}
export KSERVE_M2M_TOKEN="$(kubectl -n ${NAMESPACE} create token default-editor)"
export KSERVE_TEST_NAMESPACE=${NAMESPACE}

cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-sklearn-path
  namespace: ${NAMESPACE}
spec:
  parentRefs:
  - name: kubeflow-gateway
    namespace: istio-system
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /kserve/${NAMESPACE}/isvc-sklearn/
    filters:
    - type: URLRewrite
      urlRewrite:
        path:
          type: ReplacePrefixMatch
          replacePrefixMatch: /
    - type: RequestHeaderModifier
      requestHeaderModifier:
        set:
        - name: Host
          value: isvc-sklearn-predictor.${NAMESPACE}.svc.cluster.local
    backendRefs:
    - name: knative-local-gateway
      namespace: istio-system
      port: 80
      weight: 100
    timeouts:
      request: 300s
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
  - {}
  selector:
    matchLabels:
      serving.knative.dev/service: test-sklearn-predictor
EOF

sleep 60

echo "Testing path-based access with valid token..."
curl -v --fail --show-error \
 -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
 -H "Content-Type: application/json" \
 "http://${KSERVE_INGRESS_HOST_PORT}/kserve/${NAMESPACE}/test-sklearn/v1/models/test-sklearn:predict" \
 -d '{"instances": [[6.8, 2.8, 4.8, 1.4], [6.0, 3.4, 4.5, 1.6]]}'

echo "Testing direct service access still works..."
curl -v --fail --show-error \
  -H "Host: test-sklearn-predictor.${NAMESPACE}.example.com" \
  -H "Authorization: Bearer ${KSERVE_M2M_TOKEN}" \
  -H "Content-Type: application/json" \
  "http://${KSERVE_INGRESS_HOST_PORT}/v1/models/test-sklearn:predict" \
  -d '{"instances": [[6.8, 2.8, 4.8, 1.4]]}'

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
  - {}
  selector:
    matchLabels:
      serving.knative.dev/service: isvc-sklearn-predictor
EOF

kubectl delete inferenceservice isvc-sklearn -n ${NAMESPACE} --ignore-not-found=true

cd ${TEST_DIRECTORY} && pytest . -vs --log-level info

# TODO FOR FOLLOW-UP PR: Implement proper security with AuthorizationPolicy that restricts access