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

func writeWebhookOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/admission-webhook/webhook/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: $(generateName)
  componentKinds:
  - group: apps
    kind: Deployment
  - group: admissionregistration.k8s.io
    kind: MutatingWebhookConfiguration
  - group: core
    kind: ServiceAccount
  - group: core
    kind: Service
  descriptor:
    type: "admission-webhook-webhook"
    version: "v1alpha1"
    description: "admission-webhook webhook injects common data (env vars, volumes) into notebooks"
    keywords:
    - "admission-webook"
    links:
    - description: About
      url: "https://github.com/kubeflow/kubeflow/tree/master/components/admission-webhook"
`)
	th.writeF("/manifests/admission-webhook/webhook/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Application
- path: spec/selector/app.kubernetes.io\/instance
  kind: Service
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/template/metadata/labels/app.kubernetes.io\/instance
  kind: Deployment
`)
	th.writeF("/manifests/admission-webhook/webhook/overlays/application/params.env", `
generateName=
`)
	th.writeK("/manifests/admission-webhook/webhook/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: admission-webhook-webhook-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: admission-webhook-webhook-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: admission-webhook-webhook
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: admission-webhook
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/admission-webhook/webhook/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-role
subjects:
- kind: ServiceAccount
  name: service-account
`)
	th.writeF("/manifests/admission-webhook/webhook/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-role
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - poddefaults
  verbs:
  - get
  - watch
  - list
  - update
  - create
  - patch
  - delete
`)
	th.writeF("/manifests/admission-webhook/webhook/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - image: gcr.io/kubeflow-images-public/admission-webhook:v20190520-v0-139-gcee39dbc-dirty-0d8f4c
        name: admission-webhook
        volumeMounts:
        - mountPath: /etc/webhook/certs
          name: webhook-cert
          readOnly: true
      volumes:
      - name: webhook-cert
        secret:
          secretName: webhook-certs
      serviceAccountName: service-account    
`)
	th.writeF("/manifests/admission-webhook/webhook/base/mutating-webhook-configuration.yaml", `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: mutating-webhook-configuration
webhooks:
- clientConfig:
    caBundle: ""
    service:
      name: $(serviceName)
      namespace: $(namespace)
      path: /apply-poddefault
  name: $(deploymentName).kubeflow.org
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
`)
	th.writeF("/manifests/admission-webhook/webhook/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
`)
	th.writeF("/manifests/admission-webhook/webhook/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: service
spec:
  ports:
  - port: 443
    targetPort: 443
`)
	th.writeF("/manifests/admission-webhook/webhook/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: poddefaults.kubeflow.org
spec:
  group: kubeflow.org
  names:
    kind: PodDefault
    plural: poddefaults
    singular: poddefault
  scope: Namespaced
  version: v1alpha1
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          properties:
            desc:
              type: string
            serviceAccountName:
              type: string
            env:
              items:
                type: object
              type: array
            envFrom:
              items:
                type: object
              type: array
            selector:
              type: object
            volumeMounts:
              items:
                type: object
              type: array
            volumes:
              items:
                type: object
              type: array
          required:
          - selector
          type: object
        status:
          type: object
      type: object
`)
	th.writeF("/manifests/admission-webhook/webhook/base/params.yaml", `
varReference:
- path: webhooks/clientConfig/service/namespace
  kind: MutatingWebhookConfiguration
- path: webhooks/clientConfig/service/name
  kind: MutatingWebhookConfiguration
- path: webhooks/name
  kind: MutatingWebhookConfiguration
`)
	th.writeF("/manifests/admission-webhook/webhook/base/params.env", `
namespace=kubeflow
`)
	th.writeK("/manifests/admission-webhook/webhook/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- mutating-webhook-configuration.yaml
- service-account.yaml
- service.yaml
- crd.yaml
commonLabels:
  kustomize.component: admission-webhook
  app: admission-webhook
namePrefix: admission-webhook- 
images:
  - name: gcr.io/kubeflow-images-public/admission-webhook
    newName: gcr.io/kubeflow-images-public/admission-webhook
    newTag: v20190520-v0-139-gcee39dbc-dirty-0d8f4c
namespace: kubeflow  
configMapGenerator:
- name: admission-webhook-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: admission-webhook-parameters 
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace	
- name: serviceName
  objref:
    kind: Service
    name: service
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: deploymentName
  objref:
    kind: Deployment
    name: deployment
    apiVersion: apps/v1
  fieldref:
    fieldpath: metadata.name
configurations:
- params.yaml
`)
}

func TestWebhookOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/admission-webhook/webhook/overlays/application")
	writeWebhookOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../admission-webhook/webhook/overlays/application"
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
