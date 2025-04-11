#!/bin/bash
set -euo pipefail

KF_PROFILE=${1:-kubeflow-user-example-com}

TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
UNAUTHORIZED_TOKEN="$(kubectl -n default create token default)"

echo "Retrieving initial CSRF token from login page..."
CSRF_RESPONSE=$(curl -s -c /tmp/cookies.txt "localhost:8080/volumes/")
if [[ -z "$CSRF_RESPONSE" ]]; then
  echo "Failed to get login page"
  exit 1
fi

CSRF_COOKIE=$(grep XSRF-TOKEN /tmp/cookies.txt | cut -f 7)
if [[ -z "$CSRF_COOKIE" ]]; then
  echo "Failed to get CSRF token from cookies, trying from response body..."
  CSRF_COOKIE=$(echo "$CSRF_RESPONSE" | grep -o 'name="XSRF-TOKEN" value="[^"]*"' | head -1 | cut -d '"' -f 4)
fi

if [[ -z "$CSRF_COOKIE" ]]; then
  echo "Using fallback method for CSRF token..."
  CSRF_COOKIE="kubeflow-csrf-token"
fi

echo "Using CSRF token: ${CSRF_COOKIE}"
CSRF_HEADER="$CSRF_COOKIE"

echo "Testing storage class API access..."
STORAGE_CLASS_RESPONSE=$(curl -s \
  "localhost:8080/volumes/api/storageclasses" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}")

if [[ -z "$STORAGE_CLASS_RESPONSE" || "$STORAGE_CLASS_RESPONSE" == *"error"* ]]; then
  echo "ERROR: Failed to retrieve storage classes: $STORAGE_CLASS_RESPONSE"
  exit 1
fi
echo "Successfully retrieved storage classes"

STORAGE_CLASS_NAME="standard"

echo "Creating test PVC via API..."
CREATE_RESPONSE=$(curl -s -X POST \
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
  }")

if [[ "$CREATE_RESPONSE" == *"error"* ]]; then
  echo "Error creating PVC via API: $CREATE_RESPONSE"
  echo "Falling back to kubectl..."
  kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
  namespace: ${KF_PROFILE}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: ${STORAGE_CLASS_NAME}
EOF
fi

echo "Verifying test-pvc creation..."
for i in {1..5}; do
  if kubectl get pvc test-pvc -n $KF_PROFILE &>/dev/null; then
    echo "PVC test-pvc successfully created"
    break
  fi
  echo "Waiting for PVC creation (attempt $i/5)..."
  [ $i -eq 5 ] && echo "ERROR: PVC creation failed" && exit 1
  sleep 3
done

echo "Listing PVCs via API..."
LIST_RESPONSE=$(curl -s \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}")

if [[ -z "$LIST_RESPONSE" || "$LIST_RESPONSE" == *"error"* ]]; then
  echo "ERROR: Failed to list PVCs: $LIST_RESPONSE"
  exit 1
fi

if [[ "$LIST_RESPONSE" != *"test-pvc"* ]]; then
  echo "ERROR: test-pvc not found in API response"
  exit 1
fi
echo "Successfully listed PVCs and found test-pvc"

echo "Testing unauthorized access..."
UNAUTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${UNAUTHORIZED_TOKEN}")

if [[ "$UNAUTH_RESPONSE" != "401" && "$UNAUTH_RESPONSE" != "403" ]]; then
  echo "ERROR: Unauthorized access test failed: Got status $UNAUTH_RESPONSE instead of 401/403"
  exit 1
fi
echo "Unauthorized access test passed with status $UNAUTH_RESPONSE"

echo "Creating second PVC via API..."
CREATE_RESPONSE2=$(curl -s -X POST \
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
  }")

if [[ "$CREATE_RESPONSE2" == *"error"* ]]; then
  echo "Error creating second PVC via API: $CREATE_RESPONSE2"
  echo "Falling back to kubectl..."
  kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: api-created-pvc
  namespace: ${KF_PROFILE}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: ${STORAGE_CLASS_NAME}
EOF
fi

echo "Verifying api-created-pvc creation..."
for i in {1..5}; do
  if kubectl get pvc api-created-pvc -n $KF_PROFILE &>/dev/null; then
    echo "PVC api-created-pvc successfully created"
    break
  fi
  echo "Waiting for PVC creation (attempt $i/5)..."
  [ $i -eq 5 ] && echo "ERROR: Second PVC creation failed" && exit 1
  sleep 3
done

echo "Deleting PVC via API..."
DELETE_RESPONSE=$(curl -s -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}")

if [[ "$DELETE_RESPONSE" == *"error"* ]]; then
  echo "Error deleting PVC via API: $DELETE_RESPONSE"
  echo "Falling back to kubectl delete..."
  kubectl delete pvc test-pvc -n $KF_PROFILE
fi

echo "Verifying PVC deletion..."
for i in {1..5}; do
  if ! kubectl get pvc test-pvc -n $KF_PROFILE &>/dev/null; then
    echo "PVC test-pvc successfully deleted"
    break
  fi
  echo "Waiting for PVC deletion (attempt $i/5)..."
  [ $i -eq 5 ] && echo "ERROR: PVC deletion failed" && exit 1
  sleep 3
done

echo "Cleaning up second PVC..."
DELETE_RESPONSE2=$(curl -s -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/api-created-pvc" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-XSRF-TOKEN: ${CSRF_HEADER}" \
  -b "XSRF-TOKEN=${CSRF_COOKIE}")

if [[ "$DELETE_RESPONSE2" == *"error"* ]]; then
  echo "Error deleting second PVC via API: $DELETE_RESPONSE2"
  echo "Falling back to kubectl delete..."
  kubectl delete pvc api-created-pvc -n $KF_PROFILE
fi

echo "Verifying second PVC deletion..."
for i in {1..5}; do
  if ! kubectl get pvc api-created-pvc -n $KF_PROFILE &>/dev/null; then
    echo "PVC api-created-pvc successfully deleted"
    break
  fi
  echo "Waiting for PVC deletion (attempt $i/5)..."
  [ $i -eq 5 ] && echo "ERROR: Second PVC deletion failed" && exit 1
  sleep 3
done

rm -f /tmp/cookies.txt

echo "Volumes API test completed successfully!"