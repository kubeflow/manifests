#!/bin/bash
set -euo pipefail
echo "Installing Pipelines ..."
cd apps/pipeline/upstream
kubectl apply -f third-party/metacontroller/base/crd.yaml
echo "Waiting for crd/compositecontrollers.metacontroller.k8s.io to be available ..."
kubectl wait --for condition=established --timeout=30s crd/compositecontrollers.metacontroller.k8s.io
kustomize build env/cert-manager/platform-agnostic-multi-user | kubectl apply -f -
sleep 60
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=600s \
  --field-selector=status.phase!=Succeeded

# Verify Pipeline components
echo "Verifying Pipeline components..."
kubectl wait --for=condition=Available deployment/ml-pipeline -n kubeflow --timeout=300s || echo "ML Pipeline API server not yet available"
kubectl wait --for=condition=Available deployment/ml-pipeline-ui -n kubeflow --timeout=300s || echo "ML Pipeline UI not yet available"
kubectl wait --for=condition=Available deployment/ml-pipeline-persistenceagent -n kubeflow --timeout=300s || echo "ML Pipeline persistence agent not yet available"
kubectl wait --for=condition=Available deployment/ml-pipeline-scheduledworkflow -n kubeflow --timeout=300s || echo "ML Pipeline scheduled workflow not yet available"
kubectl wait --for=condition=Available deployment/ml-pipeline-viewer-crd -n kubeflow --timeout=300s || echo "ML Pipeline viewer CRD not yet available"
kubectl wait --for=condition=Available deployment/cache-server -n kubeflow --timeout=300s || echo "ML Pipeline cache server not yet available"
kubectl wait --for=condition=Available deployment/metadata-writer -n kubeflow --timeout=300s || echo "ML Pipeline metadata writer not yet available"

# Verify MinIO and MySQL
echo "Verifying Pipeline storage components..."
kubectl wait --for=condition=Available deployment/minio -n kubeflow --timeout=300s || echo "MinIO not yet available"
kubectl wait --for=condition=Available deployment/mysql -n kubeflow --timeout=300s || echo "MySQL not yet available"

# Display Pipeline component status
echo "Pipeline component status:"
kubectl get deployment -n kubeflow -l app=ml-pipeline

cd -
