package use_kf_user

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../pipeline/pipelines-runner/overlays/use-kf-user",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
