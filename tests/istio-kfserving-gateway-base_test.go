package tests_test

import (
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
	"testing"
)

func writeKfservingGatewayBase(th *KustTestHarness) {
	th.writeF("/manifests/istio/kfserving-gateway/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kfserving-ingressgateway
  labels:
    app: kfserving-ingressgateway
    kfserving: ingressgateway
spec:
  selector:
    matchLabels:
      app: kfserving-ingressgateway
      kfserving: ingressgateway
  template:
    metadata:
      labels:
        app: kfserving-ingressgateway
        kfserving: ingressgateway
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-ingressgateway-service-account
      containers:
        - name: istio-proxy
          image: "docker.io/istio/proxyv2:1.1.6"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 15020
            - containerPort: 80
            - containerPort: 443
            - containerPort: 31400
            - containerPort: 15029
            - containerPort: 15030
            - containerPort: 15031
            - containerPort: 15032
            - containerPort: 15443
            - containerPort: 15090
              protocol: TCP
              name: http-envoy-prom
          args:
          - proxy
          - router
          - --domain
          - $(POD_NAMESPACE).svc.cluster.local
          - --log_output_level=default:info
          - --drainDuration
          - '45s' #drainDuration
          - --parentShutdownDuration
          - '1m0s' #parentShutdownDuration
          - --connectTimeout
          - '10s' #connectTimeout
          - --serviceCluster
          - kfserving-ingressgateway
          - --zipkinAddress
          - zipkin:9411
          - --proxyAdminPort
          - "15000"
          - --statusPort
          - "15020"
          - --controlPlaneAuthPolicy
          - NONE
          - --discoveryAddress
          - istio-pilot:15010
          readinessProbe:
            failureThreshold: 30
            httpGet:
              path: /healthz/ready
              port: 15020
              scheme: HTTP
            initialDelaySeconds: 1
            periodSeconds: 2
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            limits:
              cpu: 100m
              memory: 128Mi
            requests:
              cpu: 10m
              memory: 40Mi

          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: INSTANCE_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          - name: HOST_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.hostIP
          - name: ISTIO_META_POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: ISTIO_META_CONFIG_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: ISTIO_META_ROUTER_MODE
            value: sni-dnat
          volumeMounts:
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
          - name: ingressgateway-certs
            mountPath: "/etc/istio/ingressgateway-certs"
            readOnly: true
          - name: ingressgateway-ca-certs
            mountPath: "/etc/istio/ingressgateway-ca-certs"
            readOnly: true
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-ingressgateway-service-account
          optional: true
      - name: ingressgateway-certs
        secret:
          secretName: "istio-ingressgateway-certs"
          optional: true
      - name: ingressgateway-ca-certs
        secret:
          secretName: "istio-ingressgateway-ca-certs"
          optional: true
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x
`)
	th.writeF("/manifests/istio/kfserving-gateway/base/horizontal-pod-autoscaler.yaml", `
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  labels:
    app: kfserving-ingressgateway
    kfserving: ingressgateway
  name: kfserving-ingressgateway
spec:
  maxReplicas: 5
  metrics:
  - resource:
      name: cpu
      targetAverageUtilization: 80
    type: Resource
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: kfserving-ingressgateway
`)
	th.writeF("/manifests/istio/kfserving-gateway/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: kfserving-ingressgateway
  labels:
    app: kfserving-ingressgateway
    kfserving: ingressgateway
spec:
  type: LoadBalancer
  selector:
    app: kfserving-ingressgateway
    kfserving: ingressgateway
  ports:
    - name: status-port
      port: 15020
      targetPort: 15020
    - name: http2
      nodePort: 32380
      port: 80
      targetPort: 80
    - name: https
      nodePort: 32390
      port: 443
    - name: tcp
      nodePort: 32400
      port: 31400
    - name: tcp-pilot-grpc-tls
      port: 15011
      targetPort: 15011
    - name: tcp-citadel-grpc-tls
      port: 8060
      targetPort: 8060
    - name: tcp-dns-tls
      port: 853
      targetPort: 853
    - name: https-kiali
      port: 15029
      targetPort: 15029
    - name: http2-prometheus
      port: 15030
      targetPort: 15030
    - name: http2-grafana
      port: 15031
      targetPort: 15031
    - name: https-tracing
      port: 15032
      targetPort: 15032
    - name: tls
      port: 15443
      targetPort: 15443
`)
	th.writeK("/manifests/istio/kfserving-gateway/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: istio-system
resources:
- deployment.yaml
- horizontal-pod-autoscaler.yaml
- service.yaml
`)
}

func TestKfservingGatewayBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/istio/kfserving-gateway/base")
	writeKfservingGatewayBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../istio/kfserving-gateway/base"
	fsys := fs.MakeRealFS()
	lrc := loader.RestrictionRootOnly
	_loader, loaderErr := loader.NewLoader(lrc, validators.MakeFakeValidator(), targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pc := plugins.DefaultPluginConfig()
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl(), plugins.NewLoader(pc, rf))
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
