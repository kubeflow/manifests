#!/bin/bash
set -e

python3 -m venv /tmp/dex-test-venv
source /tmp/dex-test-venv/bin/activate
pip3 install -q requests passlib

if ! kubectl get namespace auth &>/dev/null; then
  kubectl create namespace auth
  
  kustomize build ./common/dex/overlays/oauth2-proxy | kubectl apply -f - || {
    kustomize build ./common/dex/overlays/oauth2-proxy | grep -v "kind: VirtualService" | kubectl apply -f -
  }
  sleep 5
fi

if ! kubectl get deployment -n auth dex &>/dev/null; then
  kustomize build ./common/dex/overlays/oauth2-proxy | kubectl apply -f -
  sleep 5
fi

if kubectl get pod -l app=dex -n auth 2>/dev/null | grep -q "No resources found"; then
  kustomize build ./common/dex/overlays/oauth2-proxy | kubectl apply -f -
  sleep 5
fi

kubectl wait --for=condition=Ready pod -l app=dex -n auth --timeout=180s

if ! kubectl get secret -n auth dex-secret > /dev/null 2>&1; then
  kubectl create secret generic dex-secret -n auth --from-literal=DEX_USER_PASSWORD=$(python3 -c 'from passlib.hash import bcrypt; print(bcrypt.using(rounds=12, ident="2y").hash("12341234"))')
  if kubectl get deployment -n auth dex &>/dev/null; then
    kubectl rollout restart deployment -n auth dex
    kubectl wait --for=condition=Available deployment -n auth dex --timeout=180s
  fi
fi

if kubectl get cm -n auth dex -o yaml | grep -q "redirectURIs.*http://authservice"; then
  kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex
  namespace: auth
data:
  config.yaml: |
    issuer: http://dex.auth.svc.cluster.local:5556/dex
    storage:
      type: kubernetes
      config:
        inCluster: true
    web:
      http: 0.0.0.0:5556
    logger:
      level: "debug"
      format: text
    oauth2:
      skipApprovalScreen: true
    enablePasswordDB: true
    staticPasswords:
    - email: user@example.com
      hashFromEnv: DEX_USER_PASSWORD
      username: user
      userID: "15841185641784"
    staticClients:
    - idEnv: OIDC_CLIENT_ID
      redirectURIs: ["/oauth2/callback"]
      name: 'Dex Login Application'
      secretEnv: OIDC_CLIENT_SECRET
EOF
  kubectl rollout restart deployment -n auth dex
  sleep 5
fi

kubectl create secret generic oidc-client-secret -n auth \
  --from-literal=OIDC_CLIENT_ID=kubeflow-oidc-authservice \
  --from-literal=OIDC_CLIENT_SECRET=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic oidc-client-secret -n oauth2-proxy \
  --from-literal=OIDC_CLIENT_ID=kubeflow-oidc-authservice \
  --from-literal=OIDC_CLIENT_SECRET=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok \
  --dry-run=client -o yaml | kubectl apply -f -

if ! kubectl get deploy -n auth dex -o yaml | grep -q "OIDC_CLIENT_ID"; then
  kubectl create secret generic dex-secret -n auth \
    --from-literal=DEX_USER_PASSWORD=$(python3 -c 'from passlib.hash import bcrypt; print(bcrypt.using(rounds=12, ident="2y").hash("12341234"))') \
    --dry-run=client -o yaml | kubectl apply -f -
    
  kubectl create secret generic oidc-client-secret -n auth \
    --from-literal=OIDC_CLIENT_ID=kubeflow-oidc-authservice \
    --from-literal=OIDC_CLIENT_SECRET=pUBnBOY80SnXgjibTYM9ZWNzY2xreNGQok \
    --dry-run=client -o yaml | kubectl apply -f -

  kustomize build ./common/dex/overlays/oauth2-proxy | kubectl apply -f -
  
  kubectl rollout restart deployment -n auth dex
  kubectl rollout status deployment -n auth dex --timeout=180s
fi

if kubectl get deployment -n oauth2-proxy oauth2-proxy &>/dev/null; then
  kubectl rollout restart deployment -n oauth2-proxy oauth2-proxy
fi

RETRY_COUNT=0
MAX_RETRIES=3
until curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/dex/health 2>/dev/null | grep -q "200\|302\|404"; do
  RETRY_COUNT=$((RETRY_COUNT+1))
  if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "Error: Dex health endpoint not available after $MAX_RETRIES attempts"
    exit 1
  fi
  sleep 10
done

sed -i 's/raise RuntimeError/print("ERROR:"); exit 1/g' tests/gh-actions/test_dex_login.py

python3 - <<'EOF'
import re
with open('tests/gh-actions/test_dex_login.py', 'r') as f:
    content = f.read()
content = re.sub('import re', 'import re, time, sys', content, count=1)
retry_pattern = r'([ \t]+)session_cookies = dex_session_manager\.get_session_cookies\(\)'
replacement = r"""\1# Try with retries
\1for _attempt in range(3):
\1    session_cookies = dex_session_manager.get_session_cookies()
\1    if session_cookies:
\1        break
\1    if _attempt == 2:  # Last attempt failed
\1        print("Error: Failed to get Dex session cookies after 3 attempts")
\1        sys.exit(1)
\1    time.sleep(5)"""
content = re.sub(retry_pattern, replacement, content, count=1)
with open('tests/gh-actions/test_dex_login.py', 'w') as f:
    f.write(content)
EOF

python3 tests/gh-actions/test_dex_login.py 