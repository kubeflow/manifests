package cluster_local_gateway_1_3_1

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/aws/application/cluster-local-gateway-1-3-1",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
