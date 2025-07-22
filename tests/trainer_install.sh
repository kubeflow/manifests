#!/bin/bash
set -euo pipefail

cd applications/trainer/upstream
kustomize build base/crds | kubectl apply --server-side --force-conflicts -f -

for crd in trainjobs.trainer.kubeflow.org clustertrainingruntimes.trainer.kubeflow.org trainingruntimes.trainer.kubeflow.org; do
    for i in {1..20}; do
        if kubectl get crd "$crd" >/dev/null 2>&1; then
            break
        fi
        if [ $i -eq 20 ]; then
            echo "ERROR: CRD $crd not available after 60 seconds"
            exit 1
        fi
        sleep 3
    done
done
sleep 5

cd ../../../
kustomize build common/kubeflow-system-namespace/base | kubectl apply -f -
kustomize build common/networkpolicies/kubeflow-system | kubectl apply -f - || true

cd applications/trainer/upstream
kustomize build overlays/manager | kubectl apply --server-side --force-conflicts -f -

kubectl wait --for=condition=Available deployment/kubeflow-trainer-controller-manager -n kubeflow-system --timeout=300s
kubectl wait --for=condition=Available deployment/jobset-controller-manager -n kubeflow-system --timeout=300s

kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=trainer -n kubeflow-system --timeout=180s
kubectl wait --for=condition=Ready pod -l control-plane=controller-manager -n kubeflow-system --timeout=180s

for i in {1..30}; do
    if kubectl get endpoints -n kubeflow-system jobset-webhook-service | grep -q "10\\."; then
        break
    fi
    if [ $i -eq 30 ]; then
        echo "ERROR: JobSet webhook endpoints not ready after 90 seconds"
        exit 1
    fi
    sleep 3
done

sleep 45

kubectl apply -f ../../../tests/trainer_rbac_patch.yaml || true
kustomize build overlays/runtimes | kubectl apply --server-side --force-conflicts -f -
kubectl apply -f overlays/kubeflow-platform/kubeflow-trainer-roles.yaml

cd -