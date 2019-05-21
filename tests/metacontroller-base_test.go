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

func writeMetacontrollerBase(th *KustTestHarness) {
	th.writeF("/manifests/metacontroller/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: meta-controller-cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: meta-controller-service
`)
	th.writeF("/manifests/metacontroller/base/crd.yaml", `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
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
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: compositecontrollers.metacontroller.k8s.io
spec:
  group: metacontroller.k8s.io
  names:
    kind: CompositeController
    plural: compositecontrollers
    shortNames:
    - cc
    - cctl
    singular: compositecontroller
  scope: Cluster
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: controllerrevisions.metacontroller.k8s.io
spec:
  group: metacontroller.k8s.io
  names:
    kind: ControllerRevision
    plural: controllerrevisions
    singular: controllerrevision
  scope: Namespaced
  version: v1alpha1
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: decoratorcontrollers.metacontroller.k8s.io
spec:
  group: metacontroller.k8s.io
  names:
    kind: DecoratorController
    plural: decoratorcontrollers
    shortNames:
    - dec
    - decorators
    singular: decoratorcontroller
  scope: Cluster
  version: v1alpha1
`)
	th.writeF("/manifests/metacontroller/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: meta-controller-service
`)
	th.writeF("/manifests/metacontroller/base/stateful-set.yaml", `
apiVersion: apps/v1beta2
kind: StatefulSet
metadata:
  labels:
    app: metacontroller
  name: metacontroller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: metacontroller
  serviceName: ""
  template:
    metadata:
      labels:
        app: metacontroller
    spec:
      containers:
      - command:
        - /usr/bin/metacontroller
        - --logtostderr
        - -v=4
        - --discovery-interval=20s
        image: metacontroller/metacontroller:v0.3.0
        imagePullPolicy: Always
        name: metacontroller
        ports:
        - containerPort: 2345
        resources:
          limits:
            cpu: "4"
            memory: 4Gi
          requests:
            cpu: 500m
            memory: 1Gi
        securityContext:
          allowPrivilegeEscalation: true
          privileged: true
      serviceAccountName: meta-controller-service
  # Workaround for https://github.com/kubernetes-sigs/kustomize/issues/677
  volumeClaimTemplates: []
`)
	th.writeK("/manifests/metacontroller/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- crd.yaml
- service-account.yaml
- stateful-set.yaml
commonLabels:
  kustomize.component: metacontroller
images:
  - name: metacontroller/metacontroller
    newName: metacontroller/metacontroller
    newTag: v0.3.0
`)
}

func TestMetacontrollerBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/metacontroller/base")
	writeMetacontrollerBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../metacontroller/base"
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
