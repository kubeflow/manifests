#!/bin/bash

set -euxo

NAMESPACE=$1
TIMEOUT=120  # timeout in seconds
SLEEP_INTERVAL=30  # interval between checks in seconds
SPARK_VERSION=3.5.2

start_time=$(date +%s)
for ((i=0; i<TIMEOUT; i+=2)); do
  if [[ $(kubectl get namespace $NAMESPACE --no-headers 2>/dev/null | wc -l) -eq 1 ]]; then
    echo "Namespace $NAMESPACE created."
    break
  fi
  
  current_time=$(date +%s)
  elapsed_time=$((current_time - start_time))
  
  if [ "$elapsed_time" -ge "$TIMEOUT" ]; then
    echo "Timeout exceeded. Namespace $NAMESPACE not created."
    exit 1
  fi
  
  echo "Waiting for namespace $NAMESPACE to be created..."
  sleep 2
done

echo "Namespace $NAMESPACE has been created!"

kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite

kubectl get namespaces --selector=istio-injection=enabled

# Install Spark operator
kustomize build spark-operator/overlays/standalone | kubectl -n kubeflow apply --server-side -f -

# Wait for the operator to be ready.
kubectl -n kubeflow wait --for=condition=available --timeout=600s deploy/spark-operator-controller
kubectl -n kubeflow get pod -l app.kubernetes.io/name=spark-operator

# Install Spark components
kubectl -n $NAMESPACE apply -f sparkapplication_example.yaml

# Wait for the Spark to be ready.
sleep 5
kubectl -n $NAMESPACE wait --for=condition=ready pod -l sparkoperator.k8s.io/sparkapplication=spark-pi-python --timeout=900s
kubectl -n $NAMESPACE logs -l sparkoperator.k8s.io/sparkapplication=spark-pi-python, sparkoperator.k8s.io/node-type=head

# Forward the port of Spark UI
sleep 5
kubectl -n $NAMESPACE port-forward --address 0.0.0.0 svc/spark-pi-python-head-svc 4040 :4040  &
PID=$!
echo "Forward the port 4040  of Spark head in the background process: $PID"

# Send a curl command to test basic Spark functionality.
sleep 5
output=$(curl -H "Content-Type: application/json" localhost:4040/api/version)
echo "output: ${output}"

# output format: {"version": ..., "spark_version": SPARK_VERSION, "spark_commit": ...}
if echo "${output}" | grep -q $SPARK_VERSION; then
  echo "Test succeeded!"
else
  echo "Test failed!"
  exit 1
fi

# Delete Spark Deployment
kubectl -n $NAMESPACE delete -f sparkapplication_example.yaml

# Wait for all Ray Pods to be deleted.
start_time=$(date +%s)
for ((i=0; i<TIMEOUT; i+=SLEEP_INTERVAL)); do
  pods=$(kubectl -n $NAMESPACE get pods -o json | jq '.items | length')
  if [ "$pods" -eq 0 ]; then
    kill $PID
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

# Delete Spark operator
kustomize build spark-operator/base | kubectl -n kubeflow delete -f -
