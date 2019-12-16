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

func writeCentraldashboardOverlaysIstio(th *KustTestHarness) {
	th.writeF("/manifests/common/centraldashboard/overlays/istio/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: centraldashboard
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /
    rewrite:
      uri: /
    route:
    - destination:
        host: centraldashboard.$(namespace).svc.$(clusterDomain)
        port:
          number: 80
`)
	th.writeF("/manifests/common/centraldashboard/overlays/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeK("/manifests/common/centraldashboard/overlays/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- virtual-service.yaml
configurations:
- params.yaml

`)
	th.writeF("/manifests/common/centraldashboard/base/clusterrole-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: centraldashboard
subjects:
- kind: ServiceAccount
  name: centraldashboard
  namespace: $(namespace)
`)
	th.writeF("/manifests/common/centraldashboard/base/clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
rules:
- apiGroups:
  - ""
  resources:
  - events
  - namespaces
  - nodes
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/common/centraldashboard/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
spec:
  replicas: 1
  selector:
    matchLabels:
      app: centraldashboard
  template:
    metadata:
      labels:
        app: centraldashboard
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/centraldashboard
        imagePullPolicy: IfNotPresent
        name: centraldashboard
        ports:
        - containerPort: 8082
          protocol: TCP
        env:
        - name: USERID_HEADER
          value: $(userid-header)
        - name: USERID_PREFIX
          value: $(userid-prefix)
        - name: PROFILES_KFAM_SERVICE_HOST
          value: profiles-kfam.kubeflow
      serviceAccountName: centraldashboard
`)
	th.writeF("/manifests/common/centraldashboard/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: centraldashboard
subjects:
- kind: ServiceAccount
  name: centraldashboard
  namespace: $(namespace)
`)
	th.writeF("/manifests/common/centraldashboard/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app: centraldashboard
  name: centraldashboard
rules:
- apiGroups:
  - ""
  - "app.k8s.io"
  resources:
  - applications
  - pods
  - pods/exec
  - pods/log
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
`)
	th.writeF("/manifests/common/centraldashboard/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: centraldashboard
`)
	th.writeF("/manifests/common/centraldashboard/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: centralui-mapping
      prefix: /
      rewrite: /
      service: centraldashboard.$(namespace)
  labels:
    app: centraldashboard
  name: centraldashboard
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8082
  selector:
    app: centraldashboard
  sessionAffinity: None
  type: ClusterIP
`)
	th.writeF("/manifests/common/centraldashboard/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
- path: spec/http/route/destination/host
  kind: VirtualService
- path: spec/template/spec/containers/0/env/0/value
  kind: Deployment
- path: spec/template/spec/containers/0/env/1/value
  kind: Deployment`)
	th.writeF("/manifests/common/centraldashboard/base/params.env", `
clusterDomain=cluster.local
userid-header=
userid-prefix=`)
	th.writeK("/manifests/common/centraldashboard/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- clusterrole-binding.yaml
- clusterrole.yaml
- deployment.yaml
- role-binding.yaml
- role.yaml
- service-account.yaml
- service.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: centraldashboard
images:
- name: gcr.io/kubeflow-images-public/centraldashboard
  newName: gcr.io/kubeflow-images-public/centraldashboard
  digest: sha256:4299297b8390599854aa8f77e9eb717db684b32ca9a94a0ab0e73f3f73e5d8b5
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: Service
    name: centraldashboard
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
- name: userid-header
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.userid-header
- name: userid-prefix
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.userid-prefix
configurations:
- params.yaml

`)
}

func TestCentraldashboardOverlaysIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/common/centraldashboard/overlays/istio")
	writeCentraldashboardOverlaysIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../common/centraldashboard/overlays/istio"
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
