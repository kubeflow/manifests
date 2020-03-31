package cert_manager

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../tests/legacy_kustomizations/cert-manager",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
