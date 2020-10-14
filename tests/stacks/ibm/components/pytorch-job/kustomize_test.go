package pytorch_job

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/ibm/components/pytorch-job",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
