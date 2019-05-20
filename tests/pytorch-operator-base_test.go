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

func writePytorchOperatorBase(th *KustTestHarness) {
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: pytorch-operator
  name: pytorch-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: pytorch-operator
subjects:
- kind: ServiceAccount
  name: pytorch-operator
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: pytorch-operator
  name: pytorch-operator
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - pytorchjobs
  - pytorchjobs/status
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - '*'
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - configmaps
  - pods
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  verbs:
  - '*'
- apiGroups:
  - apps
  - extensions
  resources:
  - deployments
  verbs:
  - '*'
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/config-map.yaml", `
apiVersion: v1
data:
  controller_config_file.yaml: |-
    {

    }
kind: ConfigMap
metadata:
  name: pytorch-operator-config
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: pytorchjobs.kubeflow.org
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
    kind: PyTorchJob
    plural: pytorchjobs
    singular: pytorchjob
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            pytorchReplicaSpecs:
              properties:
                Master:
                  properties:
                    replicas:
                      maximum: 1
                      minimum: 1
                      type: integer
                Worker:
                  properties:
                    replicas:
                      minimum: 1
                      type: integer
  version: v1beta2
  versions:
  - name: v1beta2
    served: true
    storage: true
  - name: v1beta1
    served: true
    storage: false
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: pytorch-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: pytorch-operator
    spec:
      containers:
      - command:
        - /pytorch-operator.v1beta2
        - --alsologtostderr
        - -v=1
        env:
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        image: gcr.io/kubeflow-images-public/pytorch-operator:v0.5.0
        name: pytorch-operator
        volumeMounts:
        - mountPath: /etc/config
          name: config-volume
      serviceAccountName: pytorch-operator
      volumes:
      - configMap:
          name: pytorch-operator-config
        name: config-volume
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: pytorch-operator
  name: pytorch-operator
`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/params.env", `
pytorchDefaultImage=null
deploymentScope=cluster
deploymentNamespace=null
`)
	th.writeK("/manifests/pytorch-job/pytorch-operator/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- crd.yaml
- deployment.yaml
- service-account.yaml
commonLabels:
  kustomize.component: pytorch-operator
images:
  - name: gcr.io/kubeflow-images-public/pytorch-operator
    newName: gcr.io/kubeflow-images-public/pytorch-operator
    newTag: v0.5.0
`)
}

func TestPytorchOperatorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pytorch-job/pytorch-operator/base")
	writePytorchOperatorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pytorch-job/pytorch-operator/base"
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
