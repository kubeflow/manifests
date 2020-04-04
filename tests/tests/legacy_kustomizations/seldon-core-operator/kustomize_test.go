package seldon_core_operator

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../tests/legacy_kustomizations/seldon-core-operator",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
