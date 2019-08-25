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

func writeE2ePipelinesOverlaysE2e(th *KustTestHarness) {
	th.writeF("/manifests/e2e/e2e-pipelines/overlays/e2e/pipeline.yaml", `
- op: add
  path: /spec/tasks/-
  value:
    name: kfctl-e2e
    taskRef:
      name: kfctl-e2e
    resources:
      inputs:
      - name: image
        resource: web-image
        from:
        - kfctl-init-generate-apply
    params:
    - name: image
      value: $(image)
    - name: project
      value: $(project)
    - name: cluster
      value: $(cluster)
    - name: bucket
      value: $(bucket)
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
`)
	th.writeF("/manifests/e2e/e2e-pipelines/overlays/e2e/params.env", `
image=gcr.io/kubeflow-ci/test-worker:latest
project=kubeflow-ci
cluster=kubeflow-testing
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
`)
	th.writeF("/manifests/e2e/e2e-pipelines/base/pipeline.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: kfctl-build-apply
spec:
  resources:
  - name: source-repo
    type: git
  - name: web-image
    type: image
  tasks:
  - name: kfctl-build-push
    taskRef:
      name: kfctl-build-push
    params:
    - name: pathToDockerFile
      value: /workspace/docker-source/Dockerfile
    - name: pathToContext
      value: /workspace/docker-source
    resources:
      inputs:
      - name: docker-source
        resource: source-repo
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
    value: $(pullrequest)
  - name: url
    value: https://github.com/kubeflow/kfctl.git
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
	th.writeF("/manifests/e2e/e2e-pipelines/base/pipeline-run.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: PipelineRun
metadata:
  name: kfctl-build-apply-pipeline-run
spec:
  serviceAccount: e2e-pipelines
  pipelineRef:
    name: kfctl-build-apply
  resources:
    - name: source-repo
      resourceRef:
        name: kfctl-git
    - name: web-image
      resourceRef:
        name: kfctl-image
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
pullrequest=refs/pull/10/head
app_dir=/kubeflow/kubeflow-e2e
zone=us-west1-a
email=foo@bar.com
configPath=https://raw.githubusercontent.com/kubeflow/kubeflow/master/bootstrap/config/kfctl_gcp_iap.yaml
platform=all
`)
	th.writeK("/manifests/e2e/e2e-pipelines/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pipeline.yaml
- pipeline-resource.yaml
- pipeline-run.yaml
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
- name: pullrequest
  objref:
    kind: ConfigMap
    name: kfctl-pipelines-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pullrequest
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
	targetPath := "../e2e/e2e-pipelines/overlays/e2e"
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
