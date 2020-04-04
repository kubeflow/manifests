package tensorboard

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../tests/legacy_kustomizations/tensorboard",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
