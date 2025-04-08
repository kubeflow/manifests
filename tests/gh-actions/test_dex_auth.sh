#!/bin/bash
set -e

python3 -m venv /tmp/dex-test-venv
source /tmp/dex-test-venv/bin/activate
pip3 install -q requests passlib

# Install Dex
./tests/gh-actions/install_dex.sh


# Ensure Dex deployment exists
if ! kubectl get deployment -n auth dex &>/dev/null; then
  # Use the reusable install_dex.sh script
  ./tests/gh-actions/install_dex.sh
fi

if kubectl get pod -l app=dex -n auth 2>/dev/null | grep -q "No resources found"; then
  # Use the reusable install_dex.sh script
  ./tests/gh-actions/install_dex.sh
fi

kubectl wait --for=condition=Ready pod -l app=dex -n auth --timeout=180s

if ! kubectl get secret -n auth dex-secret > /dev/null 2>&1; then
  # Create the secret using dry-run and pipe to apply to avoid errors if it already exists
  kubectl create secret generic dex-secret -n auth --from-literal=DEX_USER_PASSWORD=$(python3 -c 'from passlib.hash import bcrypt; print(bcrypt.using(rounds=12, ident="2y").hash("12341234"))') --dry-run=client -o yaml | kubectl apply -f -
  
  if kubectl get deployment -n auth dex &>/dev/null; then
    kubectl rollout restart deployment -n auth dex
    kubectl wait --for=condition=Available deployment -n auth dex --timeout=180s
  fi
fi

# Call the OAuth2 and Dex credentials setup script
chmod +x tests/gh-actions/oauth2_dex_credentials.sh
./tests/gh-actions/oauth2_dex_credentials.sh

chmod +x tests/gh-actions/test_dex_login_wrapper.sh
./tests/gh-actions/test_dex_login_wrapper.sh
