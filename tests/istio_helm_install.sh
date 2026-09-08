#!/usr/bin/env bash
# Helm counterpart of tests/istio-cni_install.sh, proven equivalent to
# common/istio/istio-crds/base, istio-namespace/base and
# istio-install/overlays/oauth2-proxy by the istio crds and oauth2-proxy
# comparison scenarios.
#
# Two release revisions, as the chart README documents: the first renders only
# the CustomResourceDefinitions, the second the Istio installation, because
# Istio custom resources cannot be created before their definitions are
# established. The comparison scenario values enable namespaces.create so the
# Namespace is compared; here the kubeflow-namespaces chart already owns
# Namespace/istio-system, so the flag is turned off on the command line.
set -euxo pipefail
echo "Installing Istio (with ExtAuthZ from oauth2-proxy) with Helm ..."
helm install istio common/istio/helm \
  --namespace istio-system \
  --values common/istio/helm/ci/values-crds.yaml \
  --wait --timeout 5m

mapfile -t CUSTOM_RESOURCE_DEFINITION_NAMES < <(
  helm get manifest istio --namespace istio-system |
    awk '
      $0 == "kind: CustomResourceDefinition" {
        custom_resource_definition = 1
        next
      }
      custom_resource_definition && /^  name: / {
        print $2
        custom_resource_definition = 0
      }
    '
)
if [[ "${#CUSTOM_RESOURCE_DEFINITION_NAMES[@]}" -eq 0 ]]; then
  echo "No CustomResourceDefinition resources were found in the istio Helm release." >&2
  exit 1
fi
for name in "${CUSTOM_RESOURCE_DEFINITION_NAMES[@]}"; do
  kubectl wait --for=condition=Established "crd/${name}" --timeout=120s
done

helm upgrade istio common/istio/helm \
  --namespace istio-system \
  --values common/istio/helm/ci/values-oauth2-proxy.yaml \
  --set namespaces.create=false \
  --wait --timeout 5m

echo "Waiting for all Istio Pods to become ready..."
kubectl rollout status deployment/istiod -n istio-system --timeout=180s
kubectl rollout status deployment/istio-ingressgateway -n istio-system --timeout=180s
kubectl rollout status daemonset/istio-cni-node -n kube-system --timeout=180s
kubectl wait --for=condition=Ready pods --all -n istio-system --timeout 180s
kubectl wait --for=condition=Ready pods --all --all-namespaces --timeout=180s '--field-selector=status.phase!=Succeeded'
