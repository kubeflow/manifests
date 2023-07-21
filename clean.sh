
cat <<EOF | kubectl delete -f -
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


cat <<EOF | kubectl delete -f -
apiVersion: kubeflow.org/v1alpha1
kind: PodDefault
metadata:
  name: access-ml-pipeline
  namespace: "kubeflow-thanhnm777-gmail-com"
spec:
  desc: Allow access to Kubeflow Pipelines
  selector:
    matchLabels:
      access-ml-pipeline: "true"
  volumes:
    - name: volume-kf-pipeline-token
      projected:
        sources:
          - serviceAccountToken:
              path: token
              expirationSeconds: 99999
              audience: pipelines.kubeflow.org      
  volumeMounts:
    - mountPath: /var/run/secrets/kubeflow/pipelines
      name: volume-kf-pipeline-token
      readOnly: true
  env:
    - name: KF_PIPELINES_SA_TOKEN_PATH
      value: /var/run/secrets/kubeflow/pipelines/token
EOF

kubectl delete -f dex-configmap.yaml

kubectl delete ns thanhnm777-gmail-com
istioVersion="1.18.1"
kubectl delete -f /tmp/istio-${istioVersion}/samples/addons
kubectl delete -f kubeflow.yaml
