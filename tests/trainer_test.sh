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
kubectl get endpoints jobset-webhook-service -n kubeflow-system

kubectl apply -f tests/trainer_job.yaml -n $KF_PROFILE
sleep 15

kubectl get deployment kubeflow-trainer-controller-manager -n kubeflow-system
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer
kubectl get clustertrainingruntimes torch-distributed

kubectl get pods -n $KF_PROFILE --show-labels
kubectl wait --for=condition=Ready pod -l batch.kubernetes.io/job-name=pytorch-simple-node-0 -n $KF_PROFILE --timeout=60s
kubectl wait --for=condition=Complete job/pytorch-simple-node-0 -n $KF_PROFILE --timeout=300s
