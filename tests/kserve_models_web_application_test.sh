#!/bin/bash
set -euxo pipefail


KF_PROFILE=${1:-kubeflow-user-example-com}
TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
BASE_URL="localhost:8080/kserve-endpoints"

cat <<EOF | kubectl apply -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: "sklearn-iris-private"
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

kubectl wait --for=condition=Ready inferenceservice/sklearn-iris-private -n ${KF_PROFILE} --timeout=120s
kubectl get inferenceservice sklearn-iris-private -n ${KF_PROFILE}

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

echo "$RESPONSE" | grep -q "sklearn-iris-private" || exit 1
kubectl get inferenceservice sklearn-iris-private -n ${KF_PROFILE} || exit 1
READY=$(kubectl get isvc sklearn-iris-private -n ${KF_PROFILE} -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')
[[ "$READY" == "True" ]] || {
  echo "FAILURE: InferenceService Ready status is: $READY"
  exit 1
}

kubectl delete inferenceservice sklearn-iris-private -n ${KF_PROFILE} || exit 1
