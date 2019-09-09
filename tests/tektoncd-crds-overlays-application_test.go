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

func writeTektoncdCrdsOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/tektoncd/tektoncd-crds/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  componentKinds:
    - group: app.k8s.io
      kind: Application
  descriptor: 
    type: tekton-crds
    version: v1beta1
    description: tekton crds
    maintainers:
    - name: Kam Kasravi
      email: kam.d.kasravi@intel.com
    owners:
    - name: Kam Kasravi
      email: kam.d.kasravi@intel.com
    keywords:
     - kubeflow
    links:
    - description: About
      url: "https://kubeflow.org"
  addOwnerRef: true
`)
	th.writeF("/manifests/tektoncd/tektoncd-crds/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
`)
	th.writeF("/manifests/tektoncd/tektoncd-crds/overlays/application/params.env", `
generateName=
`)
	th.writeK("/manifests/tektoncd/tektoncd-crds/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: tektoncd-crds-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: tektoncd-crds-parameters 
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: tektoncd-crds
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: kubeflow
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/tektoncd/tektoncd-crds/base/crds.yaml", `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clustertasks.tekton.dev
spec:
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: ClusterTask
    plural: clustertasks
  scope: Cluster
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: images.caching.internal.knative.dev
spec:
  group: caching.internal.knative.dev
  names:
    categories:
    - all
    - knative-internal
    - caching
    kind: Image
    plural: images
    shortNames:
    - img
    singular: image
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: pipelines.tekton.dev
spec:
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: Pipeline
    plural: pipelines
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: pipelineruns.tekton.dev
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].status
    name: Succeeded
    type: string
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].reason
    name: Reason
    type: string
  - JSONPath: .status.startTime
    name: StartTime
    type: date
  - JSONPath: .status.completionTime
    name: CompletionTime
    type: date
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: PipelineRun
    plural: pipelineruns
    shortNames:
    - pr
    - prs
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: pipelineresources.tekton.dev
spec:
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: PipelineResource
    plural: pipelineresources
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tasks.tekton.dev
spec:
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: Task
    plural: tasks
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: taskruns.tekton.dev
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].status
    name: Succeeded
    type: string
  - JSONPath: .status.conditions[?(@.type=="Succeeded")].reason
    name: Reason
    type: string
  - JSONPath: .status.startTime
    name: StartTime
    type: date
  - JSONPath: .status.completionTime
    name: CompletionTime
    type: date
  group: tekton.dev
  names:
    categories:
    - all
    - tekton-pipelines
    kind: TaskRun
    plural: taskruns
    shortNames:
    - tr
    - trs
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha1
`)
	th.writeF("/manifests/tektoncd/tektoncd-crds/base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: tekton-pipelines
`)
	th.writeK("/manifests/tektoncd/tektoncd-crds/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- crds.yaml
- namespace.yaml
namespace: kubeflow
`)
}

func TestTektoncdCrdsOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/tektoncd/tektoncd-crds/overlays/application")
	writeTektoncdCrdsOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.EncodeAsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../tektoncd/tektoncd-crds/overlays/application"
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
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
