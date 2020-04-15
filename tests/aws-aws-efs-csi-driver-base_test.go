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

func writeAwsEfsCsiDriverBase(th *KustTestHarness) {
	th.writeF("/manifests/aws/aws-efs-csi-driver/base/csi-driver.yaml", `
---
apiVersion: storage.k8s.io/v1beta1
kind: CSIDriver
metadata:
  name: efs.csi.aws.com
spec:
  attachRequired: false
`)
	th.writeF("/manifests/aws/aws-efs-csi-driver/base/csi-node-daemonset.yaml", `
---
# Node Service
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: efs-csi-node
spec:
  selector:
    matchLabels:
      app: efs-csi-node
  template:
    metadata:
      labels:
        app: efs-csi-node
    spec:
      nodeSelector:
        beta.kubernetes.io/os: linux
      hostNetwork: true
      tolerations:
        - operator: Exists
      containers:
        - name: efs-plugin
          securityContext:
            privileged: true
          image: amazon/aws-efs-csi-driver:latest
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
            - name: efs-state-dir
              mountPath: /var/run/efs
          ports:
            - containerPort: 9809
              name: healthz
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /healthz
              port: healthz
            initialDelaySeconds: 10
            timeoutSeconds: 3
            periodSeconds: 2
            failureThreshold: 5
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
              value: /var/lib/kubelet/plugins/efs.csi.aws.com/csi.sock
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
            - --health-port=9809
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
            path: /var/lib/kubelet/plugins/efs.csi.aws.com/
            type: DirectoryOrCreate
        - name: efs-state-dir
          hostPath:
            path: /var/run/efs
            type: DirectoryOrCreate

`)
	th.writeK("/manifests/aws/aws-efs-csi-driver/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: kubeflow
resources:
- csi-driver.yaml
- csi-node-daemonset.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
- name: amazon/aws-efs-csi-driver
  newName: amazon/aws-efs-csi-driver
  newTag: v0.3.0
`)
}

func TestAwsEfsCsiDriverBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/aws/aws-efs-csi-driver/base")
	writeAwsEfsCsiDriverBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../aws/aws-efs-csi-driver/base"
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
