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

func writeCiPipelineOverlaysBuildPushTask(th *KustTestHarness) {
	th.writeF("/manifests/ci/ci-pipeline/overlays/build-push-task/task.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: build-push
spec:
  inputs:
    resources:
    - name: kubeflow
      type: git
    params:
    - name: imageName
      type: string
    - name: pathToDockerfile
      type: string
      description: The path to the dockerfile to build
    - name: pathToContext
      type: string
      description: The build context used by Kaniko
    - name: dockerTarget
      type: string
      description: docker target arg
  outputs:
    resources:
    - name: $(image_name)
      type: image
      outputImageDir: /kubeflow
  steps:
  - name: build-push
    image: gcr.io/kaniko-project/executor:v0.11.0
    command:
    - /kaniko/executor
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    args:
    - "--dockerfile=/workspace/$(inputs.resources.kubeflow.name)/$(inputs.params.pathToDockerfile)"
    - "--destination=$(outputs.resources.$(inputs.params.imageName).url)"
    - "--context=/workspace/$(inputs.resources.kubeflow.name)/$(inputs.params.pathToContext)"
    - "--target=$(inputs.params.dockerTarget)"
    - "--digest-file=/kubeflow/$(image_name)-digest"
    volumeMounts:
    - name: kaniko-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  volumes:
  - name: kaniko-secret
    secret:
      secretName: kaniko-secret
  - name: kubeflow
    persistentVolumeClaim:
      claimName: ci-pipeline-run-persistent-volume-claim
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/build-push-task/params.yaml", `
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
- path: spec/outputs/resources/outputImageDir
  kind: Task
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/build-push-task/pipeline_patch.yaml", `
- op: add
  path: /spec/tasks/-
  value:
    name: build-push 
    taskRef:
      name: build-push
    resources:
      inputs:
      - name: kubeflow
        resource: kubeflow
      outputs:
      - name: $(image_name)
        resource: $(image_name)
    params:
    - name: imageName
      value: "$(image_name)"
    - name: pathToDockerfile
      value: "$(path_to_docker_file)"
    - name: pathToContext
      value: "$(path_to_context)"
    - name: dockerTarget
      value: "$(docker_target)"
- op: add
  path: /spec/resources/-
  value:
    name: kubeflow
    type: git
- op: add
  path: /spec/resources/-
  value:
    name: $(image_name)
    type: image
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/build-push-task/params.env", `
image_name=
path_to_docker_file=
path_to_context=
docker_target=
`)
	th.writeK("/manifests/ci/ci-pipeline/overlays/build-push-task", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
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
- name: ci-pipeline-parameters
  behavior: merge
  env: params.env
vars:
- name: image_name
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.image_name
- name: path_to_docker_file
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.path_to_docker_file
- name: path_to_context
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.path_to_context
- name: docker_target
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.docker_target
configurations:
- params.yaml
`)
	th.writeF("/manifests/ci/ci-pipeline/base/pipeline.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Pipeline
metadata:
  name: ci-pipeline
  labels:
    scope: $(namespace)
spec:
  params: []
  resources: []
  tasks: []
`)
	th.writeF("/manifests/ci/ci-pipeline/base/params.env", `
namespace=
`)
	th.writeK("/manifests/ci/ci-pipeline/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pipeline.yaml
namespace: $(namespace)
configMapGenerator:
- name: ci-pipeline-parameters
  env: params.env
generatorOptions:
 disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
`)
}

func TestCiPipelineOverlaysBuildPushTask(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/ci/ci-pipeline/overlays/build-push-task")
	writeCiPipelineOverlaysBuildPushTask(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../ci/ci-pipeline/overlays/build-push-task"
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
