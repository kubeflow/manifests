package default_install

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../tests/legacy_kustomizations/default-install",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
