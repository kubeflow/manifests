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

func writeCiPipelineOverlaysUpdateManifestsTask(th *KustTestHarness) {
	th.writeF("/manifests/ci/ci-pipeline/overlays/update-manifests-task/config-map.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: update-manifests-commands
data:
  rebuild-manifests.sh: |-
    #!/usr/bin/env bash
    make generate; make test
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/update-manifests-task/task.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: update-manifests
spec:
  inputs:
    resources:
    - name: manifests
      type: git
    params:
    - name: pathToManifestsTestsDir
      type: string
      description: Where manifests tests are generated and run
      default: /workspace/manifests/tests
    - name: imageTag
      type: string
      description: tag of the image
    - name: container_image
      type: string
      description: pod container image
  outputs:
    resources:
    - name: manifests
      type: git
  steps:
  - name: update-manifests
    workingDir: "/workspace/${inputs.resources.manifests.name}/${inputs.params.pathToManifestsTestsDir}"
    image: ${inputs.params.container_image}
    command: ["/bin/sleep", "infinity"]
#    command: ["/bin/bash", "/kubeflow/rebuild-manifests.sh"]
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    volumeMounts:
    - name: kaniko-secret
      mountPath: /secret
    - name: update-manifests-commands
      mountPath: /kubeflow
  volumes:
  - name: kaniko-secret
    secret:
      secretName: kaniko-secret
  - name: update-manifests-commands
    configMap:
      name: update-manifests-commands
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/update-manifests-task/params.yaml", `
varReference:
- path: spec/tasks/resources/inputs/name
  kind: Pipeline
- path: spec/tasks/resources/inputs/resource
  kind: Pipeline
- path: spec/tasks/resources/outputs/name
  kind: Pipeline
- path: spec/tasks/resources/outputs/resource
  kind: Pipeline
- path: spec/tasks/resources/inputs/value
  kind: Pipeline
- path: spec/tasks/params/value
  kind: Pipeline
- path: spec/resources/name
  kind: Pipeline
- path: spec/outputs/resources/name
  kind: Task
- path: spec/steps/args
  kind: Task
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/update-manifests-task/pipeline_patch.yaml", `
- op: add
  path: /spec/tasks/-
  value:
    name: update-manifests
    taskRef:
      name: update-manifests
    resources:
      inputs:
      - name: manifests
        resource: manifests
      outputs:
      - name: manifests
        resource: manifests
    params:
    - name: pathToManifestsTestsDir
      value: "$(path_to_manifests_tests_dir)"
    - name: imageTag
      value: "$(image_tag)"
    - name: container_image
      value: "$(container_image)"
- op: add
  path: /spec/resources/-
  value:
    name: manifests
    type: git
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/update-manifests-task/params.env", `
path_to_manifests_tests_dir=
image_tag=
container_image=gcr.io/constant-cubist-173123/test-worker@sha256:6258613836e9e5cb4468a3c8841026af9fb3272b39b16fa3f8b2af3cf1d0f350
`)
	th.writeK("/manifests/ci/ci-pipeline/overlays/update-manifests-task", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- config-map.yaml
- task.yaml
namespace: kubeflow-ci
patchesJson6902:
- target:
    group: tekton.dev
    version: v1alpha1
    kind: Pipeline
    name: ci-pipeline
  path: pipeline_patch.yaml
configMapGenerator:
- name: update-manifests-parameters
  env: params.env
vars:
- name: path_to_manifests_tests_dir
  objref:
    kind: ConfigMap
    name: update-manifests-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.path_to_manifests_tests_dir
- name: image_tag
  objref:
    kind: ConfigMap
    name: update-manifests-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.image_tag
- name: container_image
  objref:
    kind: ConfigMap
    name: update-manifests-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.container_image
configurations:
- params.yaml
`)
	th.writeF("/manifests/ci/ci-pipeline/base/pipeline.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: ci-pipeline
spec:
  params: []
  resources: []
  tasks: []
`)
	th.writeK("/manifests/ci/ci-pipeline/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pipeline.yaml
namespace: kubeflow-ci
`)
}

func TestCiPipelineOverlaysUpdateManifestsTask(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/ci/ci-pipeline/overlays/update-manifests-task")
	writeCiPipelineOverlaysUpdateManifestsTask(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.EncodeAsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../ci/ci-pipeline/overlays/update-manifests-task"
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
