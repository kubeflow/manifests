package generic

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../stacks/generic",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
