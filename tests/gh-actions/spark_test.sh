#!/bin/bash
set -euxo

NAMESPACE=$1

# Make sure we're in the repository root directory
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "${GITHUB_WORKSPACE:-$(pwd)}")

# Define the path to the spark application example YAML
SPARK_APP_YAML="${REPO_ROOT}/apps/spark/sparkapplication_example.yaml"

# Verify the file exists
if [ ! -f "$SPARK_APP_YAML" ]; then
    echo "Error: Spark application YAML not found at $SPARK_APP_YAML"
    exit 1
fi

kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
kubectl get namespaces --selector=istio-injection=enabled
kubectl -n $NAMESPACE apply -f "$SPARK_APP_YAML"

# Wait for the Spark application is on the cluster.
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

# Wait for Spark to be ready.
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
kubectl -n $NAMESPACE delete -f "$SPARK_APP_YAML"