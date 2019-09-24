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

func writeCiPipelineOverlaysUpdateManifestsTask(th *KustTestHarness) {
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
    - name: kubeflow
      type: git
    - name: $(image_name)
      type: image
    - name: $(pull_request_repo)-$(pull_request_id)
      type: pullRequest
    params:
    - name: pathToManifestsDir
      type: string
      description: Where the components manifest dir is
      default: /workspace/manifests
    - name: container_image
      type: string
      description: pod container image
  outputs:
    resources:
    - name: manifests
      type: git
  steps:
  - name: update-manifests
    workingDir: "/workspace/${inputs.resources.manifests.name}/${inputs.params.pathToManifestsDir}"
    image: ${inputs.params.container_image}
    #command: ["/bin/sleep", "infinity"]
    command: ["/workspace/${inputs.resources.kubeflow.name}/py/kubeflow/kubeflow/ci/rebuild-manifests.sh"]
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    envFrom:
    - configMapRef:
        name: ci-pipeline-parameters
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
- path: spec/inputs/resources/name
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
    runAfter: 
    - build-push
    resources:
      inputs:
      - name: kubeflow
        resource: kubeflow
        from: 
        - build-push
      - name: manifests
        resource: manifests
      - name: $(pull_request_repo)-$(pull_request_id)
        resource: $(pull_request_repo)-$(pull_request_id)
      - name: $(image_name)
        resource: $(image_name)
        from: 
        - build-push
      outputs:
      - name: manifests
        resource: manifests
    params:
    - name: pathToManifestsDir
      value: "$(path_to_manifests_dir)"
    - name: container_image
      value: "$(container_image)"
- op: add
  path: /spec/resources/-
  value:
    name: manifests
    type: git
- op: add
  path: /spec/resources/-
  value:
    name: $(pull_request_repo)-$(pull_request_id)
    type: pullRequest
`)
	th.writeF("/manifests/ci/ci-pipeline/overlays/update-manifests-task/params.env", `
image_name=
path_to_manifests_dir=
container_image=
pull_request_id=
pull_request_repo=
`)
	th.writeK("/manifests/ci/ci-pipeline/overlays/update-manifests-task", `
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
- name: path_to_manifests_dir
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.path_to_manifests_dir
- name: container_image
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.container_image
- name: pull_request_repo
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pull_request_repo
- name: pull_request_id
  objref:
    kind: ConfigMap
    name: ci-pipeline-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pull_request_id
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

func TestCiPipelineOverlaysUpdateManifestsTask(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/ci/ci-pipeline/overlays/update-manifests-task")
	writeCiPipelineOverlaysUpdateManifestsTask(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../ci/ci-pipeline/overlays/update-manifests-task"
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
