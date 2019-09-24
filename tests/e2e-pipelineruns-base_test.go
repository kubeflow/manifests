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

func writeE2ePipelinerunsBase(th *KustTestHarness) {
	th.writeF("/manifests/e2e/e2e-pipelineruns/base/pipeline-run.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: PipelineRun
metadata:
  name: $(generateName)
spec:
  serviceAccount: $(serviceAccount)
  pipelineRef:
    name: $(pipeline)
  resources:
    - name: kfctl-repo
      resourceRef:
        name: kfctl-git
    - name: testing-repo
      resourceRef:
        name: testing-git
    - name: web-image
      resourceRef:
        name: kfctl-image
`)
	th.writeF("/manifests/e2e/e2e-pipelineruns/base/params.yaml", `
varReference:
- path: metadata/name
  kind: PipelineRun
- path: spec/pipelineRef/name
  kind: PipelineRun
- path: spec/serviceAccount
  kind: PipelineRun
`)
	th.writeF("/manifests/e2e/e2e-pipelineruns/base/params.env", `
generateName=
pipeline=kfctl-build-apply
serviceAccount=e2e-pipelines
`)
	th.writeK("/manifests/e2e/e2e-pipelineruns/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- pipeline-run.yaml
namespace: tekton-pipelines
configMapGenerator:
- name: kfctl-pipelineruns-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: kfctl-pipelineruns-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
- name: pipeline
  objref:
    kind: ConfigMap
    name: kfctl-pipelineruns-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pipeline
- name: serviceAccount
  objref:
    kind: ConfigMap
    name: kfctl-pipelineruns-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.serviceAccount
configurations:
- params.yaml
`)
}

func TestE2ePipelinerunsBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/e2e/e2e-pipelineruns/base")
	writeE2ePipelinerunsBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../e2e/e2e-pipelineruns/base"
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
