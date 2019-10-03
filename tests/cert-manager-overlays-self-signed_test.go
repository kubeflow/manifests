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

func writeCertManagerOverlaysSelfSigned(th *KustTestHarness) {
	th.writeF("/manifests/cert-manager/cert-manager/base/namespace.yaml", `
---
apiVersion: v1
kind: Namespace
metadata:
  name: $(namespace)
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-leaderelection
  labels:
    app: cert-manager
    app.kubernetes.io/name: cert-manager
    app.kubernetes.io/instance:  cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-leaderelection
subjects:
  - name: cert-manager
    namespace: $(namespace)
    kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-issuers
  labels:
    app: cert-manager
    app.kubernetes.io/name: cert-manager
    app.kubernetes.io/instance:  cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-issuers
subjects:
  - name: cert-manager
    namespace: $(namespace)
    kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-clusterissuers
  labels:
    app: cert-manager
    app.kubernetes.io/name: cert-manager
    app.kubernetes.io/instance:  cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-clusterissuers
subjects:
  - name: cert-manager
    namespace: $(namespace)
    kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-certificates
  labels:
    app: cert-manager
    app.kubernetes.io/name: cert-manager
    app.kubernetes.io/instance:  cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-certificates
subjects:
  - name: cert-manager
    namespace: $(namespace)
    kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-orders
  labels:
    app: cert-manager
    app.kubernetes.io/name: cert-manager
    app.kubernetes.io/instance:  cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-orders
subjects:
  - name: cert-manager
    namespace: $(namespace)
    kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-challenges
  labels:
    app: cert-manager
    app.kubernetes.io/name: cert-manager
    app.kubernetes.io/instance:  cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-challenges
subjects:
  - name: cert-manager
    namespace: $(namespace)
    kind: ServiceAccount

---

apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-controller-ingress-shim
  labels:
    app: cert-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-controller-ingress-shim
subjects:
  - name: cert-manager
    namespace: $(namespace)
    kind: ServiceAccount
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-leaderelection
  labels:
    app: cert-manager
rules:
  # Used for leader election by the controller
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "create", "update", "patch"]

---

# Issuer controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-issuers
  labels:
    app: cert-manager
rules:
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["issuers", "issuers/status"]
    verbs: ["update"]
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["issuers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

# ClusterIssuer controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-clusterissuers
  labels:
    app: cert-manager
rules:
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["clusterissuers", "clusterissuers/status"]
    verbs: ["update"]
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["clusterissuers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

# Certificates controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-certificates
  labels:
    app: cert-manager
rules:
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["certificates", "certificates/status", "certificaterequests", "certificaterequests/status"]
    verbs: ["update"]
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["certificates", "certificaterequests", "clusterissuers", "issuers", "orders"]
    verbs: ["get", "list", "watch"]
  # We require these rules to support users with the OwnerReferencesPermissionEnforcement
  # admission controller enabled:
  # https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["certificates/finalizers"]
    verbs: ["update"]
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["orders"]
    verbs: ["create", "delete"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

# Orders controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-orders
  labels:
    app: cert-manager
rules:
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["orders", "orders/status"]
    verbs: ["update"]
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["orders", "clusterissuers", "issuers", "challenges"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["challenges"]
    verbs: ["create", "delete"]
  # We require these rules to support users with the OwnerReferencesPermissionEnforcement
  # admission controller enabled:
  # https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["orders/finalizers"]
    verbs: ["update"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

# Challenges controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-challenges
  labels:
    app: cert-manager
rules:
  # Use to update challenge resource status
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["challenges", "challenges/status"]
    verbs: ["update"]
  # Used to watch challenges, issuer and clusterissuer resources
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["challenges", "issuers", "clusterissuers"]
    verbs: ["get", "list", "watch"]
  # Need to be able to retrieve ACME account private key to complete challenges
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
  # Used to create events
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]
  # HTTP01 rules
  - apiGroups: [""]
    resources: ["pods", "services"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: ["networking.k8s.io/v1"]
    resources: ["ingresses"]
    verbs: ["get", "list", "watch", "create", "delete", "update"]
  # We require these rules to support users with the OwnerReferencesPermissionEnforcement
  # admission controller enabled:
  # https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["challenges/finalizers"]
    verbs: ["update"]
  # DNS01 rules (duplicated above)
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]

---

# ingress-shim controller role
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-controller-ingress-shim
  labels:
    app: cert-manager
rules:
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["certificates", "certificaterequests"]
    verbs: ["create", "update", "delete"]
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["certificates", "certificaterequests", "issuers", "clusterissuers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["networking.k8s.io/v1"]
    resources: ["ingresses"]
    verbs: ["get", "list", "watch"]
  # We require these rules to support users with the OwnerReferencesPermissionEnforcement
  # admission controller enabled:
  # https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
  - apiGroups: ["networking.k8s.io/v1"]
    resources: ["ingresses/finalizers"]
    verbs: ["update"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cert-manager-view
  labels:
    app: cert-manager
    rbac.authorization.k8s.io/aggregate-to-view: "true"
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
rules:
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["certificates", "certificaterequests", "issuers"]
    verbs: ["get", "list", "watch"]

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cert-manager-edit
  labels:
    app: cert-manager
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
rules:
  - apiGroups: ["certmanager.k8s.io"]
    resources: ["certificates", "certificaterequests", "issuers"]
    verbs: ["create", "delete", "deletecollection", "patch", "update"]
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cert-manager
  labels:
    app: cert-manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cert-manager

  template:
    metadata:
      labels:
        app: cert-manager
      annotations:
        prometheus.io/path: "/metrics"
        prometheus.io/scrape: 'true'
        prometheus.io/port: '9402'
    spec:
      serviceAccountName: cert-manager
      containers:
        - name: cert-manager
          image: "quay.io/jetstack/cert-manager-controller:v0.10.0"
          imagePullPolicy: IfNotPresent
          args:
          - --v=2
          - --cluster-resource-namespace=$(POD_NAMESPACE)
          - --leader-election-namespace=$(POD_NAMESPACE)
          - --webhook-namespace=$(POD_NAMESPACE)
          - --webhook-ca-secret=cert-manager-webhook-ca
          - --webhook-serving-secret=cert-manager-webhook-tls
          - --webhook-dns-names=cert-manager-webhook,cert-manager-webhook.$(namespace),cert-manager-webhook.$(namespace).svc
          ports:
          - containerPort: 9402
          env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          resources:
            requests:
              cpu: 10m
              memory: 32Mi
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cert-manager
  labels:
    app: cert-manager
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: cert-manager
  labels:
    app: cert-manager
spec:
  type: ClusterIP
  ports:
    - protocol: TCP
      port: 9402
      targetPort: 9402
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/params.yaml", `
varReference:
- path: subjects/namespace
  kind: ClusterRoleBinding
- path: spec/template/spec/containers/args
  kind: Deployment
- path: metadata/name
  kind: Namespace
`)
	th.writeF("/manifests/cert-manager/cert-manager/base/params.env", `
namespace=cert-manager
`)
	th.writeK("/manifests/cert-manager/cert-manager/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: cert-manager
resources:
- namespace.yaml
- cluster-role-binding.yaml
- cluster-role.yaml
- deployment.yaml
- service-account.yaml
- service.yaml
commonLabels:
  kustomize.component: cert-manager
configMapGenerator:
- name: cert-manager-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: cert-manager-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
configurations:
- params.yaml
`)
	th.writeF("/manifests/cert-manager/cert-manager/overlays/self-signed/cluster-issuer.yaml", `
---
apiVersion: certmanager.k8s.io/v1alpha1
kind: ClusterIssuer
metadata:
  name: kubeflow-self-signing-issuer
spec:
  selfSigned: {}
`)
	th.writeK("/manifests/cert-manager/cert-manager/overlays/self-signed", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- cluster-issuer.yaml
commonLabels:
  kustomize.component: cert-manager
`)
}

func TestCertManagerOverlaysSelfSigned(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/cert-manager/cert-manager/overlays/self-signed")
	writeCertManagerOverlaysSelfSigned(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../cert-manager/cert-manager/overlays/self-signed"
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
