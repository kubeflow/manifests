package add_anonymous_user_filter

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/kubernetes/application/add-anonymous-user-filter",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
