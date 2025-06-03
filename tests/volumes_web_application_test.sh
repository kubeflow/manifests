#!/bin/bash
set -euxo pipefail

KF_PROFILE=${1:-kubeflow-user-example-com}
TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
UNAUTHORIZED_TOKEN="$(kubectl -n default create token default)"

curl --fail --show-error "localhost:8080/volumes/" -H "Authorization: Bearer ${TOKEN}" -v -c /tmp/xcrf.txt

XSRFTOKEN=$(grep XSRF-TOKEN /tmp/xcrf.txt | cut -f 7)

STORAGE_CLASS_NAME="standard"
kubectl get storageclass $STORAGE_CLASS_NAME

curl --fail --show-error \
  "localhost:8080/volumes/api/storageclasses" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

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

kubectl get pvc test-pvc -n $KF_PROFILE

UNAUTHORIZED_STATUS=$(curl --silent --output /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN")

if [[ "$UNAUTHORIZED_STATUS" != "403" ]]; then
  echo "ERROR: Expected 403 status for unauthorized access, got $UNAUTHORIZED_STATUS"
  exit 1
fi

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

kubectl get pvc -n $KF_PROFILE

UNAUTH_DELETE_STATUS=$(curl --silent --output /dev/null -w "%{http_code}" -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN")

if [[ "$UNAUTH_DELETE_STATUS" != "403" ]]; then
  echo "WARNING: Expected 403 status for unauthorized deletion, got $UNAUTH_DELETE_STATUS"
fi

if ! kubectl get pvc test-pvc -n $KF_PROFILE > /dev/null 2>&1; then
  echo "ERROR: PVC 'test-pvc' not found after unauthorized deletion attempt"
  exit 1
fi

curl --fail --show-error -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

sleep 2  
if kubectl get pvc test-pvc -n $KF_PROFILE > /dev/null 2>&1; then
  echo "ERROR: PVC 'test-pvc' still exists after deletion"
  exit 1
fi

curl --fail --show-error -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/api-created-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

sleep 2  
if kubectl get pvc api-created-pvc -n $KF_PROFILE > /dev/null 2>&1; then
  echo "ERROR: PVC 'api-created-pvc' still exists after deletion"
  exit 1
fi

echo "Test completed successfully!"