package application

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../mpi-job/mpi-operator/overlays/application",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
