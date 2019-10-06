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

func writeKfservingInstallOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/kfserving/kfserving-install/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: "kfserving"
spec:
  type: "kfserving"
  componentKinds:
    - group: apps/v1
      kind: StatefulSet
    - group: v1
      kind: Service
    - group: apps/v1
      kind: Deployment
    - group: v1
      kind: Secret
    - group: v1
      kind: ConfigMap
  version: "v1alpha1"
  description: "KFServing provides a Kubernetes Custom Resource Definition for serving ML Models on arbitrary frameworks"
  icons:
  maintainers:
    - name: Johnu George
      email: johnugeo@cisco.com
  owners:
    - name: Johnu George
      email: johnugeo@cisco.com
  keywords:
   - "kfserving"
   - "inference"
  links:
    - description: About
      url: "https://github.com/kubeflow/kfserving"
`)
	th.writeK("/manifests/kfserving/kfserving-install/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
commonLabels:
  app.kubernetes.io/name: kfserving  
  app.kubernetes.io/instance: kfserving
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: serving
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kfserving-proxy-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: proxy-role
subjects:
- kind: ServiceAccount
  name: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  creationTimestamp: null
  name: manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: manager-role
subjects:
- kind: ServiceAccount
  name: default
---
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kfserving-proxy-role
rules:
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: manager-role
rules:
- apiGroups:
  - serving.knative.dev
  resources:
  - configurations
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - serving.knative.dev
  resources:
  - configurations/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - serving.knative.dev
  resources:
  - routes
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - serving.knative.dev
  resources:
  - routes/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - serving.kubeflow.org
  resources:
  - kfservices
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - serving.kubeflow.org
  resources:
  - kfservices/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  - validatingwebhookconfigurations
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-kfserving-admin
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-kfserving-admin: "true"
rules: null

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-kfserving-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-kfserving-admin: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - kfservices
  verbs:
  - get
  - list
  - watch
  - create
  - delete
  - deletecollection
  - patch
  - update

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-kfserving-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules:
- apiGroups:
  - kubeflow.org
  resources:
  - kfservices
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/config-map.yaml", `
apiVersion: v1
data:
  credentials: |-
    {
       "gcs" : {
           "gcsCredentialFileName": "gcloud-application-credentials.json"
       },
       "s3" : {
           "s3AccessKeyIDName": "awsAccessKeyID",
           "s3SecretAccessKeyName": "awsSecretAccessKey"
       }
    }
  frameworks: |-
    {
        "tensorflow": {
            "image": "tensorflow/serving"
        },
        "sklearn": {
            "image": "$(registry)/sklearnserver"
        },
        "pytorch": {
            "image": "$(registry)/pytorchserver"
        },
        "xgboost": {
            "image": "$(registry)/xgbserver"
        },
        "tensorrt": {
            "image": "nvcr.io/nvidia/tensorrtserver"
        }
    }
kind: ConfigMap
metadata:
  name: kfservice-config
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/secret.yaml", `
apiVersion: v1
kind: Secret
metadata:
  name: kfserving-webhook-server-secret
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/statefulset.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    control-plane: kfserving-controller-manager
    controller-tools.k8s.io: "1.0"
  name: kfserving-controller-manager
spec:
  selector:
    matchLabels:
      control-plane: kfserving-controller-manager
      controller-tools.k8s.io: "1.0"
  serviceName: controller-manager-service
  template:
    metadata:
      labels:
        control-plane: kfserving-controller-manager
        controller-tools.k8s.io: "1.0"
    spec:
      containers:
      - args:
        - --secure-listen-address=0.0.0.0:8443
        - --upstream=http://127.0.0.1:8080/
        - --logtostderr=true
        - --v=10
        image: gcr.io/kubebuilder/kube-rbac-proxy:v0.4.0
        name: kube-rbac-proxy
        ports:
        - containerPort: 8443
          name: https
      - args:
        - --metrics-addr=127.0.0.1:8080
        command:
        - /manager
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: SECRET_NAME
          value: kfserving-webhook-server-secret
        image: $(registry)/kfserving-controller:v0.1.1
        imagePullPolicy: Always
        name: manager
        ports:
        - containerPort: 9876
          name: webhook-server
          protocol: TCP
        resources:
          limits:
            cpu: 100m
            memory: 300Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - mountPath: /tmp/cert
          name: cert
          readOnly: true
      terminationGracePeriodSeconds: 10
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: kfserving-webhook-server-secret
  volumeClaimTemplates: []
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/port: "8443"
    prometheus.io/scheme: https
    prometheus.io/scrape: "true"
  labels:
    control-plane: controller-manager
    controller-tools.k8s.io: "1.0"
  name: kfserving-controller-manager-metrics-service
spec:
  ports:
  - name: https
    port: 8443
    targetPort: https
  selector:
    control-plane: controller-manager
    controller-tools.k8s.io: "1.0"
---
apiVersion: v1
kind: Service
metadata:
  labels:
    control-plane: kfserving-controller-manager
    controller-tools.k8s.io: "1.0"
  name: kfserving-controller-manager-service
spec:
  ports:
  - port: 443
  selector:
    control-plane: kfserving-controller-manager
    controller-tools.k8s.io: "1.0"
---
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/image
  kind: StatefulSet
- path: data/frameworks
  kind: ConfigMap
`)
	th.writeF("/manifests/kfserving/kfserving-install/base/params.env", `
registry=gcr.io/kfserving
`)
	th.writeK("/manifests/kfserving/kfserving-install/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- secret.yaml
- statefulset.yaml
- service.yaml
commonLabels:
  kustomize.component: kfserving
configMapGenerator:
  - name: kfserving-parameters
    env: params.env
vars:
  - name: registry
    objref:
      kind: ConfigMap
      name: kfserving-parameters
      apiVersion: v1
    fieldref:
      fieldpath: data.registry
configurations:
- params.yaml
`)
}

func TestKfservingInstallOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/kfserving/kfserving-install/overlays/application")
	writeKfservingInstallOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../kfserving/kfserving-install/overlays/application"
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
