package istio_stack

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/aws/application/istio-stack",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
