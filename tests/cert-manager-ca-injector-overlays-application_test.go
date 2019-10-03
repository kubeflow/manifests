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

func writeCertManagerCAInjectorOverlaysApplicationBase(th *KustTestHarness) {
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/base/namespace.yaml", `
---
apiVersion: v1
kind: Namespace
metadata:
  name: $(namespace)
`)
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-cainjector
  labels:
    app: cainjector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-cainjector
subjects:
  - name: cert-manager-cainjector
    namespace: $(namespace)
    kind: ServiceAccount
`)
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-cainjector
  labels:
    app: cainjector
rules:
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["certificates"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["configmaps", "events"]
    verbs: ["get", "create", "update", "patch"]
  - apiGroups: ["admissionregistration.k8s.io"]
    resources: ["validatingwebhookconfigurations", "mutatingwebhookconfigurations"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["apiregistration.k8s.io"]
    resources: ["apiservices"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["get", "list", "watch", "update"]
`)
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager-cainjector
  labels:
    app: cainjector
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cainjector
  template:
    metadata:
      labels:
        app: cainjector
      annotations:
    spec:
      serviceAccountName: cert-manager-cainjector
      containers:
        - name: cainjector
          image: "quay.io/jetstack/cert-manager-cainjector:v0.10.0"
          imagePullPolicy: IfNotPresent
          args:
          - --v=2
          - --leader-election-namespace=$(POD_NAMESPACE)
          env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          resources:
            {}
`)
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cert-manager-cainjector
  labels:
    app: cainjector
`)
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/base/params.yaml", `
varReference:
- path: subjects/namespace
  kind: ClusterRoleBinding
- path: metadata/name
  kind: Namespace
`)
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/base/params.env", `
namespace=cert-manager
`)
	th.writeK("/manifests/cert-manager/cert-manager-ca-injector/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: cert-manager
resources:
- namespace.yaml
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- service-account.yaml
commonLabels:
  kustomize.component: cert-manager-ca-injector
configMapGenerator:
- name: cert-manager-ca-injector-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: cert-manager-ca-injector-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
configurations:
- params.yaml
`)
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name:  $(generateName)
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: cert-manager-ca-injector
      app.kubernetes.io/instance: $(generateName)
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: cert-manager
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v0.6
  componentKinds:
  - group: rbac
    kind: ClusterRole
  - group: rbac
    kind: ClusterRoleBinding
  - group: core
    kind: Namespace
  - group: apps
    kind: Deployment
  - group: core
    kind: ServiceAccount
  descriptor:
    type: ""
    version: "v0.10.0"
    description: "Automatically provision and manage TLS certificates in Kubernetes https://jetstack.io."
    keywords:
    - cert-manager
    - cert-manager-ca-injector
    links:
    - description: About
      url: "https://github.com/jetstack/cert-manager"
`)
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/overlays/application/params.env", `
generateName=cert-manager-ca-injector
`)
	th.writeF("/manifests/cert-manager/cert-manager-ca-injector/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Application
- path: spec/template/metadata/labels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/selector/app.kubernetes.io\/instance
  kind: Service
`)
	th.writeK("/manifests/cert-manager/cert-manager-ca-injector/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: cert-manager-ca-injector-app-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: cert-manager-ca-injector-app-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: cert-manager-ca-injector
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: cert-manager
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
}

func TestCertManagerCAInjectorOverlaysApplicationBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/cert-manager/cert-manager-ca-injector/overlays/application")
	writeCertManagerCAInjectorOverlaysApplicationBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../cert-manager/cert-manager-ca-injector/overlays/application"
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
