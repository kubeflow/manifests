package istio_1_3_1_stack_openshift

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/ibm/application/istio-1-3-1-stack-openshift",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
