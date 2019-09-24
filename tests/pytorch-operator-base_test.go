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
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pytorch-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      name: pytorch-operator
  template:
    metadata:
      labels:
        name: pytorch-operator
    spec:
      containers:
      - command:
        - /pytorch-operator.v1
        - --alsologtostderr
        - -v=1
        - --monitoring-port=8443
        env:
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        image: gcr.io/kubeflow-images-public/pytorch-operator:v0.5.1-5-ge775742
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
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "8443"
    prometheus.io/scrape: "true"
  labels:
    app: pytorch-operator
  name: pytorch-operator
spec:
  ports:
  - name: monitoring-port
    port: 8443
    targetPort: 8443
  selector:
    name: pytorch-operator
  type: ClusterIP

`)
	th.writeF("/manifests/pytorch-job/pytorch-operator/base/params.env", `
pytorchDefaultImage=null
deploymentScope=cluster
deploymentNamespace=null
`)
	th.writeK("/manifests/pytorch-job/pytorch-operator/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- deployment.yaml
- service-account.yaml
- service.yaml
commonLabels:
  kustomize.component: pytorch-operator
images:
  - name: gcr.io/kubeflow-images-public/pytorch-operator
    newName: gcr.io/kubeflow-images-public/pytorch-operator
    newTag: v1.0.0-rc.0
`)
}

func TestPytorchOperatorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pytorch-job/pytorch-operator/base")
	writePytorchOperatorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pytorch-job/pytorch-operator/base"
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
