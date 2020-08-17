package cert_manager

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/kubernetes/application/cert-manager",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
