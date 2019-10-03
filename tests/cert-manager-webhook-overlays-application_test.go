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

func writeCertManagerWebhookOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/namespace.yaml", `
---
apiVersion: v1
kind: Namespace
metadata:
  name: $(namespace)
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/api-service.yaml", `
apiVersion: apiregistration.k8s.io/v1beta1
kind: APIService
metadata:
  name: v1beta1.webhook.certmanager.k8s.io
  labels:
    app: webhook
  annotations:
    certmanager.k8s.io/inject-ca-from-secret: "cert-manager/cert-manager-webhook-tls"
spec:
  group: webhook.certmanager.k8s.io
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: cert-manager-webhook
    namespace: $(namespace)
  version: v1beta1
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-webhook:auth-delegator
  labels:
    app: webhook
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- apiGroup: ""
  kind: ServiceAccount
  name: cert-manager-webhook
  namespace: $(namespace)
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cert-manager-webhook:webhook-requester
  labels:
    app: webhook
rules:
- apiGroups:
  - admission.certmanager.k8s.io
  resources:
  - certificates
  - certificaterequests
  - issuers
  - clusterissuers
  verbs:
  - create
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webhook

  template:
    metadata:
      labels:
        app: webhook
      annotations:
    spec:
      serviceAccountName: cert-manager-webhook
      containers:
        - name: cert-manager
          image: "quay.io/jetstack/cert-manager-webhook:v0.10.0"
          imagePullPolicy: IfNotPresent
          args:
          - --v=2
          - --secure-port=6443
          - --tls-cert-file=/certs/tls.crt
          - --tls-private-key-file=/certs/tls.key
          env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          resources:
            {}

          volumeMounts:
          - name: certs
            mountPath: /certs
      volumes:
      - name: certs
        secret:
          secretName: cert-manager-webhook-tls
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/mutating-webhook-configuration.yaml", `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
  annotations:
    certmanager.k8s.io/inject-apiserver-ca: "true"
webhooks:
  - name: webhook.certmanager.k8s.io
    rules:
      - apiGroups:
          - "certmanager.k8s.io"
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - certificates
          - issuers
          - clusterissuers
          - orders
          - challenges
          - certificaterequests
    failurePolicy: Fail
    clientConfig:
      service:
        name: kubernetes
        namespace: default
        path: /apis/webhook.certmanager.k8s.io/v1beta1/mutations
      caBundle: ""
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
spec:
  type: ClusterIP
  ports:
  - name: https
    port: 443
    targetPort: 6443
  selector:
    app: webhook
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/validating-webhook-configuration.yaml", `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: cert-manager-webhook
  labels:
    app: webhook
  annotations:
    certmanager.k8s.io/inject-apiserver-ca: "true"
webhooks:
  - name: webhook.certmanager.k8s.io
    rules:
      - apiGroups:
          - "certmanager.k8s.io"
        apiVersions:
          - v1alpha1
        operations:
          - CREATE
          - UPDATE
        resources:
          - certificates
          - issuers
          - clusterissuers
          - certificaterequests
    failurePolicy: Fail
    sideEffects: None
    clientConfig:
      service:
        name: kubernetes
        namespace: default
        path: /apis/webhook.certmanager.k8s.io/v1beta1/validations
      caBundle: ""
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/params.yaml", `
varReference:
- path: subjects/namespace
  kind: ClusterRoleBinding
- path: spec/service/namespace
  kind: APIService
- path: metadata/name
  kind: Namespace
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/base/params.env", `
namespace=cert-manager
`)
	th.writeK("/manifests/cert-manager/cert-manager-webhook/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: cert-manager
resources:
- namespace.yaml
- api-service.yaml
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- mutating-webhook-configuration.yaml
- service-account.yaml
- service.yaml
- validating-webhook-configuration.yaml
commonLabels:
  kustomize.component: cert-manager-webhook
configMapGenerator:
- name: cert-manager-webhook-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: cert-manager-webhook-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
configurations:
- params.yaml
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name:  $(generateName)
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: cert-manager-webhook
      app.kubernetes.io/instance: $(generateName)
      app.kubernetes.io/managed-by: kfctl
      app.kubernetes.io/component: cert-manager
      app.kubernetes.io/part-of: kubeflow
      app.kubernetes.io/version: v0.6
  componentKinds:
  - group: apiregistration
    kind: APIService
  - group: rbac
    kind: ClusterRole
  - group: rbac
    kind: ClusterRoleBinding
  - group: core
    kind: Namespace
  - group: core
    kind: Service
  - group: apps
    kind: Deployment
  - group: core
    kind: ServiceAccount
  - group: admissionregistration
    kind: MutatingWebhookConfiguration
  - group: admissionregistration
    kind: ValidatingWebhookConfiguration
  descriptor:
    type: ""
    version: "v0.10.0"
    description: "Automatically provision and manage TLS certificates in Kubernetes https://jetstack.io."
    keywords:
    - cert-manager
    - cert-manager-webhook
    links:
    - description: About
      url: "https://github.com/jetstack/cert-manager"
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/overlays/application/params.env", `
generateName=cert-manager-webhook
`)
	th.writeF("/manifests/cert-manager/cert-manager-webhook/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Application
- path: spec/template/metadata/labels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/selector/app.kubernetes.io\/instance
  kind: Service
`)
	th.writeK("/manifests/cert-manager/cert-manager-webhook/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: cert-manager-webhook-app-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: cert-manager-webhook-app-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: cert-manager-webhook
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: cert-manager
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
}

func TestCertManagerWebhookOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/cert-manager/cert-manager-webhook/overlays/application")
	writeCertManagerWebhookOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../cert-manager/cert-manager-webhook/overlays/application"
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
