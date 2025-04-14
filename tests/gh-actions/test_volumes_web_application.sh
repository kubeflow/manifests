#!/bin/bash
set -euxo pipefail

KF_PROFILE=${1:-kubeflow-user-example-com}
TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
UNAUTHORIZED_TOKEN="$(kubectl -n default create token default)"

curl --fail --show-error "localhost:8080/volumes/" -H "Authorization: Bearer ${TOKEN}" -v -c /tmp/xcrf.txt

echo /tmp/xcrf.txt
XSRFTOKEN=$(grep XSRF-TOKEN /tmp/xcrf.txt | cut -f 7)

STORAGE_CLASS_NAME="standard"
kubectl get storageclass $STORAGE_CLASS_NAME

curl --fail --show-error \
  "localhost:8080/volumes/api/storageclasses" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

echo "Creating test-pvc..."
curl --fail --show-error -X POST \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN" \
  -d "{
    \"name\": \"test-pvc\",
    \"namespace\": \"${KF_PROFILE}\",
    \"type\": \"new\",
    \"mode\": \"ReadWriteOnce\",
    \"size\": \"1Gi\",
    \"class\": \"${STORAGE_CLASS_NAME}\"
  }"

curl --fail --show-error \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

UNAUTHORIZED_STATUS=$(curl --silent --output /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}")
[[ "$UNAUTHORIZED_STATUS" == "403" ]] || exit 1

echo "Creating api-created-pvc..."
curl --fail --show-error -X POST \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN" \
  -d "{
    \"name\": \"api-created-pvc\",
    \"namespace\": \"${KF_PROFILE}\",
    \"type\": \"new\",
    \"mode\": \"ReadWriteOnce\",
    \"size\": \"1Gi\",
    \"class\": \"${STORAGE_CLASS_NAME}\"
  }"

echo "Testing unauthorized deletion..."
UNAUTH_DELETE_STATUS=$(curl --silent --output /dev/null -w "%{http_code}" -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN")
[[ "$UNAUTH_DELETE_STATUS" == "403" ]] || echo "Warning: Unexpected status code for unauthorized delete: $UNAUTH_DELETE_STATUS"

if ! kubectl get pvc test-pvc -n $KF_PROFILE > /dev/null 2>&1; then
  echo "ERROR: PVC 'test-pvc' not found after unauthorized deletion attempt"
  exit 1
fi

echo "Deleting test-pvc with authorized request..."
curl --fail --show-error -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

DELETE_STATUS=$(curl --silent --output /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN")
[[ "$DELETE_STATUS" == "404" ]] || {
  echo "Failed to delete PVC: got status $DELETE_STATUS instead of 404"
  exit 1
}

kubectl get pvc test-pvc -n $KF_PROFILE 2>/dev/null && {
  echo "ERROR: PVC 'test-pvc' still exists after deletion"
  exit 1
} || true

echo "Cleaning up api-created-pvc..."
curl --fail --show-error -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/api-created-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

echo "Test completed successfully!"