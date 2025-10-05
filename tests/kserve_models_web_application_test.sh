#!/bin/bash
set -euxo pipefail


KF_PROFILE=${1:-kubeflow-user-example-com}
TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
BASE_URL="localhost:8080/kserve-endpoints"

cat <<EOF | kubectl apply -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "sklearn-iris"
  namespace: ${KF_PROFILE}
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

kubectl wait --for=condition=Ready inferenceservice/sklearn-iris -n ${KF_PROFILE} --timeout=120s
kubectl get inferenceservice sklearn-iris -n ${KF_PROFILE}

# Get XSRF token for API calls
curl -s "http://${BASE_URL}/" \
  -H "Authorization: Bearer ${TOKEN}" \
  -v -c /tmp/kserve_xcrf.txt 2>&1 | grep -i "set-cookie"
XSRFTOKEN=$(grep XSRF-TOKEN /tmp/kserve_xcrf.txt | awk '{print $NF}')

RESPONSE=$(curl -s --fail-with-body \
  "${BASE_URL}/api/namespaces/${KF_PROFILE}/inferenceservices" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${XSRFTOKEN}" \
  -H "Cookie: XSRF-TOKEN=${XSRFTOKEN}")

echo "$RESPONSE" | grep -q "sklearn-iris" || exit 1
kubectl get inferenceservice sklearn-iris -n ${KF_PROFILE} || exit 1
READY=$(kubectl get isvc sklearn-iris -n ${KF_PROFILE} -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')
[[ "$READY" == "True" ]] || {
  echo "FAILURE: InferenceService Ready status is: $READY"
  exit 1
}

kubectl delete inferenceservice sklearn-iris -n ${KF_PROFILE} || exit 1

# Test unauthorized access
TOKEN="$(kubectl -n default create token default)"
BASE_URL="localhost:8080/kserve-endpoints"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/namespaces/${KF_PROFILE}/inferenceservices" -H "Authorization: Bearer ${TOKEN}")
[[ "$HTTP_CODE" == "403" || "$HTTP_CODE" == "401" ]] || { echo "FAILURE: Expected 401/403, got $HTTP_CODE"; exit 1; }
echo "Test succeeded. Token from unauthorized ServiceAccount cannot list InferenceServices in $KF_PROFILE namespace."
