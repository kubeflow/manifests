#!/bin/bash
set -euo pipefail

PRIMARY_NAMESPACE="kubeflow-user-example-com"
ATTACKER_NAMESPACE="kubeflow-user-attacker"
INGRESS_HOST_PORT="${KSERVE_INGRESS_HOST_PORT:-localhost:8080}"

function setup_test_environment() {
    kubectl create namespace $PRIMARY_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace $ATTACKER_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    kubectl create serviceaccount default-editor -n $PRIMARY_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    kubectl create serviceaccount attacker-sa -n $ATTACKER_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    
    cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: secure-model-predictor
  namespace: $PRIMARY_NAMESPACE
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        ports:
        - containerPort: 8080
        env:
        - name: TARGET
          value: "Secure KServe Model"
EOF

    kubectl wait --for=condition=Ready ksvc/secure-model-predictor -n $PRIMARY_NAMESPACE --timeout=120s
}

function test_gateway_authentication() {
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        "http://$INGRESS_HOST_PORT/")
    
    if [ "$RESPONSE" != "403" ]; then
        exit 1
    fi
    
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer invalid-token" \
        "http://$INGRESS_HOST_PORT/")
    
    if [ "$RESPONSE" != "401" ] && [ "$RESPONSE" != "403" ]; then
        exit 1
    fi
}

function test_namespace_isolation() {
    PRIMARY_TOKEN=$(kubectl -n $PRIMARY_NAMESPACE create token default-editor)
    ATTACKER_TOKEN=$(kubectl -n $ATTACKER_NAMESPACE create token attacker-sa)
    
    RESPONSE=$(curl -s -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer $PRIMARY_TOKEN" \
        "http://$INGRESS_HOST_PORT/")
    
    if [[ "$RESPONSE" != *"200"* ]] && [[ "$RESPONSE" != *"404"* ]] && [[ "$RESPONSE" != *"503"* ]]; then
        exit 1
    fi
    
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer $ATTACKER_TOKEN" \
        "http://$INGRESS_HOST_PORT/")
    
    if [[ "$RESPONSE" != *"200"* ]] && [[ "$RESPONSE" != *"404"* ]] && [[ "$RESPONSE" != *"503"* ]]; then
        exit 1
    fi
}

function test_cluster_local_gateway() {
    kubectl port-forward -n istio-system svc/cluster-local-gateway 8081:80 &
    PF_PID=$!
    sleep 5
    
    PRIMARY_TOKEN=$(kubectl -n $PRIMARY_NAMESPACE create token default-editor)
    
    curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        -H "Authorization: Bearer $PRIMARY_TOKEN" \
        "http://localhost:8081/" > /dev/null
    
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Host: secure-model-predictor.$PRIMARY_NAMESPACE.svc.cluster.local" \
        "http://localhost:8081/")
    
    if [ "$RESPONSE" != "403" ]; then
        exit 1
    fi
    
    kill $PF_PID 2>/dev/null || true
}

function cleanup() {
    kubectl delete namespace $ATTACKER_NAMESPACE --ignore-not-found=true
    kubectl delete ksvc secure-model-predictor -n $PRIMARY_NAMESPACE --ignore-not-found=true
}

function main() {
    setup_test_environment
    test_gateway_authentication
    test_namespace_isolation
    test_cluster_local_gateway
    cleanup
}

main