#!/bin/bash
set -euo pipefail

kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer
kubectl get pods -n kubeflow-system -l control-plane=controller-manager

kubectl get svc -n kubeflow-system kubeflow-trainer-controller-manager
kubectl get svc -n kubeflow-system jobset-webhook-service

kubectl get endpoints -n kubeflow-system kubeflow-trainer-controller-manager
kubectl get endpoints -n kubeflow-system jobset-webhook-service

kubectl get validatingwebhookconfiguration validator.trainer.kubeflow.org
kubectl get mutatingwebhookconfiguration jobset-mutating-webhook-configuration
kubectl get validatingwebhookconfiguration jobset-validating-webhook-configuration

kubectl apply --dry-run=server -f tests/trainer_job.yaml -n default
