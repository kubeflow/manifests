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

func writeKatibV1Alpha2ManagerBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/katib-manager/base/katib-manager-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: manager
  name: katib-manager
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: manager
      name: katib-manager
    spec:
      containers:
      - command:
        - ./katib-manager
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              key: MYSQL_ROOT_PASSWORD
              name: katib-db-secrets
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        livenessProbe:
          exec:
            command:
            - /bin/grpc_health_probe
            - -addr=:6789
          initialDelaySeconds: 10
        name: katib-manager
        ports:
        - containerPort: 6789
          name: api
        readinessProbe:
          exec:
            command:
            - /bin/grpc_health_probe
            - -addr=:6789
          initialDelaySeconds: 5
`)
	th.writeF("/manifests/katib-v1alpha2/katib-manager/base/katib-manager-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: manager
  name: katib-manager
  namespace: kubeflow
spec:
  ports:
  - name: api
    port: 6789
    protocol: TCP
  selector:
    app: katib
    component: manager
  type: ClusterIP
`)
	th.writeF("/manifests/katib-v1alpha2/katib-manager/base/katib-manager-rest-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: manager-rest
  name: katib-manager-rest
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: manager-rest
      name: katib-manager-rest
    spec:
      containers:
      - command:
        - ./katib-manager-rest
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager-rest:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        name: katib-manager-rest
        ports:
        - containerPort: 80
          name: api
`)
	th.writeF("/manifests/katib-v1alpha2/katib-manager/base/katib-manager-rest-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: manager-rest
  name: katib-manager-rest
  namespace: kubeflow
spec:
  ports:
  - name: api
    port: 80
    protocol: TCP
  selector:
    app: katib
    component: manager-rest
  type: ClusterIP
`)
	th.writeK("/manifests/katib-v1alpha2/katib-manager/base", `
namespace: kubeflow
resources:
- katib-manager-deployment.yaml
- katib-manager-rest-deployment.yaml
- katib-manager-rest-service.yaml
- katib-manager-service.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager
    newTag: v0.1.2-alpha-289-g14dad8b
  - name: gcr.io/kubeflow-images-public/katib/v1alpha2/katib-manager-rest
    newTag: v0.1.2-alpha-289-g14dad8b
`)
}

func TestKatibV1Alpha2ManagerBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/katib-manager/base")
	writeKatibV1Alpha2ManagerBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/katib-manager/base"
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
