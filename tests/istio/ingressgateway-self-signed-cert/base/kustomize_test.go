package base

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../istio/ingressgateway-self-signed-cert/base",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
