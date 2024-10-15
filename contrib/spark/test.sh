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

# Wait for the operator controller to be ready.
kubectl -n kubeflow wait --for=condition=available --timeout=60s deploy/spark-operator-controller
kubectl -n kubeflow get pod -l app.kubernetes.io/name=spark-operator

# Wait for the operator webhook to be ready.
kubectl -n kubeflow wait --for=condition=available --timeout=30s deploy/spark-operator-webhook
kubectl -n kubeflow get pod -l app.kubernetes.io/name=spark-operator

# Install Spark components
kubectl -n $NAMESPACE apply -f sparkapplication_example.yaml

# Wait for the Spark to be ready.
sleep 5
# Wait until the SparkApplication reaches the "RUNNING" state
while true; do
    STATUS=$(kubectl get sparkapplication spark-pi-python -n $NAMESPACE -o jsonpath='{.status.applicationState.state}')
    
    if [ "$STATUS" == "RUNNING" ]; then
        echo "SparkApplication 'spark-pi-python' is running."

        break
    else
        echo "Waiting for SparkApplication to be in RUNNING state. Current state: $STATUS"
        sleep 5  # Check every 5 seconds
    fi
done

# Wait for the Spark to be ready.
sleep 5
# Wait until the Spark driver pod reaches the "Succeeded" or "Failed" phase
while true; do
    POD_STATUS=$(kubectl get pod spark-pi-python-driver -n $NAMESPACE -o jsonpath='{.status.phase}')
    
    if [ "$POD_STATUS" == "Succeeded" ] || [ "$POD_STATUS" == "Failed" ]; then
        echo "Driver pod has completed with status: $POD_STATUS"
        break
    else
        echo "Waiting for driver pod to complete. Current status: $POD_STATUS"
        sleep 5  # Check every 5 seconds
    fi
done

kubectl -n $NAMESPACE logs pod/spark-pi-python-driver

# Delete Spark Deployment
kubectl -n $NAMESPACE delete -f sparkapplication_example.yaml

# Delete Spark operator
kustomize build spark-operator/base | kubectl -n kubeflow delete -f -
