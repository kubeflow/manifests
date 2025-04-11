#!/bin/bash
set -euxo pipefail

KF_PROFILE=${1:-kubeflow-user-example-com}

curl -s -o /dev/null -w "%{http_code}" "localhost:8080/volumes/api/storageclasses" | grep -q "40" || exit 1

TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
UNAUTHORIZED_TOKEN="$(kubectl -n default create token default)"

CSRF_COOKIE=$(curl -s -c - "localhost:8080/volumes/" | grep XSRF-TOKEN | cut -f 7)
[ -n "$CSRF_COOKIE" ] || exit 1
CSRF_HEADER=${CSRF_COOKIE}

STORAGE_CLASS_NAME="standard"
kubectl get storageclass $STORAGE_CLASS_NAME > /dev/null 2>&1 || exit 1

[[ $(curl -s -o /dev/null -w "%{http_code}" -X POST \
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
  }") == "200" ]] || exit 1

[[ $(curl -s -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}") == "200" ]] || exit 1

[[ $(curl -s -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}") == "403" ]] || exit 1

[[ $(curl -s -o /dev/null -w "%{http_code}" -X POST \
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
  }") == "200" ]] || exit 1

[[ $(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}") == "403" ]] || exit 1

[[ $(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}") == "200" ]] || exit 1

for i in {1..5}; do
  ! kubectl get pvc test-pvc -n $KF_PROFILE &>/dev/null && break
  [ $i -eq 5 ] && exit 1
  sleep 3
done 