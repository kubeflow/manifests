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
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kfserving-ingressgateway
  labels:
    app: kfserving-ingressgateway
    kfserving: ingressgateway
spec:
  replicas: 1
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
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-ingressgateway-service-account
      containers:
        - name: istio-proxy
          image: "docker.io/istio/proxyv2:1.0.2"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
            - containerPort: 443
            - containerPort: 31400
            - containerPort: 15011
            - containerPort: 8060
            - containerPort: 853
            - containerPort: 15030
            - containerPort: 15031
          args:
            - proxy
            - router
            - -v
            - "2"
            - --discoveryRefreshDelay
            - "1s" #discoveryRefreshDelay
            - --drainDuration
            - "45s" #drainDuration
            - --parentShutdownDuration
            - "1m0s" #parentShutdownDuration
            - --connectTimeout
            - "10s" #connectTimeout
            - --serviceCluster
            - kfserving-ingressgateway
            - --zipkinAddress
            - zipkin:9411
            - --statsdUdpAddress
            - istio-statsd-prom-bridge:9125
            - --proxyAdminPort
            - "15000"
            - --controlPlaneAuthPolicy
            - NONE
            - --discoveryAddress
            - istio-pilot:8080
          resources:
            requests:
              cpu: 10m
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
            - name: ISTIO_META_POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
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
    - name: http2-prometheus
      port: 15030
      targetPort: 15030
    - name: http2-grafana
      port: 15031
      targetPort: 15031
`)
	th.writeK("/manifests/istio/kfserving-gateway/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: istio-system
resources:
- deployment.yaml
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
