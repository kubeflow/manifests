package managed_cert

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package: "../../../../../gcp/basic-auth-ingress/overlays/managed-cert",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}