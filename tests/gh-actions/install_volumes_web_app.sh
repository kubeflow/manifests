#!/bin/bash
set -euxo pipefail

kubectl get sc standard > /dev/null 2>&1 || kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: rancher.io/local-path
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
EOF

cd apps/volumes-web-app/upstream
kustomize build overlays/istio | kubectl apply -f -
cd ../../../

sleep 5 