#!/bin/bash
set -euxo pipefail

cd applications/trainer

kustomize build upstream/base/crds | kubectl apply --server-side --force-conflicts -f -
sleep 5
kubectl wait --for condition=established crd/trainjobs.trainer.kubeflow.org --timeout=60s

kustomize build overlays | kubectl apply --server-side --force-conflicts -f -
kubectl wait --for=condition=Available deployment/kubeflow-trainer-controller-manager -n kubeflow-system --timeout=240s
kubectl get crd jobsets.jobset.x-k8s.io
kubectl wait --for=condition=Available deployment/jobset-controller-manager -n kubeflow-system --timeout=120s

# Wait for the Trainer webhook to become fully ready.
# The controller generates TLS certs via CSR on startup, stores them in the
# kubeflow-trainer-webhook-cert Secret, and patches the webhook caBundle.
# However, the kubelet needs up to ~60-120s to sync the Secret update to the
# pod's volume mount. The controller-runtime webhook server only starts
# serving TLS once the cert files appear on disk. The readiness probe
# (port 8081) passes before this happens, so we must retry.
echo "Waiting for Trainer webhook caBundle to be populated..."
for i in $(seq 1 30); do
  CA_BUNDLE=$(kubectl get validatingwebhookconfiguration validator.trainer.kubeflow.org \
    -o jsonpath='{.webhooks[0].clientConfig.caBundle}' 2>/dev/null || true)
  if [ -n "${CA_BUNDLE}" ]; then
    echo "Trainer webhook caBundle is populated (attempt ${i}/30)"
    break
  fi
  if [ "${i}" -eq 30 ]; then
    echo "ERROR: Trainer webhook caBundle was not populated after 150s"
    kubectl get validatingwebhookconfiguration validator.trainer.kubeflow.org -o yaml 2>/dev/null || true
    exit 1
  fi
  echo "Attempt ${i}/30: caBundle not yet populated, waiting 5s..."
  sleep 5
done

# Apply runtimes with retry — the webhook server may not yet be serving TLS
# because the kubelet Secret volume mount sync can lag behind the caBundle
# patch by up to 60-120s. Server-side apply is idempotent, so retrying is safe.
echo "Applying ClusterTrainingRuntimes..."
RUNTIMES_APPLIED=false
for i in $(seq 1 10); do
  if kustomize build upstream/overlays/runtimes | kubectl apply --server-side --force-conflicts -f - 2>&1; then
    echo "ClusterTrainingRuntimes applied successfully (attempt ${i}/10)"
    RUNTIMES_APPLIED=true
    break
  fi
  echo "Attempt ${i}/10: webhook not yet serving, retrying in 15s..."
  sleep 15
done
if [ "${RUNTIMES_APPLIED}" != "true" ]; then
  echo "ERROR: Failed to apply ClusterTrainingRuntimes after 10 attempts"
  kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer -o wide 2>/dev/null || true
  kubectl logs -n kubeflow-system -l app.kubernetes.io/name=trainer --tail=50 2>/dev/null || true
  exit 1
fi

kubectl apply -f upstream/overlays/kubeflow-platform/kubeflow-trainer-roles.yaml

cd -


kubectl get deployment -n kubeflow-system kubeflow-trainer-controller-manager
kubectl get pods -n kubeflow-system -l app.kubernetes.io/name=trainer
kubectl get crd | grep -E 'trainer.kubeflow.org'
kubectl get clustertrainingruntimes

kubectl rollout restart deployment/jobset-controller-manager -n kubeflow-system
kubectl rollout status deployment/jobset-controller-manager -n kubeflow-system --timeout=120s
kubectl wait --for=condition=Available deployment/jobset-controller-manager -n kubeflow-system --timeout=120s

