#!/bin/bash
set -e

KF_PROFILE=${1:-kubeflow-user-example-com}

# Add timeout and retry for API availability check
echo "Checking API availability..."
MAX_RETRIES=12
RETRY_INTERVAL=5
API_READY=false

for i in $(seq 1 $MAX_RETRIES); do
  echo "Attempt $i of $MAX_RETRIES to connect to API endpoint..."
  if curl -s -o /dev/null -w "%{http_code}" "localhost:8080/volumes/api/storageclasses" | grep -q "40"; then
    echo "API endpoint is available"
    API_READY=true
    break
  fi
  echo "API endpoint not available yet, waiting ${RETRY_INTERVAL}s..."
  sleep $RETRY_INTERVAL
done

if [ "$API_READY" != "true" ]; then
  echo "ERROR: Volumes Web App API is not available after $(($MAX_RETRIES * $RETRY_INTERVAL))s"
  exit 1
fi

echo "Creating service account tokens..."
TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
if [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to create token for default-editor in $KF_PROFILE namespace"
  kubectl get serviceaccount -n $KF_PROFILE
  exit 1
fi

UNAUTH_TOKEN="$(kubectl -n default create token default)"
if [ -z "$UNAUTH_TOKEN" ]; then
  echo "ERROR: Failed to create token for default in default namespace"
  exit 1
fi

echo "Creating StorageClass..."
curl -s -o /dev/null -w "%{http_code}" \
  -X POST \
  "localhost:8080/volumes/api/storageclasses" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "standard",
    "annotations": {
      "storageclass.kubernetes.io/is-default-class": "true"
    },
    "provisioner": "rancher.io/local-path",
    "reclaimPolicy": "Delete",
    "volumeBindingMode": "WaitForFirstConsumer"
  }' | grep -q "200"


echo "Creating data directory on kind-control-plane..."
kubectl exec kind-control-plane -- mkdir -p /tmp/data


echo "Creating PersistentVolume..."
curl -s -o /dev/null -w "%{http_code}" \
  -X POST \
  "localhost:8080/volumes/api/persistentvolumes" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-pv",
    "spec": {
      "capacity": {
        "storage": "1Gi"
      },
      "accessModes": ["ReadWriteOnce"],
      "persistentVolumeReclaimPolicy": "Delete",
      "volumeMode": "Filesystem",
      "local": {
        "path": "/tmp/data"
      },
      "nodeAffinity": {
        "required": {
          "nodeSelectorTerms": [{
            "matchExpressions": [{
              "key": "kubernetes.io/hostname",
              "operator": "In",
              "values": ["kind-control-plane"]
            }]
          }]
        }
      }
    }
  }' | grep -q "200"


echo "Creating PVC in ${KF_PROFILE} namespace..."
curl -s -o /dev/null -w "%{http_code}" \
  -X POST \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-pvc",
    "namespace": "'${KF_PROFILE}'",
    "spec": {
      "accessModes": ["ReadWriteOnce"],
      "resources": {
        "requests": {
          "storage": "1Gi"
        }
      },
      "storageClassName": "standard"
    }
  }' | grep -q "200"

echo "Waiting for PVC to be created..."
sleep 5

echo "Testing authorized access to PVCs..."
curl -s -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" | grep -q "200"

echo "Testing unauthorized access to PVCs (expecting 403)..."
curl -s -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${UNAUTH_TOKEN}" | grep -q "403"

echo "Creating another PVC through API..."
curl -s -o /dev/null -w "%{http_code}" \
  -X POST \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api-created-pvc",
    "namespace": "'${KF_PROFILE}'",
    "spec": {
      "accessModes": ["ReadWriteOnce"],
      "resources": {
        "requests": {
          "storage": "1Gi"
        }
      },
      "storageClassName": "standard"
    }
  }' | grep -q "200"

echo "Waiting for PVC to be created..."
sleep 5
echo "Verifying PVC was created:"
kubectl get pvc api-created-pvc -n $KF_PROFILE

echo "Testing unauthorized PVC deletion (expecting 403)..."
curl -s -o /dev/null -w "%{http_code}" \
  -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTH_TOKEN}" | grep -q "403"

echo "Testing authorized PVC deletion..."
curl -s -o /dev/null -w "%{http_code}" \
  -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" | grep -q "200"

echo "Waiting for PVC to be deleted..."
sleep 5
echo "Verifying PVC was deleted (should fail):"
kubectl get pvc test-pvc -n $KF_PROFILE 2>/dev/null && exit 1 

echo "All tests passed successfully!" 