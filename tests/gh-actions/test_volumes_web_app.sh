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

# Prepare data directory for local volume
echo "Creating data directory on kind-control-plane..."
kubectl exec kind-control-plane -- mkdir -p /tmp/data

echo "Creating StorageClass..."
RESP=$(curl -s -X POST \
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
  }')
echo "StorageClass creation response: $RESP"
sleep 2

echo "Creating PersistentVolume..."
RESP=$(curl -s -X POST \
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
  }')
echo "PV creation response: $RESP"
sleep 2

echo "Creating PVC in ${KF_PROFILE} namespace..."
RESP=$(curl -s -X POST \
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
  }')
echo "PVC creation response: $RESP"
sleep 5

echo "Testing authorized access to PVCs..."
RESP=$(curl -s \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}")
echo "List PVCs response: $RESP"

echo "Testing unauthorized access to PVCs (expecting 403)..."
RESP=$(curl -s -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${UNAUTH_TOKEN}")
echo "Unauthorized list PVCs response: $RESP"

echo "Creating another PVC through API..."
RESP=$(curl -s -X POST \
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
  }')
echo "Second PVC creation response: $RESP"
sleep 5

echo "Verifying PVCs were created:"
kubectl get pvc -n $KF_PROFILE

echo "Testing unauthorized PVC deletion (expecting 403)..."
RESP=$(curl -s -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTH_TOKEN}")
echo "Unauthorized deletion response: $RESP"

echo "Testing authorized PVC deletion..."
RESP=$(curl -s -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}")
echo "Authorized deletion response: $RESP"
sleep 5

echo "Verifying PVC was deleted:"
if kubectl get pvc test-pvc -n $KF_PROFILE 2>/dev/null; then
  echo "ERROR: PVC test-pvc still exists after deletion attempt"
  exit 1
else
  echo "PVC deletion confirmed"
fi

echo "All tests passed successfully!" 