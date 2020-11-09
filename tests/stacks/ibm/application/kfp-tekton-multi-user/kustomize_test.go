package kfp_tekton_multi_user

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/ibm/application/kfp-tekton-multi-user",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
