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

func writeTfJobCrdsOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/tf-training/tf-job-crds/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: tf-job-crds
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: tf-job-crds
      app.kubernetes.io/instance: tf-job-crds-v1.0.0
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: tfjob
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v1.0.0
  componentKinds:
  - group: core
    kind: Service
  - group: apps
    kind: Deployment
  - group: core
    kind: ServiceAccount
  - group: kubeflow.org
    kind: TFJob
  descriptor:
    type: "tf-job-crds"
    version: "v1"
    description: "Tf-job-crds contains the \"TFJob\" custom resource definition."
    maintainers:
    - name: Richard Liu
      email: ricliu@google.com
    owners:
    - name: Richard Liu
      email: ricliu@google.com
    keywords:
    - "tfjob"
    - "tf-operator"
    - "tf-training"
    links:
    - description: About
      url: "https://github.com/kubeflow/tf-operator"
    - description: Docs
      url: "https://www.kubeflow.org/docs/reference/tfjob/v1/tensorflow/"
  addOwnerRef: true
`)
	th.writeK("/manifests/tf-training/tf-job-crds/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: tf-job-crds
  app.kubernetes.io/instance: tf-job-crds-v1.0.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: tfjob
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v1.0.0
`)
	th.writeF("/manifests/tf-training/tf-job-crds/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tfjobs.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: State
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  names:
    kind: TFJob
    plural: tfjobs
    singular: tfjob
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            tfReplicaSpecs:
              properties:
                Chief:
                  properties:
                    replicas:
                      maximum: 1
                      minimum: 1
                      type: integer
                PS:
                  properties:
                    replicas:
                      minimum: 1
                      type: integer
                Worker:
                  properties:
                    replicas:
                      minimum: 1
                      type: integer
  versions:
  - name: v1
    served: true
    storage: true
`)
	th.writeK("/manifests/tf-training/tf-job-crds/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- crd.yaml
`)
}

func TestTfJobCrdsOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/tf-training/tf-job-crds/overlays/application")
	writeTfJobCrdsOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../tf-training/tf-job-crds/overlays/application"
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
