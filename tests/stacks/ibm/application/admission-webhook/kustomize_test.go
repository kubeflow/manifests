package admission_webhook

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/ibm/application/admission-webhook",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
