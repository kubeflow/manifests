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

func writeKatibV1Alpha2DbBase(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/katib-db/base/katib-db-pvc.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: katib-mysql
  namespace: kubeflow
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`)
	th.writeF("/manifests/katib-v1alpha2/katib-db/base/katib-db-deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: katib
    component: db
  name: katib-db
  namespace: kubeflow
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: katib
        component: db
      name: katib-db
    spec:
      containers:
      - args:
        - --datadir
        - /var/lib/mysql/datadir
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              key: MYSQL_ROOT_PASSWORD
              name: katib-db-secrets
        - name: MYSQL_ALLOW_EMPTY_PASSWORD
          value: "true"
        - name: MYSQL_DATABASE
          value: katib
        image: mysql:8.0.3
        name: katib-db
        ports:
        - containerPort: 3306
          name: dbapi
        readinessProbe:
          exec:
            command:
            - /bin/bash
            - -c
            - mysql -D $$MYSQL_DATABASE -p$$MYSQL_ROOT_PASSWORD -e 'SELECT 1'
          initialDelaySeconds: 5
          periodSeconds: 2
          timeoutSeconds: 1
        volumeMounts:
        - mountPath: /var/lib/mysql
          name: katib-mysql
      volumes:
      - name: katib-mysql
        persistentVolumeClaim:
          claimName: katib-mysql
`)
	th.writeF("/manifests/katib-v1alpha2/katib-db/base/katib-db-secret.yaml", `
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: katib-db-secrets
data:
  MYSQL_ROOT_PASSWORD: dGVzdA== # "test"
`)
	th.writeF("/manifests/katib-v1alpha2/katib-db/base/katib-db-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: katib
    component: db
  name: katib-db
  namespace: kubeflow
spec:
  ports:
  - name: dbapi
    port: 3306
    protocol: TCP
  selector:
    app: katib
    component: db
  type: ClusterIP
`)
	th.writeK("/manifests/katib-v1alpha2/katib-db/base", `
namespace: kubeflow
resources:
- katib-db-deployment.yaml
- katib-db-pvc.yaml
- katib-db-secret.yaml
- katib-db-service.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
  - name: mysql
    newTag: 8.0.3
`)
}

func TestKatibV1Alpha2DbBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/katib-db/base")
	writeKatibV1Alpha2DbBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/katib-db/base"
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
