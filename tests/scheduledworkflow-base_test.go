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

func writeScheduledworkflowBase(th *KustTestHarness) {
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
apiVersion: apps/v1beta2
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
  newTag: '0.1.23'
`)
}

func TestScheduledworkflowBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/scheduledworkflow/base")
	writeScheduledworkflowBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/scheduledworkflow/base"
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
