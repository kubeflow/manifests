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

func writeCentraldashboardOverlaysStacks(th *KustTestHarness) {
	th.writeF("/manifests/common/centraldashboard/overlays/stacks/deployment_kf_config.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: centraldashboard
spec:
  template:
    spec:
        containers:
          - name: centraldashboard
            env:
            - name: USERID_HEADER
              valueFrom:
                configMapKeyRef:
                  name: kubeflow-config
                  key: userid-header
            - name: USERID_PREFIX
              valueFrom:
                configMapKeyRef:
                  name: kubeflow-config
                  key: userid-prefix
            - name: PROFILES_KFAM_SERVICE_HOST
              valueFrom:
                configMapKeyRef:
                  name: kubeflow-config
                  key: profiles_kfam_service_host
`)
	th.writeK("/manifests/common/centraldashboard/overlays/stacks", `
apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  app.kubernetes.io/component: centraldashboard
  app.kubernetes.io/instance: centraldashboard-v1.0.0
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/name: centraldashboard
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v1.0.0
kind: Kustomization
namespace: kubeflow
resources:
- ../../core
# TODO(jlewi): istio and application are really patches
# not "overlays" in that they are expected to be used as mixins.
# Perhaps move this into mixins to make this more obvious.
- ../../overlays/istio
- ../../overlays/application
patchesStrategicMerge:
# Pull in the patch which will configure central dashboard using a kubeflow
# configmap
- deployment_kf_config.yaml

`)
	th.writeF("/manifests/common/centraldashboard/base/deployment_patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: centraldashboard
spec:
  template:
    spec:
      containers:
      - name: centraldashboard
        env:
        - name: USERID_HEADER
          value: $(userid-header)
        - name: USERID_PREFIX
          value: $(userid-prefix)
        - name: PROFILES_KFAM_SERVICE_HOST
          value: profiles-kfam.kubeflow
`)
	th.writeF("/manifests/common/centraldashboard/base/params.yaml", `
varReference:
- path: metadata/annotations/getambassador.io\/config
  kind: Service
- path: spec/http/route/destination/host
  kind: VirtualService
- path: spec/template/spec/containers/0/env/0/value
  kind: Deployment
- path: spec/template/spec/containers/0/env/1/value
  kind: Deployment`)
	th.writeF("/manifests/common/centraldashboard/base/params.env", `
clusterDomain=cluster.local
userid-header=
userid-prefix=`)
	th.writeK("/manifests/common/centraldashboard/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../base_v3
patchesStrategicMerge:
- deployment_patch.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: centraldashboard
images:
- name: gcr.io/kubeflow-images-public/centraldashboard
  newName: gcr.io/kubeflow-images-public/centraldashboard
  newTag: vmaster-gf39279c0
configMapGenerator:
- envs:
  - params.env
  name: parameters
generatorOptions:
  disableNameSuffixHash: true
vars:
- fieldref:
    fieldPath: metadata.namespace
  name: namespace
  objref:
    apiVersion: v1
    kind: Service
    name: centraldashboard
- fieldref:
    fieldPath: data.clusterDomain
  name: clusterDomain
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: parameters
- fieldref:
    fieldPath: data.userid-header
  name: userid-header
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: parameters
- fieldref:
    fieldPath: data.userid-prefix
  name: userid-prefix
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: parameters
configurations:
- params.yaml
`)
}

func TestCentraldashboardOverlaysStacks(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/common/centraldashboard/overlays/stacks")
	writeCentraldashboardOverlaysStacks(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../common/centraldashboard/overlays/stacks"
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
