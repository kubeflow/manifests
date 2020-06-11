package katib_controller

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/ibm/application/katib-controller",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
