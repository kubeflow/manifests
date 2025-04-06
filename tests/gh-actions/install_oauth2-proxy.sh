#!/bin/bash
set -e

echo "Installing oauth2-proxy..."
cd common/
kustomize build oauth2-proxy/overlays/m2m-dex-and-kind/ | kubectl apply -f -

# Create a secret with client ID and secret for Dex
echo "Creating OAuth2 client secret..."
kubectl create namespace auth --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic oidc-client-secret -n auth \
  --from-literal=OIDC_CLIENT_ID=kubeflow-oidc-authservice \
  --from-literal=OIDC_CLIENT_SECRET=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok \
  --dry-run=client -o yaml | kubectl apply -f -

# Add the same secret to the oauth2-proxy namespace
kubectl create secret generic oidc-client-secret -n oauth2-proxy \
  --from-literal=OIDC_CLIENT_ID=kubeflow-oidc-authservice \
  --from-literal=OIDC_CLIENT_SECRET=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Waiting for all oauth2-proxy pods to become ready..."
kubectl wait --for=condition=Ready pod -l 'app.kubernetes.io/name=oauth2-proxy' --timeout=180s -n oauth2-proxy || echo "oauth2-proxy pods not ready, continuing anyway"

echo "Waiting for all cluster-jwks-proxy pods to become ready..."
kubectl wait --for=condition=Ready pod -l 'app.kubernetes.io/name=cluster-jwks-proxy' --timeout=180s -n istio-system || echo "cluster-jwks-proxy pods not ready, continuing anyway"

# Check if we need to restart oauth2-proxy to pick up the secret
echo "Ensuring OAuth2 proxy has proper configuration..."
kubectl rollout restart deployment -n oauth2-proxy oauth2-proxy
sleep 10
kubectl wait --for=condition=Available deployment -n oauth2-proxy oauth2-proxy --timeout=180s || echo "oauth2-proxy deployment not available, but continuing"

# Return to original directory
cd ..