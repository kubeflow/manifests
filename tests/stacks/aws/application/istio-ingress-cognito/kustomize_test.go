package istio_ingress_cognito

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/aws/application/istio-ingress-cognito",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
