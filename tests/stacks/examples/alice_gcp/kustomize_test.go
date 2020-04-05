package alice_gcp

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../stacks/examples/alice_gcp",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
