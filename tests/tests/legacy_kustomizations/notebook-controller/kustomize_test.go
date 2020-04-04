package notebook_controller

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../tests/legacy_kustomizations/notebook-controller",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
