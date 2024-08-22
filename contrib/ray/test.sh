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

# Download Istioctl and its manifests.
# export ISTIO_VERSION=1.21.1
# curl -L https://istio.io/downloadIstio | sh -
# cd istio-1.21.1
# export PATH=$PWD/bin:$PATH

# # Install Istio with:
# #   1. 100% trace sampling for demo purposes.
# #   2. "sanitize_te" disabled for proper gRPC interception. This is required by Istio 1.21.0 (https://github.com/istio/istio/issues/49685).
# #   3. TLS 1.3 enabled.
# istioctl install -y -f - <<EOF
# apiVersion: install.istio.io/v1alpha1
# kind: IstioOperator
# metadata:
#   namespace: kubeflow
# spec:
#   meshConfig:
#     defaultConfig:
#       tracing:
#         sampling: 100
#       runtimeValues:
#         envoy.reloadable_features.sanitize_te: "false"
#     meshMTLS:
#       minProtocolVersion: TLSV1_3
# EOF

# cd 

# Install KubeRay operator
kustomize build kuberay-operator/overlays/standalone | kubectl -n $NAMESPACE apply --server-side -f -

kubectl label namespace kubeflow istio-injection=enabled

# Wait for the operator to be ready.
kubectl -n $NAMESPACE wait --for=condition=available --timeout=600s deploy/kuberay-operator
kubectl -n $NAMESPACE get pod -l app.kubernetes.io/component=kuberay-operator

# Create a RayCluster Headless serivice
kubectl -n $NAMESPACE apply -f raycluster_istio_headless_svc.yaml

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
