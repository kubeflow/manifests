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

func writeSparkOperatorBase(th *KustTestHarness) {
	th.writeF("/manifests/spark-operator/base/spark-operator-spark-sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: spark-operator-spark
`)
	th.writeF("/manifests/spark-operator/base/spark-operator-sparkoperator-cr-clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: spark-operator-sparkoperator-cr
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - services
  - configmaps
  - secrets
  verbs:
  - create
  - get
  - delete
  - update
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - create
  - get
  - delete
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - update
  - patch
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
  - update
  - delete
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  verbs:
  - create
  - get
  - update
  - delete
- apiGroups:
  - sparkoperator.k8s.io
  resources:
  - sparkapplications
  - scheduledsparkapplications
  verbs:
  - '*'
`)
	th.writeF("/manifests/spark-operator/base/spark-operator-sparkoperator-crb-crb.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: spark-operator-sparkoperator-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: spark-operator-sparkoperator-cr
subjects:
- kind: ServiceAccount
  name: spark-operator-sparkoperator
`)
	th.writeF("/manifests/spark-operator/base/spark-operator-sparkoperator-crd-cleanup-job.yaml", `
apiVersion: batch/v1
kind: Job
metadata:
  name: spark-operator-sparkoperator-crd-cleanup
  namespace: default
spec:
  template:
    spec:
      containers:
      - command:
        - /bin/sh
        - -c
        - 'curl -ik -X DELETE -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
          -H "Accept: application/json" -H "Content-Type: application/json" https://kubernetes.default.svc/apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions/sparkapplications.sparkoperator.k8s.io'
        image: gcr.io/spark-operator/spark-operator:v2.4.0-v1beta1-0.8.2
        imagePullPolicy: IfNotPresent
        name: delete-sparkapp-crd
      - command:
        - /bin/sh
        - -c
        - 'curl -ik -X DELETE -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
          -H "Accept: application/json" -H "Content-Type: application/json" https://kubernetes.default.svc/apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions/scheduledsparkapplications.sparkoperator.k8s.io'
        image: gcr.io/spark-operator/spark-operator:v2.4.0-v1beta1-0.8.2
        imagePullPolicy: IfNotPresent
        name: delete-scheduledsparkapp-crd
      restartPolicy: OnFailure
      serviceAccountName: spark-operator-sparkoperator
`)
	th.writeF("/manifests/spark-operator/base/spark-operator-sparkoperator-deploy.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: spark-operator-sparkoperator
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: sparkoperator
      app.kubernetes.io/version: v2.4.0-v1beta1-0.8.2
  strategy:
    type: Recreate
  template:
    metadata:
      annotations:
        prometheus.io/path: /metrics
        prometheus.io/port: "10254"
        prometheus.io/scrape: "true"
      initializers:
        pending: []
      labels:
        app.kubernetes.io/name: sparkoperator
        app.kubernetes.io/version: v2.4.0-v1beta1-0.8.2
    spec:
      containers:
      - args:
        - -v=2
        - -namespace=
        - -ingress-url-format=
        - -install-crds=true
        - -controller-threads=10
        - -resync-interval=30
        - -logtostderr
        - -enable-metrics=true
        - -metrics-labels=app_type
        - -metrics-port=10254
        - -metrics-endpoint=/metrics
        - -metrics-prefix=
        image: gcr.io/spark-operator/spark-operator:v2.4.0-v1beta1-0.8.2
        imagePullPolicy: IfNotPresent
        name: sparkoperator
        ports:
        - containerPort: 10254
      serviceAccountName: spark-operator-sparkoperator
`)
	th.writeF("/manifests/spark-operator/base/spark-operator-sparkoperator-sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: spark-operator-sparkoperator
`)
	th.writeK("/manifests/spark-operator/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
# Labels to add to all resources and selectors.
commonLabels:
  kustomize.component: spark-operator
  app.kubernetes.io/instance: spark-operator
  app.kubernetes.io/managed-by: Tiller
  app.kubernetes.io/name: sparkoperator
  helm.sh/chart: sparkoperator-0.2.4

# Images modify the tags for images without
# creating patches.
images:
- name: gcr.io/spark-operator/spark-operator
  newTag: v2.4.0-v1beta1-0.8.2

# Value of this field is prepended to the
# names of all resources
namePrefix: spark-operator-spark

# List of resource files that kustomize reads, modifies
# and emits as a YAML string
resources:
- spark-operator-spark-sa.yaml
- spark-operator-sparkoperator-cr-clusterrole.yaml
- spark-operator-sparkoperator-crb-crb.yaml
- spark-operator-sparkoperator-crd-cleanup-job.yaml
- spark-operator-sparkoperator-deploy.yaml
- spark-operator-sparkoperator-sa.yaml
`)
}

func TestSparkOperatorBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/spark-operator/base")
	writeSparkOperatorBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../spark-operator/base"
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
