#!/bin/bash
set -euxo pipefail

KF_PROFILE=${1:-kubeflow-user-example-com}

for i in {1..30}; do
  if kubectl api-resources | grep -q "pytorchjobs.*kubeflow"; then
    echo "PyTorch job CRD is available"
    break
  fi
  echo "Waiting for PyTorch job CRD ($i/30)..."
  sleep 5
done

echo "Creating PyTorch job in namespace $KF_PROFILE"
JOB_YAML=$(sed 's/namespace: .*/namespace: '"$KF_PROFILE"'/g' tests/gh-actions/kf-objects/training_operator_job.yaml)
echo "$JOB_YAML" | kubectl apply -f -

echo "Verifying PyTorch job creation..."
for i in {1..20}; do
  if kubectl get pytorchjob/pytorch-simple -n $KF_PROFILE -o name --request-timeout=5s 2>/dev/null; then
    echo "PyTorch job exists!"
    break
  fi
  echo "Waiting for PyTorch job to be found ($i/20)..."
  sleep 5
  
  if [ $i -eq 5 ]; then
    echo "Recreating PyTorch job..."
    kubectl delete pytorchjob/pytorch-simple -n $KF_PROFILE --ignore-not-found=true
    echo "$JOB_YAML" | kubectl apply -f -
  fi
done

echo "Checking training operator logs..."
kubectl logs -n kubeflow -l app=training-operator --tail=20

# Polling instead of kubectl wait
echo "Polling for job success..."
TIMEOUT=600
INTERVAL=10
ELAPSED=0

while [ $ELAPSED -lt $TIMEOUT ]; do
  STATUS=$(kubectl get pytorchjob/pytorch-simple -n $KF_PROFILE -o jsonpath='{.status.conditions[?(@.type=="Succeeded")].status}' --request-timeout=5s 2>/dev/null || echo "")
  
  if [ "$STATUS" == "True" ]; then
    echo "PyTorch job succeeded!"
    exit 0
  fi
  
  echo "Job not yet succeeded, current status conditions:"
  kubectl get pytorchjob/pytorch-simple -n $KF_PROFILE -o jsonpath='{.status.conditions}' --request-timeout=5s 2>/dev/null || echo "Job not found"
  
  sleep $INTERVAL
  ELAPSED=$((ELAPSED + INTERVAL))
  echo "Elapsed time: $ELAPSED seconds of $TIMEOUT"
done

echo "Timed out waiting for PyTorch job to succeed."
kubectl get pytorchjobs -n $KF_PROFILE -o yaml
exit 1 