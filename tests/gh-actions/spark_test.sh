#!/bin/bash
set -euxo

NAMESPACE=$1
kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
kubectl get namespaces --selector=istio-injection=enabled
kubectl -n $NAMESPACE apply -f sparkapplication_example.yaml

# Wait for the Spark aapplication is on the custer.
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
kubectl -n $NAMESPACE delete -f sparkapplication_example.yaml