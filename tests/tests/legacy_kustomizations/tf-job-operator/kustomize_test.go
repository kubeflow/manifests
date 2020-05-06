package tf_job_operator

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../tests/legacy_kustomizations/tf-job-operator",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
