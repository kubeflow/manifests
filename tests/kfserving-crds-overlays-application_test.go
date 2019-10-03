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

func writeKfservingCrdsOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/kfserving/kfserving-crds/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: $(generateName)
  descriptor:
    type: kfserving-crds
    version: v1beta1
    description: "crds for kfserving"
    keywords:
    - "kfserving-crds"
    links:
    - description: About
      url: "https://kubeflow.org"
`)
	th.writeF("/manifests/kfserving/kfserving-crds/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Application
`)
	th.writeF("/manifests/kfserving/kfserving-crds/overlays/application/params.env", `
generateName=
`)
	th.writeK("/manifests/kfserving/kfserving-crds/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: kfserving-crds-app-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: kfserving-crds-app-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: kfserving-crds
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: kfserving
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/kfserving/kfserving-crds/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  labels:
    controller-tools.k8s.io: "1.0"
  name: kfservices.serving.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.url
    name: URL
    type: string
  - JSONPath: .status.default.traffic
    name: Default Traffic
    type: integer
  - JSONPath: .status.canary.traffic
    name: Canary Traffic
    type: integer
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: serving.kubeflow.org
  names:
    kind: KFService
    plural: kfservices
    shortNames:
    - kfservice
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          properties:
            canary:
              description: Canary defines an alternate configuration to route a percentage
                of traffic.
              properties:
                custom:
                  description: The following fields follow a "1-of" semantic. Users
                    must specify exactly one spec.
                  properties:
                    container:
                      type: object
                  required:
                  - container
                  type: object
                maxReplicas:
                  description: This is the up bound for autoscaler to scale to
                  format: int64
                  type: integer
                minReplicas:
                  description: Minimum number of replicas, pods won't scale down to
                    0 in case of no traffic
                  format: int64
                  type: integer
                pytorch:
                  properties:
                    modelClassName:
                      description: Defaults PyTorch model class name to 'PyTorchModel'
                      type: string
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest PyTorch Version
                      type: string
                  required:
                  - modelUri
                  type: object
                serviceAccountName:
                  description: Service Account Name
                  type: string
                sklearn:
                  properties:
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest SKLearn Version.
                      type: string
                  required:
                  - modelUri
                  type: object
                tensorflow:
                  properties:
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest TF Version.
                      type: string
                  required:
                  - modelUri
                  type: object
                tensorrt:
                  properties:
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest TensorRT Version.
                      type: string
                  required:
                  - modelUri
                  type: object
                xgboost:
                  properties:
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest XGBoost Version.
                      type: string
                  required:
                  - modelUri
                  type: object
              type: object
            canaryTrafficPercent:
              format: int64
              type: integer
            default:
              properties:
                custom:
                  description: The following fields follow a "1-of" semantic. Users
                    must specify exactly one spec.
                  properties:
                    container:
                      type: object
                  required:
                  - container
                  type: object
                maxReplicas:
                  description: This is the up bound for autoscaler to scale to
                  format: int64
                  type: integer
                minReplicas:
                  description: Minimum number of replicas, pods won't scale down to
                    0 in case of no traffic
                  format: int64
                  type: integer
                pytorch:
                  properties:
                    modelClassName:
                      description: Defaults PyTorch model class name to 'PyTorchModel'
                      type: string
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest PyTorch Version
                      type: string
                  required:
                  - modelUri
                  type: object
                serviceAccountName:
                  description: Service Account Name
                  type: string
                sklearn:
                  properties:
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest SKLearn Version.
                      type: string
                  required:
                  - modelUri
                  type: object
                tensorflow:
                  properties:
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest TF Version.
                      type: string
                  required:
                  - modelUri
                  type: object
                tensorrt:
                  properties:
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest TensorRT Version.
                      type: string
                  required:
                  - modelUri
                  type: object
                xgboost:
                  properties:
                    modelUri:
                      type: string
                    resources:
                      description: Defaults to requests and limits of 1CPU, 2Gb MEM.
                      type: object
                    runtimeVersion:
                      description: Defaults to latest XGBoost Version.
                      type: string
                  required:
                  - modelUri
                  type: object
              type: object
          required:
          - default
          type: object
        status:
          properties:
            canary:
              properties:
                name:
                  type: string
                replicas:
                  format: int64
                  type: integer
                traffic:
                  format: int64
                  type: integer
              type: object
            conditions:
              description: Conditions the latest available observations of a resource's
                current state. +patchMergeKey=type +patchStrategy=merge
              items:
                properties:
                  lastTransitionTime:
                    description: LastTransitionTime is the last time the condition
                      transitioned from one status to another. We use VolatileTime
                      in place of metav1.Time to exclude this from creating equality.Semantic
                      differences (all other things held constant).
                    type: string
                  message:
                    description: A human readable message indicating details about
                      the transition.
                    type: string
                  reason:
                    description: The reason for the condition's last transition.
                    type: string
                  severity:
                    description: Severity with which to treat failures of this type
                      of condition. When this is not specified, it defaults to Error.
                    type: string
                  status:
                    description: Status of the condition, one of True, False, Unknown.
                      +required
                    type: string
                  type:
                    description: Type of condition. +required
                    type: string
                required:
                - type
                - status
                type: object
              type: array
            default:
              properties:
                name:
                  type: string
                replicas:
                  format: int64
                  type: integer
                traffic:
                  format: int64
                  type: integer
              type: object
            observedGeneration:
              description: ObservedGeneration is the 'Generation' of the Service that
                was last processed by the controller.
              format: int64
              type: integer
            url:
              type: string
          type: object
  version: v1alpha1
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)
	th.writeK("/manifests/kfserving/kfserving-crds/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- crd.yaml
`)
}

func TestKfservingCrdsOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/kfserving/kfserving-crds/overlays/application")
	writeKfservingCrdsOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../kfserving/kfserving-crds/overlays/application"
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
