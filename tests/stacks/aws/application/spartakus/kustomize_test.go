package spartakus

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/aws/application/spartakus",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
