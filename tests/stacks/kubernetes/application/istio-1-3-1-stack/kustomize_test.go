package istio_1_3_1_stack

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/kubernetes/application/istio-1-3-1-stack",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
