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

curl --fail --show-error -X POST \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN" \
  -d "{
    \"name\": \"test-pvc\",
    \"namespace\": \"${KF_PROFILE}\",
    \"spec\": {
      \"accessModes\": [\"ReadWriteOnce\"],
      \"resources\": {
        \"requests\": {
          \"storage\": \"1Gi\"
        }
      },
      \"storageClassName\": \"${STORAGE_CLASS_NAME}\"
    }
  }"

curl --fail --show-error \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

UNAUTHORIZED_STATUS=$(curl --fail --silent --show-error -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}")
[[ "$UNAUTHORIZED_STATUS" == "403" ]] || exit 1

curl --fail --show-error -X POST \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN" \
  -d "{
    \"name\": \"api-created-pvc\",
    \"namespace\": \"${KF_PROFILE}\",
    \"spec\": {
      \"accessModes\": [\"ReadWriteOnce\"],
      \"resources\": {
        \"requests\": {
          \"storage\": \"1Gi\"
        }
      },
      \"storageClassName\": \"${STORAGE_CLASS_NAME}\"
    }
  }" > /dev/null

curl --fail --show-error -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

if ! kubectl get pvc test-pvc -n $KF_PROFILE > /dev/null 2>&1; then
  UNAUTHORIZED_DELETE_RESPONSE=$(curl --fail --silent --show-error \
    "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"
  
  if [[ "$UNAUTHORIZED_DELETE_RESPONSE" == *"not found"* || "$UNAUTHORIZED_DELETE_RESPONSE" == *"\"code\":404"* ]]; then
    echo "ERROR: PVC was deleted by unauthorized request or is missing"
    exit 1
  fi
fi

curl --fail --show-error -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"

DELETE_STATUS=$(curl --fail --show-error -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: $XSRFTOKEN" -H "Cookie: XSRF-TOKEN=$XSRFTOKEN"
[[ "$DELETE_STATUS" == "404" ]] || {
  echo "Failed to delete PVC: got status $DELETE_STATUS instead of 404"
  exit 1
}

kubectl get pvc test-pvc -n $KF_PROFILE 2>/dev/null && exit 1
