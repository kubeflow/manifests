#!/usr/bin/env bash
# Helm counterpart of the "Install Kubeflow Istio Resources" step and of the
# two Istio applies inside tests/knative_serving_install.sh, proven equivalent
# to common/istio/kubeflow-istio-resources/base and
# common/istio/cluster-local-gateway/overlays/m2m-auth by the istio
# platform-full comparison scenario.
#
# Third revision of the istio release created by tests/istio_helm_install.sh:
# the platform-full values keep the oauth2-proxy profile and add the
# cluster-local gateway with machine-to-machine authentication and the
# Kubeflow Istio resources. The gateway's RequestAuthentication points at the
# cluster JWKS proxy, so this runs after tests/oauth2-proxy_helm_install.sh.
set -euxo pipefail
echo "Installing Kubeflow Istio resources and the cluster-local gateway with Helm ..."
helm upgrade istio common/istio/helm \
  --namespace istio-system \
  --values common/istio/helm/ci/values-platform-full.yaml \
  --set namespaces.create=false \
  --wait --timeout 5m
kubectl rollout status deployment/cluster-local-gateway -n istio-system --timeout=180s
