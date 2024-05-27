#!/bin/bash

set -euxo

NAMESPACE=kubeflow
TIMEOUT=120  # timeout in seconds
SLEEP_INTERVAL=30  # interval between checks in seconds
RAY_VERSION=2.23.0

function trap_handler {
  kill $PID
  # Delete RayCluster
  kubectl -n $NAMESPACE delete -f raycluster_example.yaml

  # Wait for all Ray Pods to be deleted.
  start_time=$(date +%s)
  while true; do
    pods=$(kubectl -n $NAMESPACE get pods -o json | jq '.items | length')
    if [ "$pods" -eq 1 ]; then
      break
    fi
    current_time=$(date +%s)
    elapsed_time=$((current_time - start_time))
    if [ "$elapsed_time" -ge "$TIMEOUT" ]; then
      echo "Timeout exceeded. Exiting loop."
      exit 1
    fi
    sleep $SLEEP_INTERVAL
  done

  # Delete KubeRay operator
  kustomize build kuberay-operator/base | kubectl -n $NAMESPACE delete -f -
}

trap trap_handler EXIT

# Install KubeRay operator
kustomize build kuberay-operator/overlays/standalone | kubectl -n $NAMESPACE apply --server-side -f -

# Wait for the operator to be ready.
kubectl -n $NAMESPACE wait --for=condition=available --timeout=600s deploy/kuberay-operator
kubectl -n $NAMESPACE get pod -l app.kubernetes.io/component=kuberay-operator

# Create a RayCluster custom resource.
kubectl -n $NAMESPACE apply -f raycluster_example.yaml

# Wait for the RayCluster to be ready.
sleep 5
kubectl -n $NAMESPACE wait --for=condition=ready pod -l ray.io/cluster=kubeflow-raycluster --timeout=900s
kubectl -n $NAMESPACE logs -l ray.io/cluster=kubeflow-raycluster,ray.io/node-type=head

# Forward the port of Dashboard
sleep 5
kubectl -n $NAMESPACE port-forward --address 0.0.0.0 svc/kubeflow-raycluster-head-svc 8265:8265 &
PID=$!
echo "Forward the port 8265 of Ray head in the background process: $PID"

# Send a curl command to test basic Ray functionality.
sleep 5
output=$(curl -H "Content-Type: application/json" localhost:8265/api/version)
echo "output: ${output}"

# output format: {"version": ..., "ray_version": RAY_VERSION, "ray_commit": ...}
if echo "${output}" | grep -q $RAY_VERSION; then
  echo "Test succeeded!"
else
  echo "Test failed!"
  exit 1
fi
