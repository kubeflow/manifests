#!/bin/bash
# Script to setup external access for KServe testing

set -euo pipefail

NAMESPACE=${1:-kubeflow-user-example-com}
SERVICE_NAME=${2:-test-sklearn}

echo "Setting up external access for KServe..."
echo "Namespace: $NAMESPACE"
echo "Service: $SERVICE_NAME"
echo

# Create VirtualService for external access
cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: ${SERVICE_NAME}-external-access
  namespace: $NAMESPACE
spec:
  gateways:
    - kubeflow/kubeflow-gateway
  hosts:
    - '*'
  http:
    - match:
        - uri:
            prefix: /kserve/$NAMESPACE/$SERVICE_NAME/
      rewrite:
        uri: /
      route:
        - destination:
            host: cluster-local-gateway.istio-system.svc.cluster.local
          headers:
            request:
              set:
                Host: ${SERVICE_NAME}-predictor.${NAMESPACE}.svc.cluster.local
          weight: 100
      timeout: 300s
      headers:
        response:
          add:
            Access-Control-Allow-Origin: "*"
            Access-Control-Allow-Methods: "GET, POST, OPTIONS"
            Access-Control-Allow-Headers: "Authorization, Content-Type"
EOF

echo "External access configured for $SERVICE_NAME in $NAMESPACE"
echo
echo "Usage examples:"
echo "External access URL: http://YOUR_CLUSTER_IP/kserve/$NAMESPACE/$SERVICE_NAME/v1/models/$SERVICE_NAME:predict"
echo
echo "Test command:"
echo "curl -H \"Authorization: Bearer \$(kubectl -n $NAMESPACE create token default-editor)\" \\"
echo "     -H \"Content-Type: application/json\" \\"
echo "     \"http://localhost:8080/kserve/$NAMESPACE/$SERVICE_NAME/v1/models/$SERVICE_NAME:predict\" \\"
echo "     -d '{\"instances\": [[6.8, 2.8, 4.8, 1.4]]}'"