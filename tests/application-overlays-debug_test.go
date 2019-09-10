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

func writeApplicationOverlaysDebug(th *KustTestHarness) {
	th.writeF("/manifests/application/application/overlays/debug/stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: stateful-set
spec:
  template:
    spec:
      containers:
      - name: manager
        image: gcr.io/$(project)/application-controller:latest
        command: 
        - /go/bin/dlv
        args: 
        - --listen=:2345
        - --headless=true
        - --api-version=2
        - exec
        - /go/src/github.com/kubernetes-sigs/application/manager
        ports:
        - containerPort: 2345
        securityContext:
          privileged: true
`)
	th.writeK("/manifests/application/application/overlays/debug", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
patchesStrategicMerge:
- stateful-set.yaml
`)
	th.writeF("/manifests/application/application/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  name: applications.app.k8s.io
spec:
  group: app.k8s.io
  names:
    kind: Application
    plural: applications
  scope: Namespaced
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
            addOwnerRef:
              type: boolean
            assemblyPhase:
              type: string
            componentKinds:
              items:
                type: object
              type: array
            descriptor:
              properties:
                description:
                  type: string
                icons:
                  items:
                    properties:
                      size:
                        type: string
                      src:
                        type: string
                      type:
                        type: string
                    required:
                    - src
                    type: object
                  type: array
                keywords:
                  items:
                    type: string
                  type: array
                links:
                  items:
                    properties:
                      description:
                        type: string
                      url:
                        type: string
                    type: object
                  type: array
                maintainers:
                  items:
                    properties:
                      email:
                        type: string
                      name:
                        type: string
                      url:
                        type: string
                    type: object
                  type: array
                notes:
                  type: string
                owners:
                  items:
                    properties:
                      email:
                        type: string
                      name:
                        type: string
                      url:
                        type: string
                    type: object
                  type: array
                type:
                  type: string
                version:
                  type: string
              type: object
            info:
              items:
                properties:
                  name:
                    type: string
                  type:
                    type: string
                  value:
                    type: string
                  valueFrom:
                    properties:
                      configMapKeyRef:
                        properties:
                          apiVersion:
                            type: string
                          fieldPath:
                            type: string
                          key:
                            type: string
                          kind:
                            type: string
                          name:
                            type: string
                          namespace:
                            type: string
                          resourceVersion:
                            type: string
                          uid:
                            type: string
                        type: object
                      ingressRef:
                        properties:
                          apiVersion:
                            type: string
                          fieldPath:
                            type: string
                          host:
                            type: string
                          kind:
                            type: string
                          name:
                            type: string
                          namespace:
                            type: string
                          path:
                            type: string
                          resourceVersion:
                            type: string
                          uid:
                            type: string
                        type: object
                      secretKeyRef:
                        properties:
                          apiVersion:
                            type: string
                          fieldPath:
                            type: string
                          key:
                            type: string
                          kind:
                            type: string
                          name:
                            type: string
                          namespace:
                            type: string
                          resourceVersion:
                            type: string
                          uid:
                            type: string
                        type: object
                      serviceRef:
                        properties:
                          apiVersion:
                            type: string
                          fieldPath:
                            type: string
                          kind:
                            type: string
                          name:
                            type: string
                          namespace:
                            type: string
                          path:
                            type: string
                          port:
                            format: int32
                            type: integer
                          resourceVersion:
                            type: string
                          uid:
                            type: string
                        type: object
                      type:
                        type: string
                    type: object
                type: object
              type: array
            selector:
              type: object
          type: object
        status:
          properties:
            components:
              items:
                properties:
                  group:
                    type: string
                  kind:
                    type: string
                  link:
                    type: string
                  name:
                    type: string
                  status:
                    type: string
                type: object
              type: array
            conditions:
              items:
                properties:
                  lastTransitionTime:
                    format: date-time
                    type: string
                  lastUpdateTime:
                    format: date-time
                    type: string
                  message:
                    type: string
                  reason:
                    type: string
                  status:
                    type: string
                  type:
                    type: string
                required:
                - type
                - status
                type: object
              type: array
            observedGeneration:
              format: int64
              type: integer
          type: object
  version: v1beta1
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)
	th.writeF("/manifests/application/application/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-role
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - get
  - list
  - update
  - patch
  - watch
- apiGroups:
  - app.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
`)
	th.writeF("/manifests/application/application/base/cluster-role-binding.yaml", `
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
	th.writeF("/manifests/application/application/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account
`)
	th.writeF("/manifests/application/application/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: service
spec:
  ports:
  - port: 443
`)
	th.writeF("/manifests/application/application/base/stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: stateful-set
spec:
  serviceName: service
  selector:
    matchLabels:
      app: application-controller
  template:
    metadata:
      labels:
        app: application-controller
    spec:
      containers:
      - name: manager
        command:
        - /root/manager
        image: gcr.io/kubeflow-images-public/kubernetes-sigs/application
        imagePullPolicy: Always
        env:
        - name: project
          value: $(project)
      serviceAccountName: service-account
  volumeClaimTemplates: []
`)
	th.writeF("/manifests/application/application/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/image
  kind: StatefulSet
`)
	th.writeF("/manifests/application/application/base/params.env", `
project=
`)
	th.writeK("/manifests/application/application/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- crd.yaml
- cluster-role.yaml
- cluster-role-binding.yaml
- service-account.yaml
- service.yaml
- stateful-set.yaml
namespace: kubeflow
nameprefix: application-controller-
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/kubernetes-sigs/application
    newName: gcr.io/kubeflow-images-public/kubernetes-sigs/application
    newTag: 1.0-beta
vars:
- name: project
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.project
configurations:
- params.yaml
`)
}

func TestApplicationOverlaysDebug(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/application/application/overlays/debug")
	writeApplicationOverlaysDebug(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../application/application/overlays/debug"
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
