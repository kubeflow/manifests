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

func writeCiBuildPushTaskRunBase(th *KustTestHarness) {
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
      args: ["--dockerfile=${inputs.params.pathToDockerFile}",
             "--destination=${outputs.resources.builtImage.url}",
             "--context=${inputs.params.pathToContext}",
             "--target=${inputs.params.dockerTarget}"]
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

func TestCiBuildPushTaskRunBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/ci/ci-build-push-task-run/base")
	writeCiBuildPushTaskRunBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.EncodeAsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../ci/ci-build-push-task-run/base"
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
