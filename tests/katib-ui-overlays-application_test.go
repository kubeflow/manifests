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

func writeKatibUiOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/katib-ui/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  componentKinds:
  - group: core
    kind: Service
  - group: apps
    kind: Deployment
  - group: core
    kind: ServiceAccount
  descriptor:
    type: "katib-ui"
    version: "v1alpha2"
    description: "Katib ui visualizes general trend of Hyperparameter space and each training history."
    maintainers:
    - name: Zhongxuan Wu
      email: wuzhongxuan@caicloud.io
    - name: Ce Gao
      email: gaoce@caicloud.io
    - name: Johnu George
      email: johnugeo@cisco.com
    - name: Hougang Liu
      email: liuhougang6@126.com
    - name: Richard Liu
      email: ricliu@google.com
    - name: YujiOshima
      email: yuji.oshima0x3fd@gmail.com
    owners:
    - name: Ce Gao
      email: gaoce@caicloud.io
    - name: Johnu George
      email: johnugeo@cisco.com
    - name: Hougang Liu
      email: liuhougang6@126.com
    - name: Richard Liu
      email: ricliu@google.com
    - name: YujiOshima
      email: yuji.oshima0x3fd@gmail.com
    keywords:
    - katib
    - katib-ui
    - hyperparameter tuning
    links:
    - description: About
      url: "https://github.com/kubeflow/katib"
  addOwnerRef: true
`)
	th.writeF("/manifests/katib-v1alpha2/katib-ui/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
- path: spec/selector/app.kubernetes.io\/instance
  kind: Service
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/template/metadata/labels/app.kubernetes.io\/instance
  kind: Deployment
`)
	th.writeF("/manifests/katib-v1alpha2/katib-ui/overlays/application/params.env", `
generateName=
`)
	th.writeK("/manifests/katib-v1alpha2/katib-ui/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../../base
- application.yaml
configMapGenerator:
- name: katib-ui-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: katib-ui-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: katib-ui 
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: katib
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/katib-v1alpha2/katib-ui/base/katib-ui-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: katib-ui
  labels:
    app: katib
    component: ui
spec:
  replicas: 1
  template:
    metadata:
      name: katib-ui
      labels:
        app: katib
        component: ui
    spec:
      containers:
      - name: katib-ui
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-ui:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        command:
          - './katib-ui'
        ports:
        - name: ui
          containerPort: 80
      serviceAccountName: katib-ui
`)
	th.writeF("/manifests/katib-v1alpha2/katib-ui/base/katib-ui-rbac.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-ui
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - "*"
- apiGroups:
  - kubeflow.org
  resources:
  - experiments
  - trials
  verbs:
  - "*"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: katib-ui
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: katib-ui
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: katib-ui
subjects:
- kind: ServiceAccount
  name: katib-ui
`)
	th.writeF("/manifests/katib-v1alpha2/katib-ui/base/katib-ui-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-ui
  labels:
    app: katib
    component: ui
spec:
  type: ClusterIP
  ports:
    - port: 80
      protocol: TCP
      name: ui
  selector:
    app: katib
    component: ui
`)
	th.writeF("/manifests/katib-v1alpha2/katib-ui/base/params.yaml", `
varReference:
- path: data/config
  kind: ConfigMap
- path: data/config
  kind: Deployment
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/katib-v1alpha2/katib-ui/base/params.env", `
clusterDomain=cluster.local
`)
	th.writeK("/manifests/katib-v1alpha2/katib-ui/base", `
namespace: kubeflow
resources:
- katib-ui-deployment.yaml
- katib-ui-rbac.yaml
- katib-ui-service.yaml
configMapGenerator:
- name: katib-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-ui
    newTag: v0.6.0-rc.0
vars:
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: katib-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
- name: namespace
  objref:
    kind: Service
    name: katib-ui
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
configurations:
- params.yaml
`)
}

func TestKatibUiOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/katib-ui/overlays/application")
	writeKatibUiOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/katib-ui/overlays/application"
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
