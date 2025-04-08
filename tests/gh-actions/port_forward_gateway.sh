#!/bin/bash
set -e

echo "Setting up port forwarding for Istio ingress gateway..."

# Get the ingress gateway service name
GATEWAY_SERVICE=$(kubectl get svc -n istio-system -l app=istio-ingressgateway -o jsonpath='{.items[0].metadata.name}')

if [ -z "$GATEWAY_SERVICE" ]; then
  echo "Error: Could not find istio-ingressgateway service"
  exit 1
fi

echo "Found ingress gateway service: $GATEWAY_SERVICE"

# Start port forwarding in the background
nohup kubectl port-forward -n istio-system svc/$GATEWAY_SERVICE 8080:80 &
PORT_FORWARD_PID=$!

# Wait for port forwarding to be ready
echo "Waiting for port forwarding to be ready..."
RETRY_COUNT=0
MAX_RETRIES=10
until curl -s localhost:8080 > /dev/null || curl -s -I localhost:8080 | grep -q "HTTP/"; do
  RETRY_COUNT=$((RETRY_COUNT+1))
  if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "Port forwarding failed after $MAX_RETRIES attempts"
    exit 1
  fi
  echo "Waiting for port-forwarding... (attempt $RETRY_COUNT/$MAX_RETRIES)"
  sleep 1
done

echo "Port forwarding is ready on localhost:8080"
echo "Port forwarding process ID: $PORT_FORWARD_PID" 