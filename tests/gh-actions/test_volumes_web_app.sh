#!/bin/bash
set -e

KF_PROFILE=${1:-kubeflow-user-example-com}

curl -s -o /dev/null -w "%{http_code}" "localhost:8080/volumes/api/storageclasses" | grep -q "40" || {
  echo "Volumes Web App API is not available"
  exit 1
}

TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
UNAUTHORIZED_TOKEN="$(kubectl -n default create token default)"

CSRF_COOKIE=$(curl -s -c - "localhost:8080/volumes/" | grep XSRF-TOKEN | cut -f 7)
CSRF_HEADER=${CSRF_COOKIE}

STORAGE_CLASS_NAME="standard"
kubectl get storageclass $STORAGE_CLASS_NAME > /dev/null 2>&1

curl -s -X POST \
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

curl -s \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}" > /dev/null


UNAUTHORIZED_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}")
[[ "$UNAUTHORIZED_STATUS" == "403" ]] || exit 1

curl -s -X POST \
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

curl -s -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}" > /dev/null

UNAUTH_DELETE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}")
[[ "$UNAUTH_DELETE_STATUS" == "403" ]] || {
  echo "Unauthorized DELETE didn't return 403 Forbidden: got $UNAUTH_DELETE_STATUS"
  exit 1
}

kubectl get pvc test-pvc -n $KF_PROFILE > /dev/null 2>&1 || {
  echo "PVC was deleted by unauthorized request"
  exit 1
}

curl -s -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}" > /dev/null

DELETE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
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