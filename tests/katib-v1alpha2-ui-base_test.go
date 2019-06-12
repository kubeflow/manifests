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

func writeKatibV1Alpha2UIBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/katib-ui/base/katib-ui-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: ui
  name: katib-ui
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: ui
      name: katib-ui
    spec:
      containers:
      - command:
        - ./katib-ui
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-ui:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        name: katib-ui
        ports:
        - containerPort: 80
          name: ui
      serviceAccountName: katib-ui
`)
	th.writeF("/manifests/katib-v1alpha2/katib-ui/base/katib-ui-rbac.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: katib-ui
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - experiments
  - trials
  verbs:
  - '*'
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: katib-ui
  namespace: kubeflow
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: katib-ui
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: katib-ui
subjects:
- kind: ServiceAccount
  name: katib-ui
  namespace: kubeflow
`)
	th.writeF("/manifests/katib-v1alpha2/katib-ui/base/katib-ui-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: ui
  name: katib-ui
  namespace: kubeflow
spec:
  ports:
  - name: ui
    port: 80
    protocol: TCP
  selector:
    app: katib
    component: ui
  type: ClusterIP
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
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-ui
    newTag: v0.1.2-alpha-280-gb0e0dd5
vars:
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: parameters
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
`)
}

func TestKatibV1Alpha2UIBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/katib-ui/base")
	writeKatibV1Alpha2UIBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/katib-ui/base"
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
