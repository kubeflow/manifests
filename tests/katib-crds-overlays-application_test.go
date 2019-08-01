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

func writeKatibCrdsOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/katib-crds/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: katib-crds
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: katib-crds
      app.kubernetes.io/instance: katib-crds 
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: katib
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v0.6
  componentKinds:
  - group: kubeflow.org
    kind: Experiment
  - group: kubeflow.org
    kind: Trial
  descriptor:
    type: "katib-crds"
    version: "v1alpha2"
    description: "Katib-crds contains \"Experiment\" and \"Trial\" CRDs which are used by katib."
    maintainers:
    - name: Zhongxuan Wu
      email: wuzhongxuan@caicloud.io
    - name: Ce Gao
      email: gaoce@caicloud.io
    - name: Johnu George
      email: johnugeo@cisco.com
    - name: Hougang Liu
      email: liuhougang6@126.com
    - name: Richard Liu
      email: ricliu@google.com
    - name: YujiOshima
      email: yuji.oshima0x3fd@gmail.com
    owners:
    - name: Ce Gao
      email: gaoce@caicloud.io
    - name: Johnu George
      email: johnugeo@cisco.com
    - name: Hougang Liu
      email: liuhougang6@126.com
    - name: Richard Liu
      email: ricliu@google.com
    - name: YujiOshima
      email: yuji.oshima0x3fd@gmail.com
    keywords:
    - katib
    - katib-crds
    - hyperparameter tuning
    - neural architecture search
    links:
    - description: About
      url: "https://github.com/kubeflow/katib"
  addOwnerRef: true
`)
	th.writeK("/manifests/katib-v1alpha2/katib-crds/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: katib-crds
  app.kubernetes.io/instance: katib-crds
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: katib
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/katib-v1alpha2/katib-crds/base/experiment-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: experiments.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: Status
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  version: v1alpha2
  scope: Namespaced
  subresources:
    status: {}
  names:
    kind: Experiment
    singular: experiment
    plural: experiments
    categories:
    - all
    - kubeflow
    - katib
`)
	th.writeF("/manifests/katib-v1alpha2/katib-crds/base/trial-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: trials.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: Status
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  version: v1alpha2
  scope: Namespaced
  subresources:
    status: {}
  names:
    kind: Trial
    singular: trial
    plural: trials
    categories:
    - all
    - kubeflow
    - katib
`)
	th.writeK("/manifests/katib-v1alpha2/katib-crds/base", `
namespace: kubeflow
resources:
- experiment-crd.yaml
- trial-crd.yaml
generatorOptions:
  disableNameSuffixHash: true
`)
}

func TestKatibCrdsOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/katib-crds/overlays/application")
	writeKatibCrdsOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/katib-crds/overlays/application"
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
