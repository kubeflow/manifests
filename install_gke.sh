kustomize build example | awk '!/well-defined/' > kubeflow.yaml

while ! kubectl apply -f kubeflow.yaml; do echo "Retrying to apply resources"; sleep 120; done
# while ! kustomize build example | awk '!/well-defined/' | kubectl apply -f -; do echo "Retrying to apply resources"; sleep 10; done


#not working but still keeptrack!
go install github.com/arttor/helmify/cmd/helmify@latest
kustomize build example | awk '!/well-defined/' | ~/go/bin/helmify kbf

TIMEOUT=600s  # 10mins


echo "---"
echo "Waiting for all Kubeflow components to become ready."

echo "Waiting for Cert Manager pods to become ready..."
kubectl wait --timeout=${TIMEOUT} -n cert-manager --all --for=condition=Ready pod

echo "Waiting for istio-system Pods to become ready..."
kubectl wait --timeout=${TIMEOUT} -n istio-system --all --for=condition=Ready pod

echo "Waiting for knative-serving Pods to become ready..."
kubectl wait --timeout=${TIMEOUT} -n knative-serving --all --for=condition=Ready pod

echo "Waiting for kubeflow/ml-pipelines to become ready..."
kubectl wait --timeout=${TIMEOUT} -n kubeflow -l app=ml-pipeline --for=condition=Ready pod

echo "Waiting for kubeflow/kfserving to become ready..."
kubectl wait --timeout=${TIMEOUT} -n kubeflow -l app=kfserving --for=condition=Ready pod

echo "Waiting for kubeflow/katib to become ready..."
kubectl wait --timeout=${TIMEOUT} -n kubeflow -l katib.kubeflow.org/component=controller --for=condition=Ready pod

echo "Waiting for kubeflow/training-operator to become ready..."
kubectl wait --timeout=${TIMEOUT} -n kubeflow -l control-plane=kubeflow-training-operator --for=condition=Ready pod


cat <<EOF | kubectl apply -f -
---
  apiVersion: networking.gke.io/v1
  kind: ManagedCertificate
  metadata:
    name: managed-cert-kbf
    namespace: istio-system
  spec:
    domains:
      - kbf.rdenginno.info
---
  apiVersion: cloud.google.com/v1
  kind: BackendConfig
  metadata:
    name: istio-ingressgateway-lsj4h-healthcheck
    namespace: "istio-system"
  spec:
    healthCheck:
      type: HTTP
      port: 15021
      requestPath: /healthz/ready
      # checkIntervalSec: 5
      # timeoutSec: 5
      # unhealthyThreshold: 3
      # healthyThreshold: 1
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    cloud.google.com/backend-config: '{"ports": {"8080":"istio-ingressgateway-lsj4h-healthcheck"}}'
    cloud.google.com/neg: '{"ingress": true}'
  name: istio-ingressgateway-lsj4h
  namespace: istio-system
  # labels:
  #   app: istio-ingressgateway
  #   install.operator.istio.io/owning-resource: unknown
  #   istio: ingressgateway
  #   istio.io/rev: default
  #   operator.istio.io/component: IngressGateways
  #   release: istio
spec:
  ports:
  - port: 8080
    protocol: TCP
    targetPort: 8080
    name: http
  selector:
    app: istio-ingressgateway
    istio: ingressgateway
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    # ingress.kubernetes.io/forwarding-rule: k8s2-fr-fqjykvvn-istio-system-kbflow-esvnmlh0
    # ingress.kubernetes.io/https-forwarding-rule: k8s2-fs-fqjykvvn-istio-system-kbflow-esvnmlh0
    # ingress.kubernetes.io/https-target-proxy: k8s2-ts-fqjykvvn-istio-system-kbflow-esvnmlh0
    # ingress.kubernetes.io/target-proxy: k8s2-tp-fqjykvvn-istio-system-kbflow-esvnmlh0
    # ingress.kubernetes.io/url-map: k8s2-um-fqjykvvn-istio-system-kbflow-esvnmlh0
    networking.gke.io/managed-certificates: managed-cert-kbf
    ingress.kubernetes.io/static-ip: kbf
    # kubernetes.io/ingress.allow-http: "false"
    kubernetes.io/ingress.class: "gce"
  name: kbflow
  namespace: istio-system
spec:
  defaultBackend:
    service:
      name: istio-ingressgateway-lsj4h
      port:
        number: 8080
---
EOF

bash install_mlflow.sh
kubectl apply -f dex-configmap.yaml
kubectl delete pod `kubectl get pods -n auth | awk '{print $1}' | grep -iv name` -n auth 


istioVersion="1.18.2"
curl -L https://istio.io/downloadIstio | sh -
mv istio-${istioVersion} /tmp/
kubectl apply -f /tmp/istio-${istioVersion}/samples/addons
kubectl rollout status deployment/kiali -n istio-system
# /tmp/istio-1.18.1/bin/istioctl dashboard kiali
cd ..


#kubectl port-forward -n kubeflow svc/minio-service 9000:9000
#kubectl port-forward -n kubeflow svc/mlflowserver 5000:5000