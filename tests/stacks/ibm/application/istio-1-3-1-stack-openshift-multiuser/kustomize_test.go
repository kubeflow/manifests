package istio_1_3_1_stack_openshift_multiuser

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/ibm/application/istio-1-3-1-stack-openshift-multiuser",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
