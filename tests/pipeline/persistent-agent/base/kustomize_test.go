package base

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../pipeline/persistent-agent/base",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
