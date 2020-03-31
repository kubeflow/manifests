package gcp_1_1_6

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../istio/gcp-1-1-6",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
