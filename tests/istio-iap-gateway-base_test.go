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

func writeIapGatewayBase(th *KustTestHarness) {
	th.writeF("/manifests/istio/iap-gateway/base/istio-ingressgateway.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: istio-ingressgateway
  annotations:
    beta.cloud.google.com/backend-config: '{"ports": {"http2":"iap-backendconfig"}}'
  labels:
    app: istio-ingressgateway
    istio: ingressgateway
    release: istio
spec:
  ports:
    - name: status-port
      port: 15020
      protocol: TCP
      targetPort: 15020
    - name: http2
      port: 80
      protocol: TCP
      targetPort: 80
    - name: https
      port: 443
      protocol: TCP
      targetPort: 443
    - name: kiali
      port: 15029
      protocol: TCP
      targetPort: 15029
    - name: prometheus
      port: 15030
      protocol: TCP
      targetPort: 15030
    - name: grafana
      port: 15031
      protocol: TCP
      targetPort: 15031
    - name: tracing
      port: 15032
      protocol: TCP
      targetPort: 15032
    - name: tls
      port: 15443
      protocol: TCP
      targetPort: 15443
  selector:
    app: istio-ingressgateway
    istio: ingressgateway
  sessionAffinity: None
  type: NodePort
---
apiVersion: rbac.istio.io/v1alpha1
kind: ClusterRbacConfig
metadata:
  name: default
spec:
  mode: ON_WITH_EXCLUSION
  exclusion:
    namespaces:
    - istio-system
`)
	th.writeK("/manifests/istio/iap-gateway/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - istio-ingressgateway.yaml
namespace: istio-system
`)
}

func TestIapGatewayBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/istio/iap-gateway/base")
	writeIapGatewayBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../istio/iap-gateway/base"
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
