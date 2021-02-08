package openshift_without_registration_flow

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../stacks/openshift_without_registration_flow",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
