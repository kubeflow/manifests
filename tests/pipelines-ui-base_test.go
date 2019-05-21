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

func writePipelinesUiBase(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/pipelines-ui/base/deployment.yaml", `
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - name: deployment
        image: gcr.io/ml-pipeline/frontend:0.1.18
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 3000
      serviceAccountName: service-account
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: role
subjects:
- kind: ServiceAccount
  name: service-account
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: role
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  verbs:
  - create
  - get
  - list
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/service.yaml", `
---
apiVersion: v1
kind: Service
metadata:
  name: service
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: ml-pipeline-ui-service-mapping
      prefix: /pipeline
      rewrite: /pipeline
      timeout_ms: 300000
      service: $(service).$(ui-namespace)
      use_websocket: true
spec:
  ports:
  - port: 80
    targetPort: 3000
---
apiVersion: v1
kind: Service
metadata:
  name: tensorboard-service
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: ml-pipeline-ui-tensorboard-service-mapping
      prefix: /data
      rewrite: /data
      timeout_ms: 300000
      service: $(tensorboard-service).$(ui-namespace)
      use_websocket: true
spec:
  ports:
  - port: 80
    targetPort: 3000
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: ml-pipeline-ui
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /pipeline
    rewrite:
      uri: /pipeline
    route:
    - destination:
        host: ml-pipeline-ui.$(ui-namespace).svc.$(ui-clusterDomain)
        port:
          number: 80
    timeout: 300s
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeF("/manifests/pipeline/pipelines-ui/base/params.env", `
uiClusterDomain=cluster.local
`)
	th.writeK("/manifests/pipeline/pipelines-ui/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
nameprefix: ml-pipeline-ui-
commonLabels:
  app: ml-pipeline-ui
resources:
- deployment.yaml
- role-binding.yaml
- role.yaml
- service-account.yaml
- service.yaml
- virtual-service.yaml
configMapGenerator:
- name: ui-parameters
  env: params.env
images:
- name: gcr.io/ml-pipeline/frontend
  newTag: '0.1.18'
vars:
- name: ui-namespace
  objref:
    kind: Service
    name: service
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
- name: ui-clusterDomain
  objref:
    kind: ConfigMap
    name: ui-parameters
    version: v1
  fieldref:
    fieldpath: data.uiClusterDomain
- name: service
  objref:
    kind: Service
    name: service
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: tensorboard-service
  objref:
    kind: Service
    name: tensorboard-service
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
configurations:
- params.yaml
`)
}

func TestPipelinesUiBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/pipelines-ui/base")
	writePipelinesUiBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/pipelines-ui/base"
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
