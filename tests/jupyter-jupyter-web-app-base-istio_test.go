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

func writeJupyterWebAppBaseIstio(th *KustTestHarness) {
	th.writeF("/manifests/jupyter/jupyter-web-app/base/istio/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: jupyter-web-app
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - headers:
      request:
        add:
          x-forwarded-prefix: /jupyter
    match:
    - uri:
        prefix: /jupyter/
    rewrite:
      uri: /
    route:
    - destination:
        host: jupyter-web-app-service.$(namespace).svc.$(clusterDomain)
        port:
          number: 80
`)
	th.writeF("/manifests/jupyter/jupyter-web-app/base/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeK("/manifests/jupyter/jupyter-web-app/base/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- virtual-service.yaml
configurations:
- params.yaml
`)
	th.writeK("/manifests/jupyter/jupyter-web-app/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- application
- core
- istio
commonLabels:
  app.kubernetes.io/name: jupyter-web-app
  app.kubernetes.io/instance: jupyter-web-app-v0.7.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: jupyter-web-app
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.7.0
`)
}

func TestJupyterWebAppBaseIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/jupyter/jupyter-web-app/base/istio")
	writeJupyterWebAppBaseIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../jupyter/jupyter-web-app/base/istio"
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
