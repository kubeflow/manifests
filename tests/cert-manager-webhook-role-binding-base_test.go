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

func writeCertManagerWebhookRoleBindingBase(th *KustTestHarness) {
	th.writeF("/manifests/cert-manager/cert-manager-webhook-role-binding/base/role-binding.yaml", `
---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: cert-manager-webhook:webhook-authentication-reader
  labels:
    app: webhook
    app.kubernetes.io/name: webhook
    app.kubernetes.io/instance:  cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: extension-apiserver-authentication-reader
subjects:
- apiGroup: ""
  kind: ServiceAccount
  name: cert-manager-webhook
  namespace: $(certManagerNamespace)
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook-role-binding/base/params.yaml", `
varReference:
- path: subjects/namespace
  kind: RoleBinding
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook-role-binding/base/params.env", `
certManagerNamespace=cert-manager
`)
	th.writeK("/manifests/cert-manager/cert-manager-webhook-role-binding/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kube-system
resources:
- role-binding.yaml
commonLabels:
  kustomize.component: cert-manager
configMapGenerator:
- name: cert-manager-webhook-role-binding-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: certManagerNamespace
  objref:
    kind: ConfigMap
    name: cert-manager-webhook-role-binding-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.certManagerNamespace
configurations:
- params.yaml
`)
}

func TestCertManagerWebhookRoleBindingBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/cert-manager/cert-manager-webhook-role-binding/base")
	writeCertManagerWebhookRoleBindingBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../cert-manager/cert-manager-webhook-role-binding/base"
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
