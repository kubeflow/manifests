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

func writeE2ePipelinerunsBase(th *KustTestHarness) {
	th.writeF("/manifests/e2e/e2e-pipelineruns/base/pipeline-run.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: PipelineRun
metadata:
  name: $(pipelinerun)
spec:
  serviceAccount: $(serviceAccount)
  pipelineRef:
    name: $(pipeline)
  resources:
    - name: source-repo
      resourceRef:
        name: kfctl-git
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
pipelinerun=kfctl-build-apply-pipeline-run
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
- name: pipelinerun
  objref:
    kind: ConfigMap
    name: kfctl-pipelineruns-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.pipelinerun
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
	expected, err := m.EncodeAsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../e2e/e2e-pipelineruns/base"
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
