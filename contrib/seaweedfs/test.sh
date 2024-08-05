#!/usr/bin/env bash

set -xe

kubectl create ns kubeflow || echo "namespace kubeflow already exists"
kustomize build base/ | kubectl apply --server-side -f -
kubectl -n kubeflow wait --for=condition=available --timeout=600s deploy/seaweedfs

kubectl -n kubeflow port-forward svc/minio-service 8333:9000
echo "S3 endpoint available on localhost:8333" &

function trap_handler {
 kubectl -n kubeflow logs -l app=seaweedfs --tail=100
 kustomize build base/ | kubectl delete -f -
}

trap trap_handler EXIT
