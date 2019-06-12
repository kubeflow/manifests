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

func writeKatibV1Alpha2ControllerBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/katib-controller/base/experiment-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: experiments.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: Status
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  names:
    categories:
    - all
    - kubeflow
    - katib
    kind: Experiment
    plural: experiments
    singular: experiment
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha2
`)
	th.writeF("/manifests/katib-v1alpha2/katib-controller/base/katib-controller-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: katib-controller
  name: katib-controller
  namespace: kubeflow
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib-controller
  template:
    metadata:
      labels:
        app: katib-controller
    spec:
      containers:
      - command:
        - ./katib-controller
        env:
        - name: KATIB_CORE_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-controller:v0.1.2-alpha-280-gb0e0dd5
        imagePullPolicy: IfNotPresent
        name: katib-controller
        ports:
        - containerPort: 443
          name: webhook
          protocol: TCP
        volumeMounts:
        - mountPath: /tmp/cert
          name: cert
          readOnly: true
      serviceAccountName: katib-controller
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: katib-controller
`)
	th.writeF("/manifests/katib-v1alpha2/katib-controller/base/katib-controller-rbac.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: katib-controller
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - serviceaccounts
  - services
  - secrets
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - pods/status
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - jobs
  - cronjobs
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - validatingwebhookconfigurations
  - mutatingwebhookconfigurations
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - experiments
  - experiments/status
  - trials
  - trials/status
  verbs:
  - '*'
- apiGroups:
  - kubeflow.org
  resources:
  - tfjobs
  - pytorchjobs
  verbs:
  - '*'
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: katib-controller
  namespace: kubeflow
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: katib-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: katib-controller
subjects:
- kind: ServiceAccount
  name: katib-controller
  namespace: kubeflow
`)
	th.writeF("/manifests/katib-v1alpha2/katib-controller/base/katib-controller-secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: katib-controller
  namespace: kubeflow
`)
	th.writeF("/manifests/katib-v1alpha2/katib-controller/base/katib-controller-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-controller
  namespace: kubeflow
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    app: katib-controller
`)
	th.writeF("/manifests/katib-v1alpha2/katib-controller/base/trial-crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: trials.kubeflow.org
spec:
  additionalPrinterColumns:
  - JSONPath: .status.conditions[-1:].type
    name: Status
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  group: kubeflow.org
  names:
    categories:
    - all
    - kubeflow
    - katib
    kind: Trial
    plural: trials
    singular: trial
  scope: Namespaced
  subresources:
    status: {}
  version: v1alpha2
`)
	th.writeF("/manifests/katib-v1alpha2/katib-controller/base/trial-template.yaml", `
apiVersion: v1
data:
  defaultTrialTemplate.yaml: |-
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: {{.Trial}}
      namespace: {{.NameSpace}}
    spec:
      template:
        spec:
          containers:
          - name: {{.Trial}}
            image: alpine
          restartPolicy: Never
kind: ConfigMap
metadata:
  name: trial-template
  namespace: kubeflow
`)
	th.writeK("/manifests/katib-v1alpha2/katib-controller/base", `
namespace: kubeflow
resources:
- experiment-crd.yaml
- katib-controller-deployment.yaml
- katib-controller-rbac.yaml
- katib-controller-secret.yaml
- katib-controller-service.yaml
- trial-crd.yaml
- trial-template.yaml
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-controller
    newTag: v0.1.2-alpha-280-gb0e0dd5
`)
}

func TestKatibV1Alpha2ControllerBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/katib-controller/base")
	writeKatibV1Alpha2ControllerBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/katib-controller/base"
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
