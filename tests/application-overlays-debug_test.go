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

func writeApplicationOverlaysDebug(th *KustTestHarness) {
	th.writeF("/manifests/application/overlays/debug/stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: stateful-set
spec:
  template:
    spec:
      containers:
      - name: manager
        command: ["/go/bin/dlv"]
        args: ["--listen=:2345", "--headless=true", "--api-version=2", "exec", "/go/src/github.com/kubernetes-sigs/application/manager"]
        image: gcr.io/$(project)/application-controller:latest
        env:
        - name: project
          valueFrom:
            configMapKeyRef:
              name: application-controller-parameters
              key: project
        ports:
        - containerPort: 2345
        securityContext:
          privileged: true
`)
	th.writeF("/manifests/application/overlays/debug/params.yaml", `
varReference:
- path: spec/template/spec/containers/image
  kind: StatefulSet
`)
	th.writeF("/manifests/application/overlays/debug/params.env", `
project=
`)
	th.writeK("/manifests/application/overlays/debug", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
patchesStrategicMerge:
- stateful-set.yaml
configMapGenerator:
- name: application-controller-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: project
  objref:
    kind: ConfigMap
    name: application-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.project
configurations:
- params.yaml
`)
	th.writeF("/manifests/application/base/crd.yaml", `
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
	th.writeF("/manifests/application/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: controller-cluster-role
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
	th.writeF("/manifests/application/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: controller-cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: controller-cluster-role
subjects:
- kind: ServiceAccount
  name: controller-service-account
`)
	th.writeF("/manifests/application/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: controller-service-account
`)
	th.writeF("/manifests/application/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: controller-service
spec:
  ports:
  - port: 443
`)
	th.writeF("/manifests/application/base/stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: stateful-set
spec:
  serviceName: controller-service
  template:
    spec:
      containers:
      - name: manager
        command:
        - /root/manager
        image: controller:latest
        imagePullPolicy: Always
        env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
      serviceAccountName: controller-service-account
  volumeClaimTemplates: []
`)
	th.writeK("/manifests/application/base", `
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
commonLabels:
  kustomize-component: application-controller
`)
}

func TestApplicationOverlaysDebug(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/application/overlays/debug")
	writeApplicationOverlaysDebug(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../application/overlays/debug"
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
