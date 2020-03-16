package application

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../pipeline/scheduledworkflow/overlays/application",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
