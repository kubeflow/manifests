#!/bin/bash
set -euxo pipefail
KF_PROFILE=${1:-kubeflow-user-example-com}

kubectl get crd jobsets.jobset.x-k8s.io
kubectl get service jobset-webhook-service -n kubeflow-system
kubectl get mutatingwebhookconfiguration jobset-mutating-webhook-configuration
kubectl get validatingwebhookconfiguration jobset-validating-webhook-configuration


kubectl wait --for=condition=Available deployment/jobset-controller-manager -n kubeflow-system --timeout=120s
kubectl wait --for=condition=Ready pod -l control-plane=controller-manager -n kubeflow-system --timeout=60s

sleep 10
# Verify webhook certs are ready before creating resources
kubectl wait --timeout=120s --for='jsonpath={.webhooks[0].clientConfig.caBundle}' mutatingwebhookconfiguration/jobset-mutating-webhook-configuration
kubectl wait --timeout=120s --for='jsonpath={.webhooks[0].clientConfig.caBundle}' validatingwebhookconfiguration/validator.trainer.kubeflow.org
kubectl get endpoints jobset-webhook-service -n kubeflow-system

# Create TrainJob with retry to handle transient webhook unreachability
for i in $(seq 1 5); do
  if kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE; then
    echo "TrainJob created successfully"
    break
  fi
  if [ $i -eq 5 ]; then
    echo "Failed to create TrainJob after 5 attempts"
    exit 1
  fi
  echo "Attempt $i/5 failed, retrying in 15s..."
  sleep 15
done

sleep 30

kubectl get deployment kubeflow-trainer-controller-manager -n kubeflow-system
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer
kubectl get clustertrainingruntimes torch-distributed

kubectl get pods -n $KF_PROFILE --show-labels
kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=120s
kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s
