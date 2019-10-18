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

func writeCertManagerKubeSystemResourcesOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/cert-manager/cert-manager-kube-system-resources/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: cert-manager-kube-system-resources
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: cert-manager-kube-system-resources
      app.kubernetes.io/instance: cert-manager-kube-system-resources-v0.7.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: cert-manager-kube-system-resources
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v0.7.0
  componentKinds:
  - group: rbac
    kind: Role
  - group: rbac
    kind: RoleBinding
  descriptor:
    type: ""
    version: "v0.11.0"
    description: "Automatically provision and manage TLS certificates in Kubernetes https://jetstack.io."
    keywords:
    - cert-manager
    links:
    - description: About
      url: "https://github.com/jetstack/cert-manager"
`)
	th.writeK("/manifests/cert-manager/cert-manager-kube-system-resources/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: cert-manager-kube-system-resources
  app.kubernetes.io/instance: cert-manager-kube-system-resources-v0.7.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: cert-manager-kube-system-resources
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.7.0
`)
	th.writeF("/manifests/cert-manager/cert-manager-kube-system-resources/base/role-binding.yaml", `
# grant cert-manager permission to manage the leaderelection configmap in the
# leader election namespace
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: cert-manager-cainjector:leaderelection
  labels:
    app: cainjector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: cert-manager-cainjector:leaderelection
subjects:
- apiGroup: ""
  kind: ServiceAccount
  name: cert-manager-cainjector
  namespace: $(certManagerNamespace)

---

# grant cert-manager permission to manage the leaderelection configmap in the
# leader election namespace
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: cert-manager:leaderelection
  labels:
    app: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: cert-manager:leaderelection
subjects:
- apiGroup: ""
  kind: ServiceAccount
  name: cert-manager
  namespace: $(certManagerNamespace)

---

# apiserver gets the ability to read authentication. This allows it to
# read the specific configmap that has the requestheader-* entries to
# api agg
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: cert-manager-webhook:webhook-authentication-reader
  labels:
    app: webhook
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
	th.writeF("/manifests/cert-manager/cert-manager-kube-system-resources/base/role.yaml", `
# leader election rules
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: cert-manager-cainjector:leaderelection
  labels:
    app: cainjector
rules:
  # Used for leader election by the controller
  # TODO: refine the permission to *just* the leader election configmap
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "create", "update", "patch"]

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: cert-manager:leaderelection
  labels:
    app: cert-manager
rules:
  # Used for leader election by the controller
  # TODO: refine the permission to *just* the leader election configmap
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "create", "update", "patch"]
`)
	th.writeF("/manifests/cert-manager/cert-manager-kube-system-resources/base/params.yaml", `
varReference:
- path: subjects/namespace
  kind: RoleBinding
`)
	th.writeF("/manifests/cert-manager/cert-manager-kube-system-resources/base/params.env", `
certManagerNamespace=cert-manager
`)
	th.writeK("/manifests/cert-manager/cert-manager-kube-system-resources/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kube-system
resources:
- role-binding.yaml
- role.yaml
commonLabels:
  kustomize.component: cert-manager
configMapGenerator:
- name: cert-manager-kube-params-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: certManagerNamespace
  objref:
    kind: ConfigMap
    name: cert-manager-kube-params-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.certManagerNamespace
configurations:
- params.yaml
`)
}

func TestCertManagerKubeSystemResourcesOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/cert-manager/cert-manager-kube-system-resources/overlays/application")
	writeCertManagerKubeSystemResourcesOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../cert-manager/cert-manager-kube-system-resources/overlays/application"
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
