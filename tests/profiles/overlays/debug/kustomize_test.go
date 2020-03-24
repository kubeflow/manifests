package debug

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../profiles/overlays/debug",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
