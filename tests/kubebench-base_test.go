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

func writeKubebenchBase(th *KustTestHarness) {
	th.writeF("/manifests/kubebench/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kubebench-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubebench-operator
subjects:
- kind: ServiceAccount
  name: default
`)
	th.writeF("/manifests/kubebench/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: kubebench-operator
rules:
- apiGroups:
  - kubebench.operator
  resources:
  - kubebenchjobs.kubebench.operator
  - kubebenchjobs
  verbs:
  - create
  - update
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  - pods
  - pods/exec
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  - secrets
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - pytorchjobs
  verbs:
  - '*'
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - '*'
`)
	th.writeF("/manifests/kubebench/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: kubebenchjobs.kubebench.operator
spec:
  group: kubebench.operator
  names:
    kind: KubebenchJob
    plural: kubebenchjobs
  scope: Namespaced
  version: v1
`)
	th.writeF("/manifests/kubebench/base/deployment.yaml", `
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: kubebench-dashboard
  name: kubebench-dashboard
spec:
  template:
    metadata:
      labels:
        app: kubebench-dashboard
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/kubebench/kubebench-dashboard:v0.4.0-13-g262c593
        name: kubebench-dashboard
        ports:
        - containerPort: 8084
      seviceAccountName: kubebench-dashboard
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kubebench-operator
spec:
  selector:
    matchLabels:
      app: kubebench-operator
  template:
    metadata:
      labels:
        app: kubebench-operator
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/kubebench/kubebench-operator-v1alpha2:v0.5.0-11-gea53ad5
        name: kubebench-operator
      seviceAccountName: kubebench-operator

`)
	th.writeF("/manifests/kubebench/base/role-binding.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  labels:
    app: kubebench-dashboard
  name: kubebench-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kubebench-dashboard
subjects:
- kind: ServiceAccount
  name: kubebench-dashboard
`)
	th.writeF("/manifests/kubebench/base/role.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  labels:
    app: kubebench-dashboard
  name: kubebench-dashboard
rules:
- apiGroups:
  - ""
  resources:
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
	th.writeF("/manifests/kubebench/base/service-account.yaml", `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubebench-dashboard
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
`)
	th.writeF("/manifests/kubebench/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: kubebench-dashboard-ui-mapping
      prefix: /dashboard/
      rewrite: /dashboard/
      service: kubebench-dashboard.$(namespace)
  name: kubebench-dashboard
spec:
  ports:
  - port: 80
    targetPort: 9303
  selector:
    app: kubebench-dashboard
`)
	th.writeF("/manifests/kubebench/base/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: kubebench-dashboard
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /dashboard/
    rewrite:
      uri: /dashboard/
    route:
    - destination:
        host: kubebench-dashboard.$(namespace).svc.$(clusterDomain)
        port:
          number: 80
`)
	th.writeF("/manifests/kubebench/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeF("/manifests/kubebench/base/params.env", `
clusterDomain=cluster.local
`)
	th.writeK("/manifests/kubebench/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- crd.yaml
- deployment.yaml
- role-binding.yaml
- role.yaml
- service-account.yaml
- service.yaml
- virtual-service.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: kubebench
configMapGenerator:
- name: parameters
  env: params.env
images:
  - name: gcr.io/kubeflow-images-public/kubebench/kubebench-dashboard
    newName: gcr.io/kubeflow-images-public/kubebench/kubebench-dashboard
    newTag: v0.4.0-13-g262c593
  - name: gcr.io/kubeflow-images-public/kubebench/kubebench-operator-v1alpha2
    newName: gcr.io/kubeflow-images-public/kubebench/kubebench-operator-v1alpha2
    newTag: v0.5.0-11-gea53ad5
  - name: gcr.io/kubeflow-images-public/kubebench/kubebench-controller
    newName: gcr.io/kubeflow-images-public/kubebench/kubebench-controller
    newTag: v0.4.0-13-g262c593
  - name: gcr.io/kubeflow-images-public/kubebench/kubebench-example-tf-cnn-post-processor
    newName: gcr.io/kubeflow-images-public/kubebench/kubebench-example-tf-cnn-post-processor
    newTag: v0.4.0-13-g262c593
vars:
- name: namespace
  objref:
    kind: Service
    name: kubebench-dashboard
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

func TestKubebenchBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/kubebench/base")
	writeKubebenchBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../kubebench/base"
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
