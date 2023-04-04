#!/bin/bash

set -xe

kubectl create ns kubeflow || echo "namespace kubeflow already exists"
kustomize build bentoml-yatai-stack/default | kubectl apply --server-side -f -
kubectl -n kubeflow wait --for=condition=available --timeout=600s deploy/yatai-image-builder
kubectl -n kubeflow wait --for=condition=available --timeout=600s deploy/yatai-deployment
sleep 5
kubectl apply -n kubeflow -f example.yaml
sleep 5
kubectl -n kubeflow logs deploy/yatai-deployment
sleep 5
kubectl -n kubeflow wait --for=condition=available --timeout=600s deploy/test-yatai
kubectl -n kubeflow wait --for=condition=available --timeout=600s deploy/test-yatai-runner-0

kubectl -n kubeflow port-forward svc/test-yatai 3333:3000 &
PID=$!

function trap_handler {
 kill $PID
 kubectl -n kubeflow logs -l yatai.ai/bento-deployment=test-yatai --tail=100
 kubectl -n kubeflow delete -f example.yaml
 kustomize build bentoml-yatai-stack/default | kubectl delete -f -
}

trap trap_handler EXIT

sleep 5

output=$(curl --fail -X 'POST' http://localhost:3333/classify -d '[[0,1,2,3]]')
echo "output: '${output}'"
if [[ $output != *'[2]'* ]]; then
  echo "Test failed"
  exit 1
fi
