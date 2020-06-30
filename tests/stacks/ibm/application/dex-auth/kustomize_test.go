package dex_auth

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/ibm/application/dex-auth",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
