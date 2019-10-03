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

func writeCiPipelineRunOverlaysImageResource(th *KustTestHarness) {
	th.writeF("/manifests/ci/ci-pipeline-run/overlays/image-resource/pipeline-resource.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: PipelineResource
metadata:
  name: $(image_name)
spec:
  type: image
  params:
  - name: url
    value: $(image_url)
`)
	th.writeF("/manifests/ci/ci-pipeline-run/overlays/image-resource/params.yaml", `
varReference:
- path: metadata/name
  kind: PipelineResource
- path: spec/params/value
  kind: PipelineResource
- path: spec/resources/name
  kind: PipelineRun
- path: spec/resources/resourceRef/name
  kind: PipelineRun
`)
	th.writeF("/manifests/ci/ci-pipeline-run/overlays/image-resource/pipeline-run-patch.yaml", `
- op: add
  path: /spec/resources/-
  value:
    name: $(image_name)
    resourceRef:
      name: $(image_name)
`)
	th.writeF("/manifests/ci/ci-pipeline-run/overlays/image-resource/params.env", `
image_name=
image_url=
`)
	th.writeK("/manifests/ci/ci-pipeline-run/overlays/image-resource", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- pipeline-resource.yaml
patchesJson6902:
- target:
    group: tekton.dev
    version: v1alpha1
    kind: PipelineRun
    name: $(generateName)
  path: pipeline-run-patch.yaml
configMapGenerator:
- name: ci-pipeline-run-parameters
  behavior: merge
  env: params.env
vars:
- name: image_name
  objref:
    kind: ConfigMap
    name: ci-pipeline-run-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.image_name
- name: image_url
  objref:
    kind: ConfigMap
    name: ci-pipeline-run-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.image_url
configurations:
- params.yaml
`)
	th.writeF("/manifests/ci/ci-pipeline-run/base/persistent-volume-claim.yaml", `
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: ci-pipeline-run-persistent-volume-claim
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Gi
`)
	th.writeF("/manifests/ci/ci-pipeline-run/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ci-pipeline-run-service-account
imagePullSecrets:
- name: docker-secret
secrets:
- name: github-ssh
`)
	th.writeF("/manifests/ci/ci-pipeline-run/base/secrets.yaml", `
apiVersion: v1
data:
  .dockerconfigjson: redacted
kind: Secret
metadata:
  name: docker-secret
type: kubernetes.io/dockerconfigjson
---
apiVersion: v1
data:
  kaniko-secret.json: redacted
kind: Secret
metadata:
  name: kaniko-secret
type: Opaque
---
apiVersion: v1
data:
  key.json: redacted
kind: Secret
metadata:
  name: gcp-credentials
type: Opaque
---
apiVersion: v1
data:
  CLIENT_ID: redacted
  CLIENT_SECRET: redacted
kind: Secret
metadata:
  name: kubeflow-oauth
type: Opaque
---
apiVersion: v1
kind: Secret
metadata:
  name: github-ssh
  annotations:
    tekton.dev/git-0: github.com
type: kubernetes.io/ssh-auth
data:
  known_hosts: redacted
  ssh-privatekey: redacted
  ssh-publickey: redacted
---
apiVersion: v1
kind: Secret
metadata:
  name: github-token
type: Opaque
data:
  token: redacted
`)
	th.writeF("/manifests/ci/ci-pipeline-run/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: ci-pipeline-run-cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: ci-pipeline-run-service-account
`)
	th.writeF("/manifests/ci/ci-pipeline-run/base/pipeline-run.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: PipelineRun
metadata:
  name: $(generateName)
  labels:
    scope: $(namespace)
spec:
  serviceAccount: ci-pipeline-run-service-account
  pipelineRef:
    name: $(pipeline)
  resources: []
`)
	th.writeF("/manifests/ci/ci-pipeline-run/base/params.yaml", `
varReference:
- path: metadata/name
  kind: PipelineRun
- path: metadata/labels/scope
  kind: PipelineRun
- path: metadata/namespace
  kind: PersistentVolumeClaim
- path: spec/pipelineRef/name
  kind: PipelineRun
`)
	th.writeF("/manifests/ci/ci-pipeline-run/base/params.env", `
generateName=
namespace=
pipeline=
`)
	th.writeK("/manifests/ci/ci-pipeline-run/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- persistent-volume-claim.yaml
- service-account.yaml
- secrets.yaml
- cluster-role-binding.yaml
- pipeline-run.yaml
namespace: $(namespace)
configMapGenerator:
- name: ci-pipeline-run-parameters
  env: params.env
generatorOptions:
 disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: ci-pipeline-run-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
- name: generateName
  objref:
    kind: ConfigMap
    name: ci-pipeline-run-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
- name: pipeline
  objref:
    kind: ConfigMap
    name: ci-pipeline-run-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pipeline
configurations:
- params.yaml
`)
}

func TestCiPipelineRunOverlaysImageResource(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/ci/ci-pipeline-run/overlays/image-resource")
	writeCiPipelineRunOverlaysImageResource(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../ci/ci-pipeline-run/overlays/image-resource"
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
