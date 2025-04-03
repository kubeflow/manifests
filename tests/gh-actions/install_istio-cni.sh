#!/bin/bash
set -e
echo "Installing Istio-cni (with ExtAuthZ from oauth2-proxy) ..."
cd common/istio-cni-1-24
kustomize build istio-crds/base | kubectl apply -f -
kustomize build istio-namespace/base | kubectl apply -f -
kustomize build istio-install/overlays/oauth2-proxy | kubectl apply -f -

echo "Waiting for all Istio Pods to become ready..."
kubectl wait --for=condition=Ready pods --all -n istio-system --timeout 300s

# Ensure the gateway is properly applied
echo "Ensuring Istio gateway is properly configured..."
kubectl apply -f istio-install/base/gateway.yaml

# Verify gateway exists
echo "Verifying gateway service..."
if ! kubectl get svc -n istio-system -l app=istio-ingressgateway | grep -q "istio-ingressgateway"; then
  echo "Gateway service not found with app=istio-ingressgateway label, checking with other labels..."
  
  # Check with alternative labels
  if ! kubectl get svc -n istio-system -l istio=ingressgateway | grep -q "ingressgateway"; then
    echo "Warning: Gateway service not found with expected labels."
    
    # List all services in istio-system
    echo "Available services in istio-system namespace:"
    kubectl get svc -n istio-system
    
    echo "Adding standard ingressgateway service if missing..."
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: istio-ingressgateway
  namespace: istio-system
  labels:
    app: istio-ingressgateway
    istio: ingressgateway
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: istio-ingressgateway
    istio: ingressgateway
EOF
  fi
fi

# Print the final ingressgateway service state for debugging
echo "Final Istio ingress gateway service state:"
kubectl get svc -n istio-system -l app=istio-ingressgateway -o wide
