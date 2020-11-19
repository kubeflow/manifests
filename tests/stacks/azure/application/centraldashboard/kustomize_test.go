package centraldashboard

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/azure/application/centraldashboard",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
