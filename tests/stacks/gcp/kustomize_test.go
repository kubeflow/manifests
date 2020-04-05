package gcp

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../stacks/gcp",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
