#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE

kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=trainer -n kubeflow-system --timeout=60s

kubectl describe trainjob pytorch-simple -n $KF_PROFILE

kubectl logs -l app.kubernetes.io/name=trainer -n kubeflow-system --tail=20

sleep 20

kubectl logs -l app.kubernetes.io/name=trainer -n kubeflow-system --since=30s | grep -E "error|Error|ERROR|failed|Failed|FAILED" || echo "No explicit errors found"

kubectl get trainjob pytorch-simple -n $KF_PROFILE -o yaml | grep -A 10 "status:" || echo "No status field"

kubectl get jobset -n $KF_PROFILE
if ! kubectl get jobset pytorch-simple -n $KF_PROFILE >/dev/null 2>&1; then
  kubectl logs -l app.kubernetes.io/name=trainer -n kubeflow-system --since=60s
  exit 1
fi

kubectl get pods -n $KF_PROFILE --show-labels
kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=180s

kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s