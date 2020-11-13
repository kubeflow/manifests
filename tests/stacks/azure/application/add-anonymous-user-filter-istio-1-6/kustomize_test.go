package add_anonymous_user_filter_istio_1_6

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/azure/application/add-anonymous-user-filter-istio-1-6",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
