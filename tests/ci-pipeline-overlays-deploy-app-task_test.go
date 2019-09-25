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

func writeCiPipelineOverlaysDeployAppTask(th *KustTestHarness) {
	th.writeF("/manifests/ci/ci-pipeline/overlays/deploy-app-task/task.yaml", `
---
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: build-push
spec:
  inputs:
    resources:
    - name: kfctl
      type: git
    params:
    - name: pathToDockerFile
      type: string
      description: The path to the dockerfile to build
      default: /workspace/kfctl/Dockerfile
    - name: pathToContext
      type: string
      description:
        The build context used by Kaniko
        (https://github.com/GoogleContainerTools/kaniko#kaniko-build-contexts)
      default: /workspace/kfctl
  outputs:
    resources:
    - name: builtImage
      type: image
      outputImageDir: /workspace/builtImage
  steps:
  - name: kfctl-build-and-push
    image: gcr.io/kaniko-project/executor:v0.10.0
    command:
    - /kaniko/executor
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    args: ["--dockerfile=$(inputs.params.pathToDockerFile)",
           "--destination=$(outputs.resources.builtImage.url)",
           "--context=$(inputs.params.pathToContext)",
           "--target=kfctl_base"]
    volumeMounts:
    - name: kaniko-secret
      mountPath: /secret
  volumes:
  - name: kaniko-secret
    secret:
      secretName: kaniko-secret
---
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: kfctl-init-generate-apply
spec:
  inputs:
    resources:
    - name: image
      type: image
    params:
    - name: namespace
      type: string
      description: the namespace to deploy kf 
    - name: app_dir
      type: string
      description: where to create the kf app
    - name: configPath
      type: string
      description: url for config arg
    - name: project
      type: string
      description: name of project
    - name: zone
      type: string
      description: zone of project
    - name: platform
      type: string
      description: all | k8s
    - name: email
      type: string
      description: email for gcp
    - name: cluster
      type: string
      description: name of the cluster
  outputs:
    resources:
    - name: builtImage
      type: image
      outputImageDir: /workspace/builtImage
  steps:
  - name: kfctl-init
    image: "$(inputs.resources.image.url)"
    command: ["/usr/local/bin/kfctl"]
    args:
    - "init"
    - "--config"
    - "$(inputs.params.configPath)"
    - "--project"
    - "$(inputs.params.project)"
    - "--namespace"
    - "$(inputs.params.namespace)"
    - "$(inputs.params.app_dir)"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    volumeMounts:
    - name: kaniko-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
    imagePullPolicy: IfNotPresent
  - name: kfctl-generate
    image: "$(inputs.resources.image.url)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.app_dir)"
    command: ["/usr/local/bin/kfctl"]
    args:
    - "generate"
    - "$(inputs.params.platform)"
    - "--zone"
    - "$(inputs.params.zone)"
    - "--email"
    - "$(inputs.params.email)"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: kaniko-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-activate-service-account
    image: "$(inputs.resources.image.url)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.app_dir)"
    command: ["/opt/google-cloud-sdk/bin/gcloud"]
    args:
    - "auth"
    - "activate-service-account"
    - "--key-file"
    - "/secret/ci-secret.json"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: ci-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-set-account
    image: "$(inputs.resources.image.url)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.app_dir)"
    command: ["/opt/google-cloud-sdk/bin/gcloud"]
    args:
    - "config"
    - "set"
    - "account"
    - "$(inputs.params.email)"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: ci-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-apply
    image: "$(inputs.resources.image.url)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.app_dir)"
#    command: ["/bin/sleep", "infinity"]
    command: ["/usr/local/bin/kfctl"]
    args:
    - "apply"
    - "$(inputs.params.platform)"
    - "--verbose"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: ci-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-configure-kubectl
    image: "$(inputs.resources.image.url)"
    imagePullPolicy: IfNotPresent
    workingDir: "$(inputs.params.app_dir)"
    command: ["/opt/google-cloud-sdk/bin/gcloud"]
    args:
    - "--project"
    - "$(inputs.params.project)"
    - "container"
    - "clusters"
    - "--zone"
    - "$(inputs.params.zone)"
    - "get-credentials"
    - "$(inputs.params.cluster)"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/ci-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: ci-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  volumes:
  - name: kaniko-secret
    secret:
      secretName: kaniko-secret
  - name: docker-secret
    secret:
      secretName: docker-secret
  - name: ci-secret
    secret:
      secretName: ci-secret
  - name: kubeflow
    persistentVolumeClaim:
      claimName: kubeflow-pvc
`)
	th.writeK("/manifests/ci/ci-pipeline/overlays/deploy-app-task", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- task.yaml
namespace: kubeflow-ci
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

func TestCiPipelineOverlaysDeployAppTask(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/ci/ci-pipeline/overlays/deploy-app-task")
	writeCiPipelineOverlaysDeployAppTask(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../ci/ci-pipeline/overlays/deploy-app-task"
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
