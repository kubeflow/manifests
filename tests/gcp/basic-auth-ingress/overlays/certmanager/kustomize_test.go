package certmanager

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package: "../../../../../gcp/basic-auth-ingress/overlays/certmanager",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}