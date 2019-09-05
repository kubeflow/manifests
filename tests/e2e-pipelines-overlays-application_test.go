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

func writeE2ePipelinesOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/e2e/e2e-pipelines/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  componentKinds:
  - group: tekton.dev/v1alpha1
    kind: PipelineResource
  - group: tekton.dev/v1alpha1
    kind: Pipeline
  descriptor:
    type: tektoncd
    version: v1beta1
    description: Launches a PipelineRun
    maintainers: []
    owners: []
    keywords: []
    links:
    - description: About
      url: "" 
  addOwnerRef: true
`)
	th.writeF("/manifests/e2e/e2e-pipelines/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
`)
	th.writeF("/manifests/e2e/e2e-pipelines/overlays/application/params.env", `
generateName=
`)
	th.writeK("/manifests/e2e/e2e-pipelines/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: e2e-pipelines-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: e2e-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: e2e-pipelines
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: e2e
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/e2e/e2e-pipelines/base/pipeline.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: kfctl-build-apply
spec:
  resources:
  - name: kfctl-repo
    type: git
  - name: testing-repo
    type: git
  - name: web-image
    type: image
  tasks:
  - name: kfctl-build-push
    taskRef:
      name: kfctl-build-push
    params:
    - name: pathToDockerFile
      value: /workspace/kfctl/Dockerfile
    - name: pathToContext
      value: /workspace/kfctl
    resources:
      inputs:
      - name: kfctl
        resource: kfctl-repo
      outputs:
      - name: builtImage
        resource: web-image
  - name: kfctl-init-generate-apply
    taskRef:
      name: kfctl-init-generate-apply
    resources:
      inputs:
      - name: image
        resource: web-image
        from:
        - kfctl-build-push
      outputs:
      - name: builtImage
        resource: web-image
        from:
        - kfctl-build-push
    params:
    - name: namespace
      value: $(namespace)
    - name: app_dir
      value: $(app_dir)
    - name: project
      value: $(project)
    - name: configPath
      value: $(configPath)
    - name: zone
      value: $(zone)
    - name: email
      value: $(email)
    - name: platform
      value: $(platform)
    - name: cluster
      value: $(cluster)
`)
	th.writeF("/manifests/e2e/e2e-pipelines/base/pipeline-resource.yaml", `
---
apiVersion: tekton.dev/v1alpha1
kind: PipelineResource
metadata:
  name: kfctl-git
spec:
  type: git
  params:
  - name: revision
    value: $(kfctl_pullrequest)
  - name: url
    value: https://github.com/kubeflow/kfctl.git
---
apiVersion: tekton.dev/v1alpha1
kind: PipelineResource
metadata:
  name: testing-git
spec:
  type: git
  params:
  - name: revision
    value: $(testing_pullrequest)
  - name: url
    value: https://github.com/kubeflow/testing.git
---
apiVersion: tekton.dev/v1alpha1
kind: PipelineResource
metadata:
  name: kfctl-image
spec:
  type: image
  params:
  - name: url
    value: gcr.io/$(project)/kfctl
`)
	th.writeF("/manifests/e2e/e2e-pipelines/base/params.yaml", `
varReference:
- path: spec/tasks/params/value
  kind: Pipeline
- path: spec/params/value
  kind: PipelineResource
`)
	th.writeF("/manifests/e2e/e2e-pipelines/base/params.env", `
namespace=kubeflow
project=constant-cubist-173123
kfctl_pullrequest=refs/pull/10/head
testing_pullrequest=refs/pull/446/head
app_dir=/kubeflow/kubeflow-e2e
zone=us-west1-a
email=foo@bar.com
configPath=https://raw.githubusercontent.com/kubeflow/kubeflow/master/bootstrap/config/kfctl_gcp_iap.yaml
platform=all
cluster=kubeflow-e2e
`)
	th.writeK("/manifests/e2e/e2e-pipelines/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pipeline.yaml
- pipeline-resource.yaml
namespace: tekton-pipelines
configMapGenerator:
- name: kfctl-pipelines-parameters
  env: params.env
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
- name: project
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.project
- name: configPath
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.configPath
- name: kfctl_pullrequest
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.kfctl_pullrequest
- name: testing_pullrequest
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.testing_pullrequest
- name: app_dir
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.app_dir
- name: zone
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.zone
- name: email
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.email
- name: platform
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.platform
- name: cluster
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.cluster
configurations:
- params.yaml
`)
}

func TestE2ePipelinesOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/e2e/e2e-pipelines/overlays/application")
	writeE2ePipelinesOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.EncodeAsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../e2e/e2e-pipelines/overlays/application"
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
