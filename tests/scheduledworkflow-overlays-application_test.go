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

func writeScheduledworkflowOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/scheduledworkflow/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: $(generateName)
  componentKinds:
  - group: apps
    kind: Deployment
  - group: rbac.authorization.k8s.io
    kind: Role
  - group: rbac.authorization.k8s.io
    kind: RoleBinding
  - group: core
    kind: ServiceAccount
  descriptor:
    type: "pipeline-scheduledworkflow"
    version: "v1beta1"
    description: "scheduledworkflow component of kubeflow pipelines"
    keywords:
    - "kubeflow"
    - "pipelines"
    links:
    - description: About
      url: "https://github.com/kubeflow/pipelines"
`)
	th.writeF("/manifests/pipeline/scheduledworkflow/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Application
- path: spec/selector/app.kubernetes.io\/instance
  kind: Service
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/template/metadata/labels/app.kubernetes.io\/instance
  kind: Deployment
`)
	th.writeF("/manifests/pipeline/scheduledworkflow/overlays/application/params.env", `
generateName=
`)
	th.writeK("/manifests/pipeline/scheduledworkflow/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: pipeline-scheduledworkflow-app-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: pipeline-scheduledworkflow-app-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: pipeline-scheduledworkflow
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: pipeline-scheduledworkflow
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/pipeline/scheduledworkflow/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: scheduledworkflows.kubeflow.org
spec:
  group: kubeflow.org
  names:
    kind: ScheduledWorkflow
    listKind: ScheduledWorkflowList
    plural: scheduledworkflows
    shortNames:
    - swf
    singular: scheduledworkflow
  scope: Namespaced
  versions:
  - name: v1beta1
    served: true
    storage: true
`)
	th.writeF("/manifests/pipeline/scheduledworkflow/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ml-pipeline-scheduledworkflow
spec:
  template:
    spec:
      containers:
      - name: ml-pipeline-scheduledworkflow
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/ml-pipeline/scheduledworkflow:0.1.23
        imagePullPolicy: IfNotPresent
      serviceAccountName: ml-pipeline-scheduledworkflow
`)
	th.writeF("/manifests/pipeline/scheduledworkflow/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: ml-pipeline-scheduledworkflow
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: ml-pipeline-scheduledworkflow
`)
	th.writeF("/manifests/pipeline/scheduledworkflow/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: ml-pipeline-scheduledworkflow
rules:
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - create
  - get
  - list
  - watch
  - update
  - patch
  - delete
- apiGroups:
  - kubeflow.org
  resources:
  - scheduledworkflows
  verbs:
  - create
  - get
  - list
  - watch
  - update
  - patch
  - delete
`)
	th.writeF("/manifests/pipeline/scheduledworkflow/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ml-pipeline-scheduledworkflow
`)
	th.writeK("/manifests/pipeline/scheduledworkflow/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
commonLabels:
  app: ml-pipeline-scheduledworkflow
resources:
- crd.yaml
- deployment.yaml
- role-binding.yaml
- role.yaml
- service-account.yaml
images:
- name: gcr.io/ml-pipeline/scheduledworkflow
  newTag: 0.1.23
  newName: gcr.io/ml-pipeline/scheduledworkflow
`)
}

func TestScheduledworkflowOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/scheduledworkflow/overlays/application")
	writeScheduledworkflowOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/scheduledworkflow/overlays/application"
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
