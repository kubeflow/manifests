#!/bin/bash
set -euxo -pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

curl --fail --silent --show-error "localhost:8080/volumes/api/storageclasses"

TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
UNAUTHORIZED_TOKEN="$(kubectl -n default create token default)"

CSRF_COOKIE=$(curl -s -c - "localhost:8080/volumes/" | grep XSRF-TOKEN | cut -f 7)
CSRF_HEADER=${CSRF_COOKIE}

STORAGE_CLASS_NAME="standard"
kubectl get storageclass $STORAGE_CLASS_NAME > /dev/null 2>&1

curl --fail --silent --show-error -X POST \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}" \
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
  }" > /dev/null

curl --fail --silent --show-error \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}"


UNAUTHORIZED_STATUS=$(curl --fail --silent --show-error -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}")
[[ "$UNAUTHORIZED_STATUS" == "403" ]] || exit 1

curl --fail --silent --show-error -X POST \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}" \
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

curl --fail --silent --show-error -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}" > /dev/null

if ! kubectl get pvc test-pvc -n $KF_PROFILE > /dev/null 2>&1; then
  UNAUTHORIZED_DELETE_RESPONSE=$(curl --fail --silent --show-error \
    "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
    -b "XSRF-TOKEN=${CSRF_COOKIE}")
  
  if [[ "$UNAUTHORIZED_DELETE_RESPONSE" == *"not found"* || "$UNAUTHORIZED_DELETE_RESPONSE" == *"\"code\":404"* ]]; then
    echo "ERROR: PVC was deleted by unauthorized request or is missing"
    exit 1
  fi
fi

curl --fail --silent --show-error -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}" > /dev/null

DELETE_STATUS=$(curl --fail --silent --show-error -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}")
[[ "$DELETE_STATUS" == "404" ]] || {
  echo "Failed to delete PVC: got status $DELETE_STATUS instead of 404"
  exit 1
}

kubectl get pvc test-pvc -n $KF_PROFILE 2>/dev/null && exit 1

echo "All tests passed successfully!" 
