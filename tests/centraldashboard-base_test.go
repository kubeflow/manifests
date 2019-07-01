package tests_test

import (
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/target"
	"testing"
)

func writeCentraldashboardBase(th *KustTestHarness) {
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
apiVersion: extensions/v1beta1
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
      - image: gcr.io/kubeflow-images-public/centraldashboard:v0.5.0
        imagePullPolicy: IfNotPresent
        name: centraldashboard
        ports:
        - containerPort: 8082
          protocol: TCP
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
`)
	th.writeF("/manifests/common/centraldashboard/base/params.env", `
clusterDomain=cluster.local
`)
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
    newTag: v0.6.0-rc2
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
configurations:
- params.yaml

`)
}

func TestCentraldashboardBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/common/centraldashboard/base")
	writeCentraldashboardBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../common/centraldashboard/base"
	fsys := fs.MakeRealFS()
	_loader, loaderErr := loader.NewLoader(targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl())
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	n, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := n.EncodeAsYaml()
	th.assertActualEqualsExpected(m, string(expected))
}
