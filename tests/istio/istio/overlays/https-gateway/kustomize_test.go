package https_gateway

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package: "../../../../../istio/istio/overlays/https-gateway",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}