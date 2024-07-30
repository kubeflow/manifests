#!/bin/bash
set -euo pipefail

ISTIO_SYSTEM="istio-system"
AUTH="auth"
CERT_MANAGER="cert-manager"
OAUTH2_PROXY="oauth2-proxy"
KUBEFLOW="kubeflow"

if kubectl get namespace "$ISTIO_SYSTEM" >/dev/null 2>&1; then
    echo "Patching the PSS-baseline labels for namespace $ISTIO_SYSTEM..."
    kubectl patch namespace $ISTIO_SYSTEM --patch-file ./contrib/security/PSS/static/baseline/patches/istio-labels.yaml
fi

if kubectl get namespace "$AUTH" >/dev/null 2>&1; then
    echo "Patching the PSS-baseline labels for namespace $AUTH..."
    kubectl patch namespace $AUTH --patch-file ./contrib/security/PSS/static/baseline/patches/dex-labels.yaml
fi

if kubectl get namespace "$CERT_MANAGER" >/dev/null 2>&1; then
    echo "Patching the PSS-baseline labels for namespace $CERT_MANAGER..."
    kubectl patch namespace $CERT_MANAGER --patch-file ./contrib/security/PSS/static/baseline/patches/cert-manager-labels.yaml
fi

if kubectl get namespace "$OAUTH2_PROXY" >/dev/null 2>&1; then
    echo "Patching the PSS-baseline labels for namespace $OAUTH2_PROXY..."
    kubectl patch namespace $OAUTH2_PROXY --patch-file ./contrib/security/PSS/static/baseline/patches/oauth2-proxy-labels.yaml
fi

if kubectl get namespace "$KUBEFLOW" >/dev/null 2>&1; then
    echo "Patching the PSS-baseline labels for namespace $KUBEFLOW..."
    kubectl patch namespace $KUBEFLOW --patch-file ./contrib/security/PSS/static/baseline/patches/kubeflow-labels.yaml
fi
