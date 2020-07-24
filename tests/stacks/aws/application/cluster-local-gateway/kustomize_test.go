package cluster_local_gateway

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/aws/application/cluster-local-gateway",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
