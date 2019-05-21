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

func writeMysqlBase(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/mysql/base/deployment.yaml", `
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: deployment
spec:
  strategy:
    type: Recreate
  template:
    spec:
      containers:
      - name: container
        env:
        - name: MYSQL_ALLOW_EMPTY_PASSWORD
          value: "true"
        image: mysql:5.6
        ports:
        - containerPort: 3306
          name: mysql
        volumeMounts:
        - mountPath: /var/lib/mysql
          name: persistent-storage
      volumes:
      - name: persistent-storage
        persistentVolumeClaim:
          claimName: ml-pipeline-mysql-persistent-volume-claim
`)
	th.writeF("/manifests/pipeline/mysql/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: service
spec:
  ports:
  - port: 3306
`)
	th.writeK("/manifests/pipeline/mysql/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
nameprefix: ml-pipeline-mysql-
resources:
- deployment.yaml
- service.yaml
images:
- name: mysql
  newTag: '5.6'
`)
}

func TestMysqlBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/mysql/base")
	writeMysqlBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/mysql/base"
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
