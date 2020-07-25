package oidc_authservice

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/aws/application/oidc-authservice",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
