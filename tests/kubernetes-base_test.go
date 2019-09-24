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

func writeKubeflowRolesBase(th *KustTestHarness) {
	th.writeF("/manifests/kubeflow-roles/base/cluster-roles.yaml", `
---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-admin
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-edit
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-admin: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
rules: []

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubeflow-view
  labels:
    rbac.authorization.kubeflow.org/aggregate-to-kubeflow-edit: "true"
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.authorization.kubeflow.org/aggregate-to-kubeflow-view: "true"
rules: []
`)
	th.writeK("/manifests/kubeflow-roles/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-roles.yaml
`)
}

func TestKubeflowRolesBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/kubeflow-roles/base")
	writeKubeflowRolesBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../kubeflow-roles/base"
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
