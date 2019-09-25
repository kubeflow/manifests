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

func writeCiBuildPushTaskRunOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/ci/ci-build-push-task-run/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  componentKinds:
    - group: app.k8s.io
      kind: Application
  descriptor: 
    type: ci-build-push
    version: v1beta1
    description: application that builds an app images and pushes it to a registry
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
	th.writeF("/manifests/ci/ci-build-push-task-run/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
`)
	th.writeF("/manifests/ci/ci-build-push-task-run/overlays/application/params.env", `
generateName=
`)
	th.writeK("/manifests/ci/ci-build-push-task-run/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: ci-build-push-app-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: ci-build-push-app-parameters 
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: ci-build-push
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: kubeflow
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/ci/ci-build-push-task-run/base/task-run.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: TaskRun
metadata:
  name: $(generateName)
spec:
  serviceAccount: ci-create-cluster-service-account
  inputs:
    params:
    - name: namespace
      value: $(namespace)
    - name: app_dir
      value: $(app_dir)
    - name: cluster
      value: $(cluster)
    - name: kfctl_image
      value: $(kfctl_image)
    - name: pvc_mount_path
      value: $(pvc_mount_path)
    - name: pathToDockerFile
      value: $(path_to_docker_file)
    - name: pathToContext
      value: $(path_to_context)
    - name: dockerTarget
      value: $(docker_target)
    resources:
    - name: kubeflow
      resourceSpec:
        type: git
        params:
          - name: revision
            value: master
          - name: url
            value: https://github.com/kubeflow/kubeflow.git
  outputs:
    resources:
    - name: builtImage
      outputImageDir: /workspace/builtImage
      resourceSpec:
        type: image
        params:
        - name: url
          value: $(IMG)
  taskSpec:
    inputs:
      resources:
      - name: kubeflow
        type: git
      params:
      - name: pathToDockerFile
        type: string
        description: The path to the dockerfile to build
        default: /workspace/kubeflow/Dockerfile
      - name: pathToContext
        type: string
        description:
          The build context used by Kaniko
          (https://github.com/GoogleContainerTools/kaniko#kaniko-build-contexts)
        default: /workspace/kubeflow
    outputs:
      resources:
      - name: builtImage
        type: image
        outputImageDir: /workspace/builtImage
    steps:
    - name: build-and-push
      image: gcr.io/kaniko-project/executor:v0.10.0
      command:
      - /kaniko/executor
      env:
      - name: GOOGLE_APPLICATION_CREDENTIALS
        value: /secret/kaniko-secret.json
      args: ["--dockerfile=$(inputs.params.pathToDockerFile)",
             "--destination=$(outputs.resources.builtImage.url)",
             "--context=$(inputs.params.pathToContext)",
             "--target=$(inputs.params.dockerTarget)"]
      volumeMounts:
      - name: kaniko-secret
        mountPath: /secret
    volumes:
    - name: kaniko-secret
      secret:
        secretName: kaniko-secret
`)
	th.writeF("/manifests/ci/ci-build-push-task-run/base/params.env", `
generateName=
namespace=kubeflow-ci
app_dir=/kubeflow/kubeflow-ci
pvc_mount_path=/kubeflow
kfctl_image=gcr.io/constant-cubist-173123/kfctl@sha256:ab0c4986322e3e6a755056278c7270983b0f3bdc0751aefff075fb2b3d0c3254
path_to_docker_file=
path_to_context=
docker_target=
IMG=
`)
	th.writeK("/manifests/ci/ci-build-push-task-run/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- task-run.yaml
namespace: ci-build-push
`)
}

func TestCiBuildPushTaskRunOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/ci/ci-build-push-task-run/overlays/application")
	writeCiBuildPushTaskRunOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../ci/ci-build-push-task-run/overlays/application"
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
