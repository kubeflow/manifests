#!/bin/bash
set -euxo pipefail

curl -sL https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml > metrics-server.yaml
sed -i 's/- args:/- args:\n        - --kubelet-insecure-tls/' metrics-server.yaml
kubectl apply -f metrics-server.yaml
kubectl wait --for=condition=Available deployment/metrics-server -n kube-system --timeout=180s
kubectl wait --for=condition=Ready pods -l k8s-app=metrics-server -n kube-system --timeout=180s

echo "Waiting for metrics API to become available..."
max_retries=10
retry_interval=10
for i in $(seq 1 $max_retries); do
  if kubectl get --raw "/apis/metrics.k8s.io/v1beta1/nodes" >/dev/null 2>&1; then
    echo "Metrics API is available"
    break
  fi
  if [[ $i -eq $max_retries ]]; then
    echo "Metrics API did not become available"
    echo "Debugging metrics-server:"
    kubectl get pods -n kube-system -l k8s-app=metrics-server
    kubectl describe pods -n kube-system -l k8s-app=metrics-server
    kubectl logs -n kube-system -l k8s-app=metrics-server --tail=50
    exit 1
  fi
  echo "Metrics API not available yet, retrying..."
  sleep $retry_interval
done
