#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

cat tests/training_operator_job.yaml | \
sed 's/name: pytorch-simple/name: pytorch-simple\n  namespace: '"$KF_PROFILE"'/g' > /tmp/pytorch-job.yaml
kubectl apply -f /tmp/pytorch-job.yaml

# Wait for the PyTorchJob status conditions to be populated by the operator.
echo "Waiting for PyTorchJob status to be populated..."
pytorch_job_status_timeout_seconds=120
pytorch_job_status_poll_interval_seconds=2
pytorch_job_status_is_populated=false
for ((elapsed_seconds=0; elapsed_seconds<pytorch_job_status_timeout_seconds; elapsed_seconds+=pytorch_job_status_poll_interval_seconds)); do
    pytorch_job_condition_type=$(kubectl get pytorchjob pytorch-simple -n "$KF_PROFILE" -o jsonpath='{.status.conditions[0].type}' 2>/dev/null || true)
    if [[ -n "$pytorch_job_condition_type" ]]; then
        pytorch_job_status_is_populated=true
        break
    fi
    sleep "$pytorch_job_status_poll_interval_seconds"
done
if [[ "$pytorch_job_status_is_populated" != "true" ]]; then
    echo "ERROR: Timeout waiting for PyTorchJob status. Collecting diagnostics..."
    kubectl describe pytorchjob pytorch-simple -n "$KF_PROFILE"
    kubectl get pods -n "$KF_PROFILE" -l training.kubeflow.org/job-name=pytorch-simple
    kubectl get events -n "$KF_PROFILE" --sort-by=.metadata.creationTimestamp
    exit 1
fi

echo "PyTorchJob created successfully. Waiting for pods..."
kubectl get pods -n $KF_PROFILE --show-labels

kubectl wait --for=condition=Ready pod -l training.kubeflow.org/replica-type=master -n $KF_PROFILE --timeout=240s

kubectl wait --for=condition=Ready pod -l training.kubeflow.org/replica-type=worker -n $KF_PROFILE --timeout=240s

echo "Checking PyTorchJob status..."
kubectl get pytorchjob pytorch-simple -n $KF_PROFILE -o yaml

echo "Checking pod logs for debugging..."
kubectl logs -l training.kubeflow.org/replica-type=master -n $KF_PROFILE --tail=50 || echo "Master logs not available yet"
kubectl logs -l training.kubeflow.org/replica-type=worker -n $KF_PROFILE --tail=50 || echo "Worker logs not available yet"

kubectl wait --for=condition=Succeeded pytorchjob/pytorch-simple -n $KF_PROFILE --timeout=300s
