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

func writeNotebookControllerOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/jupyter/notebook-controller/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: application
spec:
  componentKinds:
    - group: core
      kind: Service
    - group: apps
      kind: Deployment
    - group: core
      kind: ServiceAccount
  descriptor: 
    type: notebook-controller
    version: v1alpha1
    description: Notebooks controller allows users to create a custom resource \"Notebook\" (jupyter notebook).
    icons:
    maintainers:
    - name: Lun-kai Hsu
      email: lunkai@google.com
    owners:
    - name: Lun-kai Hsu
      email: lunkai@gogle.com
    keywords:
     - jupyter
     - notebook
     - notebook-controller
     - jupyterhub  
    links:
    - description: About
      url: "https://github.com/kubeflow/kubeflow/tree/master/components/notebook-controller"
  addOwnerRef: true
`)
	th.writeK("/manifests/jupyter/notebook-controller/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: notebook-controller
  app.kubernetes.io/instance: notebook-controller
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: notebook
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: role
subjects:
- kind: ServiceAccount
  name: service-account
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: role
rules:
- apiGroups:
  - apps
  resources:
  - statefulsets
  - deployments
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - notebooks
  - notebooks/status
  verbs:
  - '*'
- apiGroups:
  - networking.istio.io
  resources:
  - virtualservices
  verbs:
  - '*'
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: notebooks.kubeflow.org
spec:
  group: kubeflow.org
  names:
    kind: Notebook
    plural: notebooks
    singular: notebook
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/deployment.yaml", `
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - name: manager
        image: gcr.io/kubeflow-images-public/notebook-controller:v20190401-v0.4.0-rc.1-308-g33618cc9-e3b0c4
        command:
          - /manager
        env:
          - name: USE_ISTIO
            value: "false"
          - name: POD_LABELS
            value: $(POD_LABELS)
        imagePullPolicy: Always
      serviceAccountName: service-account
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: service
spec:
  ports:
  - port: 443
`)
	th.writeF("/manifests/jupyter/notebook-controller/base/params.env", `
POD_LABELS=gcp-cred-secret=user-gcp-sa,gcp-cred-secret-filename=user-gcp-sa.json
USE_ISTIO=false
`)
	th.writeK("/manifests/jupyter/notebook-controller/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- crd.yaml
- deployment.yaml
- service-account.yaml
- service.yaml
namePrefix: notebook-controller-
commonLabels:
  app: notebook-controller
  kustomize.component: notebook-controller
images:
  - name: gcr.io/kubeflow-images-public/notebook-controller
    newName: gcr.io/kubeflow-images-public/notebook-controller
    newTag: v20190603-v0-175-geeca4530-e3b0c4
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: POD_LABELS
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.POD_LABELS
- name: USE_ISTIO
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.USE_ISTIO
`)
}

func TestNotebookControllerOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/jupyter/notebook-controller/overlays/application")
	writeNotebookControllerOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../jupyter/notebook-controller/overlays/application"
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
