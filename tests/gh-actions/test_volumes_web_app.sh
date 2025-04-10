#!/bin/bash
set -e

KF_PROFILE=${1:-kubeflow-user-example-com}

TOKEN="$(kubectl -n $KF_PROFILE create token default-editor)"
UNAUTH_TOKEN="$(kubectl -n default create token default)"

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


kubectl exec -it kind-control-plane -- mkdir -p /tmp/data


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

sleep 5

curl -s -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${TOKEN}" | grep -q "200"

curl -s -o /dev/null -w "%{http_code}" \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs" \
  -H "Authorization: Bearer ${UNAUTH_TOKEN}" | grep -q "403"

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

sleep 5
kubectl get pvc api-created-pvc -n $KF_PROFILE

curl -s -o /dev/null -w "%{http_code}" \
  -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${UNAUTH_TOKEN}" | grep -q "403"

curl -s -o /dev/null -w "%{http_code}" \
  -X DELETE \
  "localhost:8080/volumes/api/namespaces/${KF_PROFILE}/pvcs/test-pvc" \
  -H "Authorization: Bearer ${TOKEN}" | grep -q "200"

sleep 5
kubectl get pvc test-pvc -n $KF_PROFILE 2>/dev/null && exit 1 