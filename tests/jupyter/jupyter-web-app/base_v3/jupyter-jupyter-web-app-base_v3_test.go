package base_v3

import (
	"github.com/ghodss/yaml"
	"github.com/kubeflow/manifests/tests"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

type expectedResource struct  {
	fileName string
	yaml string
	u *unstructured.Unstructured
}

func Key(kind string, name string ) string {
	return kind + "." + name
}
// Key returns a unique identifier fo this resource that can be used to index it
func (r* expectedResource) Key() string {
	if r.u == nil {
		return ""
	}

	return Key(r.u.GetKind(), r.u.GetName())
}

func TestJupyterJupyterWebAppBase_v3(t *testing.T) {
	expected := map[string]*expectedResource{}

	// Read all the YAML files containing expected resources and parse them.
	files, err := ioutil.ReadDir("test_data")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		contents, err := ioutil.ReadFile("test_data/" + f.Name())
		if err != nil {
			t.Fatalf("Err: %v", err)
		}

		u := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(contents), u); err != nil {
			t.Fatalf("Error: %v", err)
		}

		r := &expectedResource{
			fileName: f.Name(),
			yaml: string(contents),
			u: u,
		}

		expected[r.Key()] = r
	}

	targetPath := "../../../../jupyter/jupyter-web-app/base_v3"
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
		t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}

	actualNames := map[string]bool {}

	// Check that all the actual resources match the expected resources
	for _, r := range actual.Resources() {
		rKey := Key(r.GetKind(), r.GetName())
		actualNames[rKey] = true

		e, ok := expected[rKey]

		if !ok {
			t.Errorf("Expected resources missing resource: %v", rKey)
			continue
		}

		actualYaml, err := r.AsYAML()

		if err != nil {
			t.Errorf("Could not generate YAML for resource: %v; error: %v", rKey, err)
			continue
		}
		// Ensure the actual YAML matches.
		if string(actualYaml) != e.yaml {
			tests.ReportDiffAndFail(t, actualYaml, e.yaml)
		}
	}

	// Make sure we aren't missing any expected resources
	for name, _  := range expected {
		if _, ok := actualNames[name]; !ok {
			t.Errorf("Actual resources is missing expected resource: %v", name)
		}
	}
}
