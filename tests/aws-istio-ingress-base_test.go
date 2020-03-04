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

func writeIstioIngressBase(th *KustTestHarness) {
	th.writeF("/manifests/aws/istio-ingress/base/ingress.yaml", `
apiVersion: extensions/v1beta1 # networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: $(loadBalancerScheme)
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}]'
  name: istio-ingress
spec:
  rules:
  - http:
      paths:
      - backend:
          serviceName: istio-ingressgateway
          servicePort: 80
        path: /*
`)
	th.writeF("/manifests/aws/istio-ingress/base/params.yaml", `
varReference:
- path: metadata/annotations
  kind: Ingress
`)
	th.writeF("/manifests/aws/istio-ingress/base/params.env", `
loadBalancerScheme=internet-facing
`)
	th.writeK("/manifests/aws/istio-ingress/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ingress.yaml
commonLabels:
  kustomize.component: istio-ingress
configMapGenerator:
- name: istio-ingress-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: loadBalancerScheme
  objref:
    kind: ConfigMap
    name: istio-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.loadBalancerScheme
configurations:
- params.yaml
`)
}

func TestIstioIngressBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/aws/istio-ingress/base")
	writeIstioIngressBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../aws/istio-ingress/base"
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
