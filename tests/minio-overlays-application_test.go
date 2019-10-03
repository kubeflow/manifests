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

func writeMinioOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/pipeline/minio/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: $(generateName)
  componentKinds:
  - group: apps
    kind: Deployment
  - group: core
    kind: PersistentVolumeClaim
  - group: core
    kind: Secret
  - group: core
    kind: Service
  descriptor:
    type: "pipeline-minio"
    version: "v1beta1"
    description: "minio component for kubeflow pipelines"
    keywords:
    - "pipelines"
    links:
    - description: About
      url: "https://github.com/kubeflow/pipelines"
`)
	th.writeF("/manifests/pipeline/minio/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Application
- path: spec/selector/app.kubernetes.io\/instance
  kind: Service
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/template/metadata/labels/app.kubernetes.io\/instance
  kind: Deployment
`)
	th.writeF("/manifests/pipeline/minio/overlays/application/params.env", `
generateName=
`)
	th.writeK("/manifests/pipeline/minio/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: pipeline-minio-app-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: pipeline-minio-app-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: pipeline-minio-app-parameters
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: pipeline-minio
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/pipeline/minio/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
spec:
  strategy:
    type: Recreate
  template:
    spec:
      containers:
      - name: minio
        args:
        - server
        - /data
        env:
        - name: MINIO_ACCESS_KEY
          value: minio
        - name: MINIO_SECRET_KEY
          value: minio123
        image: minio/minio:RELEASE.2018-02-09T22-40-05Z
        ports:
        - containerPort: 9000
        volumeMounts:
        - mountPath: /data
          name: data
          subPath: minio
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: $(minioPvcName)
`)
	th.writeF("/manifests/pipeline/minio/base/secret.yaml", `
apiVersion: v1
data:
  accesskey: bWluaW8=
  secretkey: bWluaW8xMjM=
kind: Secret
metadata:
  name: mlpipeline-minio-artifact
type: Opaque
`)
	th.writeF("/manifests/pipeline/minio/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: minio-service
spec:
  ports:
  - port: 9000
    protocol: TCP
    targetPort: 9000
  selector:
    app: minio
`)
	th.writeF("/manifests/pipeline/minio/base/persistent-volume-claim.yaml", `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: $(minioPvcName)
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
`)
	th.writeF("/manifests/pipeline/minio/base/params.yaml", `
varReference:
- path: spec/template/spec/volumes/persistentVolumeClaim/claimName
  kind: Deployment
- path: metadata/name
  kind: PersistentVolumeClaim
`)
	th.writeF("/manifests/pipeline/minio/base/params.env", `
minioPvcName=
`)
	th.writeK("/manifests/pipeline/minio/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:
  app: minio
resources:
- deployment.yaml
- secret.yaml
- service.yaml
- persistent-volume-claim.yaml
configMapGenerator:
- name: pipeline-minio-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: minioPvcName
  objref:
    kind: ConfigMap
    name: pipeline-minio-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.minioPvcName
images:
- name: minio/minio
  newTag: RELEASE.2018-02-09T22-40-05Z
configurations:
- params.yaml
`)
}

func TestMinioOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pipeline/minio/overlays/application")
	writeMinioOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pipeline/minio/overlays/application"
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
