package application_crds

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../tests/legacy_kustomizations/application-crds",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
