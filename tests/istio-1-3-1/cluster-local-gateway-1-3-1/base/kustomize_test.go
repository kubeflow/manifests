package base

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package: "../../../../istio-1-3-1/cluster-local-gateway-1-3-1/base",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}