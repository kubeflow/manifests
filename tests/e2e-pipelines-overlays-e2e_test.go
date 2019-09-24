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

func writeE2ePipelinesOverlaysE2e(th *KustTestHarness) {
	th.writeF("/manifests/e2e/e2e-pipelines/overlays/e2e/params.yaml", `
varReference:
- path: spec/tasks/params/value
  kind: Pipeline
- path: spec/params/value
  kind: PipelineResource
`)
	th.writeF("/manifests/e2e/e2e-pipelines/overlays/e2e/pipeline.yaml", `
- op: add
  path: /spec/tasks/-
  value:
    name: kfctl-e2e
    taskRef:
      name: kfctl-e2e
    resources:
      inputs:
      - name: kfctl
        resource: kfctl-repo
      - name: testing
        resource: testing-repo
      - name: image
        resource: web-image
        from:
        - kfctl-init-generate-apply
    params:
    - name: image
      value: $(image)
    - name: project
      value: $(project)
    - name: bucket
      value: $(bucket)
    - name: cluster
      value: $(cluster)
    - name: repos_dir
      value: $(repos_dir)
    - name: zone
      value: $(zone)
    - name: configPath
      value: $(configPath)
    - name: email
      value: $(email)
    - name: platform
      value: $(platform)
    - name: REPO_OWNER
      value: $(REPO_OWNER)
    - name: REPO_NAME
      value: $(REPO_NAME)
    - name: config_file
      value: /src/$(REPO_OWNER)/$(REPO_NAME)/prow_config.yaml
    - name: JOB_NAME
      value: $(JOB_NAME)
    - name: JOB_TYPE
      value: $(JOB_TYPE)
    - name: PULL_NUMBER
      value: "$(PULL_NUMBER)"
    - name: PULL_BASE_REF
      value: "$(PULL_BASE_REF)"
    - name: PULL_PULL_SHA
      value: "$(PULL_PULL_SHA)"
    - name: BUILD_NUMBER
      value: "$(BUILD_NUMBER)"
    - name: tests
      value: "$(tests)"
`)
	th.writeF("/manifests/e2e/e2e-pipelines/overlays/e2e/params.env", `
image=gcr.io/kubeflow-ci/test-worker:latest
project=kubeflow-ci
bucket=kubernetes-jenkins
repos_dir=/src
zone=us-west1-a
configPath=https://raw.githubusercontent.com/kubeflow/kubeflow/master/bootstrap/config/kfctl_gcp_iap.yaml
email=kfctl-e2e@constant-cubist-173123.iam.gserviceaccount.com
platform=all
REPO_OWNER=kubeflow
REPO_NAME=kfctl
config_file=prow_config.yaml
JOB_NAME=kubeflow-test-presubmit-test
JOB_TYPE=presubmit
PULL_NUMBER=33
PULL_BASE_REF=master
PULL_PULL_SHA=123abc
BUILD_NUMBER=a3bc
tests=
`)
	th.writeK("/manifests/e2e/e2e-pipelines/overlays/e2e", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
patchesJson6902:
- target:
    group: tekton.dev
    version: v1alpha1
    kind: Pipeline
    name: kfctl-build-apply
  path: pipeline.yaml
configMapGenerator:
- name: kfctl-e2e-parameters
  env: params.env
vars:
- name: image
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.image
- name: bucket
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.bucket
- name: repos_dir
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.repos_dir
- name: REPO_OWNER
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.REPO_OWNER
- name: REPO_NAME
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.REPO_NAME
- name: JOB_NAME
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.JOB_NAME
- name: JOB_TYPE
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.JOB_TYPE
- name: PULL_NUMBER
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.PULL_NUMBER
- name: PULL_BASE_REF
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.PULL_BASE_REF
- name: PULL_PULL_SHA
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.PULL_PULL_SHA
- name: BUILD_NUMBER
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.BUILD_NUMBER
- name: tests
  objref:
    kind: ConfigMap
    name: kfctl-e2e-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.tests
configurations:
- params.yaml
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

func TestE2ePipelinesOverlaysE2e(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/e2e/e2e-pipelines/overlays/e2e")
	writeE2ePipelinesOverlaysE2e(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../e2e/e2e-pipelines/overlays/e2e"
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
