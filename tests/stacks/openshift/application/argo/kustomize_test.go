package argo

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/openshift/application/argo",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
