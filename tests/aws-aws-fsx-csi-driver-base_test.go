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

func writeAwsFsxCsiDriverBase(th *KustTestHarness) {
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-driver.yaml", `
---
apiVersion: storage.k8s.io/v1beta1
kind: CSIDriver
metadata:
  name: fsx.csi.aws.com
spec:
  attachRequired: false
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-controller.yaml", `
---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: fsx-csi-controller
spec:
  replicas: 2
  selector:
    matchLabels:
      app: fsx-csi-controller
  template:
    metadata:
      labels:
        app: fsx-csi-controller
    spec:
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/arch: amd64
      serviceAccount: fsx-csi-controller-sa
      tolerations:
        - key: CriticalAddonsOnly
          operator: Exists
      containers:
        - name: fsx-plugin
          image: amazon/aws-fsx-csi-driver:latest
          args :
            - --endpoint=$(CSI_ENDPOINT)
            - --logtostderr
            - --v=5
          env:
            - name: CSI_ENDPOINT
              value: unix:///var/lib/csi/sockets/pluginproxy/csi.sock
            - name: AWS_ACCESS_KEY_ID
              valueFrom:
                secretKeyRef:
                  name: aws-secret
                  key: AWS_ACCESS_KEY_ID
                  optional: true
            - name: AWS_SECRET_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: aws-secret
                  key: AWS_SECRET_ACCESS_KEY
                  optional: true
          volumeMounts:
            - name: socket-dir
              mountPath: /var/lib/csi/sockets/pluginproxy/
        - name: csi-provisioner
          image: quay.io/k8scsi/csi-provisioner:v1.3.0
          args:
            - --timeout=5m
            - --csi-address=$(ADDRESS)
            - --v=5
            - --enable-leader-election
            - --leader-election-type=leases
          env:
            - name: ADDRESS
              value: /var/lib/csi/sockets/pluginproxy/csi.sock
          volumeMounts:
            - name: socket-dir
              mountPath: /var/lib/csi/sockets/pluginproxy/
      volumes:
        - name: socket-dir
          emptyDir: {}
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-controller-sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fsx-csi-controller-sa
  namespace: kubeflow
  #Enable if EKS IAM for SA is used
  #annotations:
  #  eks.amazonaws.com/role-arn: arn:aws:iam::111122223333:role/fsx-csi-role
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-node-daemonset.yaml", `
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: fsx-csi-node
spec:
  selector:
    matchLabels:
      app: fsx-csi-node
  template:
    metadata:
      labels:
        app: fsx-csi-node
    spec:
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/arch: amd64
      hostNetwork: true
      containers:
        - name: fsx-plugin
          securityContext:
            privileged: true
          image: amazon/aws-fsx-csi-driver:latest
          args:
            - --endpoint=$(CSI_ENDPOINT)
            - --logtostderr
            - --v=5
          env:
            - name: CSI_ENDPOINT
              value: unix:/csi/csi.sock
          volumeMounts:
            - name: kubelet-dir
              mountPath: /var/lib/kubelet
              mountPropagation: "Bidirectional"
            - name: plugin-dir
              mountPath: /csi
          ports:
            - containerPort: 9810
              name: healthz
              protocol: TCP
          livenessProbe:
            failureThreshold: 5
            httpGet:
              path: /healthz
              port: healthz
            initialDelaySeconds: 10
            timeoutSeconds: 3
            periodSeconds: 2
        - name: csi-driver-registrar
          image: quay.io/k8scsi/csi-node-driver-registrar:v1.1.0
          args:
            - --csi-address=$(ADDRESS)
            - --kubelet-registration-path=$(DRIVER_REG_SOCK_PATH)
            - --v=5
          env:
            - name: ADDRESS
              value: /csi/csi.sock
            - name: DRIVER_REG_SOCK_PATH
              value: /var/lib/kubelet/plugins/fsx.csi.aws.com/csi.sock
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts:
            - name: plugin-dir
              mountPath: /csi
            - name: registration-dir
              mountPath: /registration
        - name: liveness-probe
          imagePullPolicy: Always
          image: quay.io/k8scsi/livenessprobe:v1.1.0
          args:
            - --csi-address=/csi/csi.sock
            - --health-port=9810
          volumeMounts:
            - mountPath: /csi
              name: plugin-dir
      volumes:
        - name: kubelet-dir
          hostPath:
            path: /var/lib/kubelet
            type: Directory
        - name: registration-dir
          hostPath:
            path: /var/lib/kubelet/plugins_registry/
            type: Directory
        - name: plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins/fsx.csi.aws.com/
            type: DirectoryOrCreate
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-provisioner-cluster-role.yaml", `
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: fsx-csi-external-provisioner-role
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["csinodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
`)
	th.writeF("/manifests/aws/aws-fsx-csi-driver/base/csi-provisioner-cluster-role-binding.yaml", `
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: fsx-csi-external-provisioner-binding
subjects:
  - kind: ServiceAccount
    name: fsx-csi-controller-sa
    namespace: kubeflow
roleRef:
  kind: ClusterRole
  name: fsx-csi-external-provisioner-role
  apiGroup: rbac.authorization.k8s.io
`)
	th.writeK("/manifests/aws/aws-fsx-csi-driver/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- csi-driver.yaml
- csi-controller.yaml
- csi-controller-sa.yaml
- csi-node-daemonset.yaml
- csi-provisioner-cluster-role.yaml
- csi-provisioner-cluster-role-binding.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
- name: amazon/aws-fsx-csi-driver
  newName: amazon/aws-fsx-csi-driver
  newTag: v0.3.0
`)
}

func TestAwsFsxCsiDriverBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/aws/aws-fsx-csi-driver/base")
	writeAwsFsxCsiDriverBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../aws/aws-fsx-csi-driver/base"
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
