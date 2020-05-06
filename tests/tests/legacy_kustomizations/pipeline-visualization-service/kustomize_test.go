package pipeline_visualization_service

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../tests/legacy_kustomizations/pipeline-visualization-service",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
