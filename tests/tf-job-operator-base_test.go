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

func writeTfJobOperatorBase(th *KustTestHarness) {
	th.writeF("/manifests/tf-training/tf-job-operator/base/cluster-role-binding.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: tf-job-dashboard
  name: tf-job-dashboard
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tf-job-dashboard
subjects:
- kind: ServiceAccount
  name: tf-job-dashboard
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: tf-job-operator
  name: tf-job-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tf-job-operator
subjects:
- kind: ServiceAccount
  name: tf-job-operator
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/cluster-role.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: tf-job-dashboard
  name: tf-job-dashboard
rules:
- apiGroups:
  - tensorflow.org
  - kubeflow.org
  resources:
  - tfjobs
  - tfjobs/status
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
  - pods/log
  - namespaces
  verbs:
  - '*'
- apiGroups:
  - apps
  - extensions
  resources:
  - deployments
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: tf-job-operator
  name: tf-job-operator
rules:
- apiGroups:
  - tensorflow.org
  - kubeflow.org
  resources:
  - tfjobs
  - tfjobs/status
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
	th.writeF("/manifests/tf-training/tf-job-operator/base/config-map.yaml", `
apiVersion: v1
data:
  controller_config_file.yaml: |-
    {
        "grpcServerFilePath": "/opt/mlkube/grpc_tensorflow_server/grpc_tensorflow_server.py"
    }
kind: ConfigMap
metadata:
  name: tf-job-operator-config
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tf-job-dashboard
spec:
  template:
    metadata:
      labels:
        name: tf-job-dashboard
    spec:
      containers:
      - command:
        - /opt/tensorflow_k8s/dashboard/backend
        env:
        - name: KUBEFLOW_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/kubeflow-images-public/tf_operator:kubeflow-tf-operator-postsubmit-v1-5adee6f-6109-a25c
        name: tf-job-dashboard
        ports:
        - containerPort: 8080
      serviceAccountName: tf-job-dashboard
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tf-job-operator
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: tf-job-operator
    spec:
      containers:
      - command:
        - /opt/kubeflow/tf-operator.v1
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
        image: gcr.io/kubeflow-images-public/tf_operator:v0.6.0.rc0
        name: tf-job-operator
        volumeMounts:
        - mountPath: /etc/config
          name: config-volume
      serviceAccountName: tf-job-operator
      volumes:
      - configMap:
          name: tf-job-operator-config
        name: config-volume
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/service-account.yaml", `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: tf-job-dashboard
  name: tf-job-dashboard
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: tf-job-operator
  name: tf-job-operator
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: tfjobs-ui-mapping
      prefix: /tfjobs/
      rewrite: /tfjobs/
      service: tf-job-dashboard.$(namespace)
  name: tf-job-dashboard
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    name: tf-job-dashboard
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/scrape: "true"
    prometheus.io/port: "8443"
  labels:
    app: tf-job-operator
  name: tf-job-operator
spec:
  ports:
  - name: monitoring-port
    port: 8443
    targetPort: 8443
  selector:
    name: tf-job-operator
  type: ClusterIP
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/tf-training/tf-job-operator/base/params.env", `
namespace=
clusterDomain=cluster.local
`)
	th.writeK("/manifests/tf-training/tf-job-operator/base", `
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
  kustomize.component: tf-job-operator
configMapGenerator:
- name: parameters
  env: params.env
vars:
- name: namespace
  objref:
    kind: Service
    name: tf-job-dashboard
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
configurations:
- params.yaml
`)
}

func TestTfJobOperatorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/tf-training/tf-job-operator/base")
	writeTfJobOperatorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../tf-training/tf-job-operator/base"
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
